package main

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"github.com/dark-software-factory/mattermost-plugin-twinmate/server/agent"
	"github.com/dark-software-factory/mattermost-plugin-twinmate/server/core"
)

type SettingsStore interface {
	Get(ownerID string) (core.Connection, error)
	Put(ownerID string, connection core.Connection) error
}

type SettingsHandler struct {
	Store        SettingsStore
	AllowPrivate bool
	ApprovedHosts []string
}

// Tokens are scoped to the original destination, not to arbitrary URLs a
// settings update might redirect them to. A path change on the same origin
// does not require re-entering the token.
func sameEndpointOrigin(oldEndpoint, newEndpoint string) bool {
	oldURL, oldErr := url.Parse(oldEndpoint)
	newURL, newErr := url.Parse(newEndpoint)
	if oldErr != nil || newErr != nil || oldURL.Host == "" || newURL.Host == "" {
		return false
	}
	return strings.EqualFold(oldURL.Scheme, newURL.Scheme) && strings.EqualFold(oldURL.Hostname(), newURL.Hostname()) && oldURL.Port() == newURL.Port()
}

func (h *SettingsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Type", "application/json")
	owner := r.Header.Get("Mattermost-User-Id")
	if owner == "" {
		http.Error(w, `{"error":"authentication required"}`, http.StatusUnauthorized)
		return
	}
	current, err := h.Store.Get(owner)
	if err != nil {
		http.Error(w, `{"error":"cannot load connection"}`, http.StatusInternalServerError)
		return
	}
	if r.Method == http.MethodGet {
		_ = json.NewEncoder(w).Encode(struct {
			Endpoint string `json:"endpoint"`
			Enabled bool `json:"enabled"`
			HasToken bool `json:"has_token"`
		}{current.Endpoint, current.Enabled, current.Token != ""})
		return
	}
	if r.Method != http.MethodPut {
		w.Header().Set("Allow", "GET, PUT")
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}
	if !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
		http.Error(w, `{"error":"expected application/json"}`, http.StatusUnsupportedMediaType)
		return
	}
	var incoming struct {
		Endpoint string `json:"endpoint"`
		Token string `json:"token"`
		Enabled bool `json:"enabled"`
		RemoveToken bool `json:"remove_token"`
	}
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&incoming); err != nil {
		http.Error(w, `{"error":"invalid JSON"}`, http.StatusBadRequest)
		return
	}
	if incoming.Endpoint != "" {
		if err := agent.ValidateEndpoint(incoming.Endpoint, h.AllowPrivate, h.ApprovedHosts...); err != nil {
			http.Error(w, `{"error":"unsafe agent URL"}`, http.StatusBadRequest)
			return
		}
	} else if incoming.Enabled {
		http.Error(w, `{"error":"endpoint required"}`, http.StatusBadRequest)
		return
	}
	if incoming.RemoveToken && incoming.Token != "" {
		http.Error(w, `{"error":"cannot replace and remove token together"}`, http.StatusBadRequest)
		return
	}
	if incoming.RemoveToken || (incoming.Token == "" && !sameEndpointOrigin(current.Endpoint, incoming.Endpoint)) {
		current.Token = ""
	}
	current.Endpoint = incoming.Endpoint
	current.Enabled = incoming.Enabled
	if incoming.Token != "" {
		current.Token = incoming.Token
	}
	if err := h.Store.Put(owner, current); err != nil {
		http.Error(w, `{"error":"cannot save connection"}`, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"ok":true}`))
}
