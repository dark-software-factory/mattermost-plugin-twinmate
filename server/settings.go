package main

import (
	"encoding/json"
	"net/http"
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
	}
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&incoming); err != nil {
		http.Error(w, `{"error":"invalid JSON"}`, http.StatusBadRequest)
		return
	}
	if incoming.Endpoint != "" {
		if err := agent.ValidateEndpoint(incoming.Endpoint, h.AllowPrivate); err != nil {
			http.Error(w, `{"error":"unsafe agent URL"}`, http.StatusBadRequest)
			return
		}
	} else if incoming.Enabled {
		http.Error(w, `{"error":"endpoint required"}`, http.StatusBadRequest)
		return
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
