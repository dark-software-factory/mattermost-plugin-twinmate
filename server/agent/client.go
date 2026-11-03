package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const maxReplyBytes = 64 * 1024

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
}

func ValidateEndpoint(raw string, allowPrivate bool) error {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme != "https" || u.Hostname() == "" || u.User != nil || u.Fragment != "" || u.Opaque != "" {
		return errors.New("endpoint must be an absolute HTTPS URL without credentials or fragments")
	}
	host := strings.ToLower(u.Hostname())
	if !allowPrivate && (host == "localhost" || strings.HasSuffix(host, ".localhost") || strings.HasSuffix(host, ".local")) {
		return errors.New("private agent endpoints are disabled")
	}
	if ip := net.ParseIP(host); ip != nil && !allowPrivate && !allowedIP(ip) {
		return errors.New("private agent endpoints are disabled")
	}
	return nil
}

func allowedIP(ip net.IP) bool {
	return ip.IsGlobalUnicast() && !ip.IsPrivate() && !ip.IsLoopback() && !ip.IsLinkLocalUnicast() && !ip.IsLinkLocalMulticast()
}

func NewClient(allowPrivate bool) *Client {
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
			if !allowPrivate && !allowedIP(candidate.IP) {
				continue
			}
			return (&net.Dialer{Timeout: 5 * time.Second}).DialContext(ctx, network, net.JoinHostPort(candidate.IP.String(), port))
		}
		return nil, fmt.Errorf("no permitted address for agent host %q", host)
	}
	return &Client{HTTP: &http.Client{Timeout: 25 * time.Second, Transport: transport}, AllowPrivate: allowPrivate}
}

func (c *Client) Call(ctx context.Context, endpoint, bearer string, event Event) (string, error) {
	if err := ValidateEndpoint(endpoint, c.AllowPrivate); err != nil {
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
		client = NewClient(c.AllowPrivate).HTTP
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
