package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/netip"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const maxReplyBytes = 64 * 1024

// IsGlobalUnicast alone is not a public-internet test: shared/benchmarking
// networks and documentation ranges can still route inside organizations.
// Be conservative about special-use networks when admin opt-in is disabled.
var specialUseNetworks = []netip.Prefix{
	netip.MustParsePrefix("0.0.0.0/8"),
	netip.MustParsePrefix("100.64.0.0/10"),
	netip.MustParsePrefix("192.0.0.0/24"),
	netip.MustParsePrefix("192.0.2.0/24"),
	netip.MustParsePrefix("192.88.99.0/24"),
	netip.MustParsePrefix("198.18.0.0/15"),
	netip.MustParsePrefix("198.51.100.0/24"),
	netip.MustParsePrefix("203.0.113.0/24"),
	netip.MustParsePrefix("240.0.0.0/4"),
	netip.MustParsePrefix("64:ff9b::/96"),
	netip.MustParsePrefix("64:ff9b:1::/48"),
	netip.MustParsePrefix("100::/64"),
	netip.MustParsePrefix("2001::/23"),
	netip.MustParsePrefix("2001:db8::/32"),
	netip.MustParsePrefix("2002::/16"),
}

// Event is the deliberately small contract shared with independently hosted agents.
// No Mattermost credentials or conversation history are included.
type Event struct {
	EventID   string `json:"event_id"`
	OwnerID   string `json:"owner_id"`
	SenderID  string `json:"sender_id"`
	ChannelID string `json:"channel_id"`
	MessageID string `json:"message_id"`
	Text      string `json:"text"`
}

// Client calls a user-managed agent. A nil HTTP client uses a transport that
// resolves hosts at connection time and rejects private destinations by default.
type Client struct {
	HTTP         *http.Client
	AllowPrivate bool
	ApprovedHosts []string
}

// The administrator's global opt-in does not authorize arbitrary internal hosts.
func approvedHost(host, port string, allowPrivate bool, approved []string) bool {
	if !allowPrivate { return false }
	if port == "" { port = "443" }
	for _, candidate := range approved {
		if strings.EqualFold(candidate, net.JoinHostPort(host, port)) { return true }
	}
	return false
}

func ValidateEndpoint(raw string, allowPrivate bool, approved ...string) error {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme != "https" || u.Hostname() == "" || u.User != nil || u.Fragment != "" || u.Opaque != "" {
		return errors.New("endpoint must be an absolute HTTPS URL without credentials or fragments")
	}
	host := strings.ToLower(u.Hostname())
	allowedPrivateHost := approvedHost(host, u.Port(), allowPrivate, approved)
	if !allowedPrivateHost && (host == "localhost" || strings.HasSuffix(host, ".localhost") || strings.HasSuffix(host, ".local")) {
		return errors.New("private agent endpoints are disabled")
	}
	if ip := net.ParseIP(host); ip != nil && !allowedPrivateHost && !allowedIP(ip) {
		return errors.New("private agent endpoints are disabled")
	}
	return nil
}

func allowedIP(ip net.IP) bool {
	if !ip.IsGlobalUnicast() || ip.IsPrivate() || ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
		return false
	}
	address, ok := netip.AddrFromSlice(ip)
	if !ok {
		return false
	}
	address = address.Unmap()
	for _, block := range specialUseNetworks {
		if block.Contains(address) {
			return false
		}
	}
	return true
}

func NewClient(allowPrivate bool, approved ...string) *Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	// Avoid using a general-purpose proxy to bypass the address restrictions.
	transport.Proxy = nil
	transport.DialContext = func(ctx context.Context, network, address string) (net.Conn, error) {
		host, port, err := net.SplitHostPort(address)
		if err != nil {
			return nil, err
		}
		ips, err := net.DefaultResolver.LookupIPAddr(ctx, host)
		if err != nil {
			return nil, err
		}
		for _, candidate := range ips {
			if !allowedIP(candidate.IP) && !approvedHost(host, port, allowPrivate, approved) {
				continue
			}
			return (&net.Dialer{Timeout: 5 * time.Second}).DialContext(ctx, network, net.JoinHostPort(candidate.IP.String(), port))
		}
		return nil, fmt.Errorf("no permitted address for agent host %q", host)
	}
	return &Client{HTTP: &http.Client{Timeout: 25 * time.Second, Transport: transport}, AllowPrivate: allowPrivate, ApprovedHosts: append([]string(nil),approved...)}
}

func (c *Client) Call(ctx context.Context, endpoint, bearer string, event Event) (string, error) {
	if err := ValidateEndpoint(endpoint, c.AllowPrivate, c.ApprovedHosts...); err != nil {
		return "", err
	}
	data, err := json.Marshal(event)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Twinmate-Event-ID", event.EventID)
	if bearer != "" {
		req.Header.Set("Authorization", "Bearer "+bearer)
	}
	client := c.HTTP
	if client == nil {
		client = NewClient(c.AllowPrivate, c.ApprovedHosts...).HTTP
	}
	// Even a caller-supplied test/client transport must never forward secrets
	// to the Location of an HTTP redirect.
	noRedirect := *client
	noRedirect.CheckRedirect = func(*http.Request, []*http.Request) error { return errors.New("agent redirect rejected") }
	resp, err := noRedirect.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("agent returned HTTP %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxReplyBytes+1))
	if err != nil {
		return "", err
	}
	if len(body) > maxReplyBytes {
		return "", errors.New("agent response too large")
	}
	var answer struct {
		Reply string `json:"reply"`
	}
	if err := json.Unmarshal(body, &answer); err != nil {
		return "", fmt.Errorf("invalid agent JSON: %w", err)
	}
	return strings.TrimSpace(answer.Reply), nil
}
