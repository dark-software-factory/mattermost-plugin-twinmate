package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dark-software-factory/mattermost-plugin-twinmate/server/core"
)

type memorySettings struct{ values map[string]core.Connection }

func (s *memorySettings) Get(owner string) (core.Connection, error) { return s.values[owner], nil }
func (s *memorySettings) Put(owner string, config core.Connection) error { s.values[owner] = config; return nil }

func TestSettingsRequireMattermostAuthentication(t *testing.T) {
	store := &memorySettings{values: make(map[string]core.Connection)}
	request := httptest.NewRequest(http.MethodPut, "/api/config", bytes.NewBufferString(`{"endpoint":"https://agent.example.com","enabled":true}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	(&SettingsHandler{Store: store}).ServeHTTP(response, request)
	if response.Code != http.StatusUnauthorized || len(store.values) != 0 {
		t.Fatalf("unauthenticated update allowed: %d", response.Code)
	}
}

func TestSettingsNeverReturnPersonalAgentToken(t *testing.T) {
	store := &memorySettings{values: map[string]core.Connection{"misha": {Endpoint: "https://agent.example.com", Token: "top-secret", Enabled: true}}}
	request := httptest.NewRequest(http.MethodGet, "/api/config", nil)
	request.Header.Set("Mattermost-User-Id", "misha")
	response := httptest.NewRecorder()
	(&SettingsHandler{Store: store}).ServeHTTP(response, request)
	if response.Code != http.StatusOK || bytes.Contains(response.Body.Bytes(), []byte("top-secret")) {
		t.Fatalf("secret leaked: %s", response.Body.String())
	}
	var body struct { Endpoint string `json:"endpoint"`; HasToken bool `json:"has_token"` }
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil || body.Endpoint != "https://agent.example.com" || !body.HasToken {
		t.Fatalf("unexpected response: %+v, %v", body, err)
	}
}

func TestSettingsPreserveTokenWhenUpdatingEndpoint(t *testing.T) {
	store := &memorySettings{values: map[string]core.Connection{"misha": {Endpoint: "https://old.example.com", Token: "existing"}}}
	request := httptest.NewRequest(http.MethodPut, "/api/config", bytes.NewBufferString(`{"endpoint":"https://new.example.com/events","enabled":true}`))
	request.Header.Set("Mattermost-User-Id", "misha")
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	(&SettingsHandler{Store: store}).ServeHTTP(response, request)
	if response.Code != http.StatusOK || store.values["misha"].Token != "existing" || !store.values["misha"].Enabled {
		t.Fatalf("unexpected save: %d, %+v", response.Code, store.values["misha"])
	}
}

func TestSettingsRejectUnsafeEndpoint(t *testing.T) {
	store := &memorySettings{values: make(map[string]core.Connection)}
	request := httptest.NewRequest(http.MethodPut, "/api/config", bytes.NewBufferString(`{"endpoint":"https://127.0.0.1/metadata","enabled":true}`))
	request.Header.Set("Mattermost-User-Id", "misha")
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	(&SettingsHandler{Store: store}).ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest || len(store.values) != 0 {
		t.Fatalf("unsafe endpoint accepted: %d", response.Code)
	}
}
