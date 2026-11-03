package agent

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func sampleEvent() Event {
	return Event{EventID: "event-1", OwnerID: "misha", SenderID: "petya", ChannelID: "dm-1", MessageID: "post-1", Text: "When?"}
}

func TestClientSendsEventAndReturnsReply(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.Header.Get("Authorization") != "Bearer agent-secret" || r.Header.Get("X-Twinmate-Event-ID") != "event-1" {
			t.Errorf("unexpected request method or headers: %s, %v", r.Method, r.Header)
		}
		var event Event
		if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
			t.Fatal(err)
		}
		if event.Text != "When?" || event.OwnerID != "misha" || event.SenderID != "petya" {
			t.Errorf("unexpected event: %+v", event)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"reply":"Next sprint."}`))
	}))
	defer server.Close()

	client := Client{HTTP: server.Client(), AllowPrivate: true}
	reply, err := client.Call(context.Background(), server.URL, "agent-secret", sampleEvent())
	if err != nil || reply != "Next sprint." {
		t.Fatalf("reply=%q, err=%v", reply, err)
	}
}

func TestClientAcceptsAgentDecline(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"reply":"  "}`))
	}))
	defer server.Close()
	reply, err := (&Client{HTTP: server.Client(), AllowPrivate: true}).Call(context.Background(), server.URL, "", sampleEvent())
	if err != nil || reply != "" {
		t.Fatalf("reply=%q, err=%v", reply, err)
	}
}

func TestClientRejectsUnsafeEndpoints(t *testing.T) {
	for _, endpoint := range []string{
		"http://example.com/agent", "https://localhost/agent", "https://127.0.0.1/agent",
		"https://169.254.169.254/latest/meta-data", "https://u:p@example.com/agent", "https://example.com/#fragment",
	} {
		t.Run(endpoint, func(t *testing.T) {
			if err := ValidateEndpoint(endpoint, false); err == nil {
				t.Fatal("expected endpoint rejection")
			}
		})
	}
	if err := ValidateEndpoint("https://agents.company.example/events", false); err != nil {
		t.Fatalf("valid HTTPS endpoint rejected: %v", err)
	}
}

func TestClientNeverFollowsRedirect(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "https://example.com/steal", http.StatusTemporaryRedirect)
	}))
	defer server.Close()
	_, err := (&Client{HTTP: server.Client(), AllowPrivate: true}).Call(context.Background(), server.URL, "secret", sampleEvent())
	if err == nil || !strings.Contains(err.Error(), "redirect") {
		t.Fatalf("expected redirect rejection, got %v", err)
	}
}

func TestClientLimitsResponseBody(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"reply":"` + strings.Repeat("x", 70*1024) + `"}`))
	}))
	defer server.Close()
	_, err := (&Client{HTTP: server.Client(), AllowPrivate: true}).Call(context.Background(), server.URL, "", sampleEvent())
	if err == nil || !strings.Contains(err.Error(), "large") {
		t.Fatalf("expected response size rejection, got %v", err)
	}
}

func TestClientRespectsCancellation(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		_, _ = w.Write([]byte(`{"reply":"too late"}`))
	}))
	defer server.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	_, err := (&Client{HTTP: server.Client(), AllowPrivate: true}).Call(ctx, server.URL, "", sampleEvent())
	if err == nil {
		t.Fatal("expected request cancellation")
	}
}
