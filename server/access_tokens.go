package main

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"io"
	"strings"

	"github.com/mattermost/mattermost/server/public/model"
)

var ErrInvalidAgentToken = errors.New("invalid or revoked Twinmate agent token")

type AccessTokenStore interface {
	Load(key string) ([]byte, error)
	Save(key string, data []byte) error
	Remove(key string) error
}

type TokenManager struct { Store AccessTokenStore }

type storedAgentToken struct {
	OwnerID string `json:"owner_id"`
	Digest []byte `json:"digest"`
	ScopeVersion int `json:"scope_version,omitempty"`
	AllowSend bool `json:"allow_send,omitempty"`
	ChannelIDs []string `json:"channel_ids,omitempty"`
}

type AgentClaims struct {
	OwnerID string
	AllowSend bool
	ChannelIDs []string
}

func (c AgentClaims) AllowsChannel(channelID string) bool {
	if len(c.ChannelIDs)==0{return true}
	for _,id:=range c.ChannelIDs {if id==channelID{return true}}
	return false
}

func ownerTokenKey(owner string) string { return "agent-token-owner:" + owner }
func agentTokenKey(id string) string { return "agent-token-id:" + id }

func (m *TokenManager) Rotate(owner string) (string, error) {
	return m.RotateScoped(owner,false,nil)
}

func (m *TokenManager) RotateScoped(owner string, allowSend bool, channelIDs []string) (string, error) {
	if owner == "" { return "", errors.New("owner required") }
	if len(channelIDs)>32 {return "",errors.New("at most 32 channel IDs allowed")}
	unique:=make(map[string]struct{},len(channelIDs))
	for _,id:=range channelIDs {
		if !model.IsValidId(id) {return "",errors.New("invalid channel ID")}
		if _,exists:=unique[id];exists {return "",errors.New("duplicate channel ID")}
		unique[id]=struct{}{}
	}
	idBytes := make([]byte, 12)
	secret := make([]byte, 32)
	if _, err := rand.Read(idBytes); err != nil { return "", err }
	if _, err := rand.Read(secret); err != nil { return "", err }
	id := hex.EncodeToString(idBytes)
	digest := sha256.Sum256(secret)
	record, err := json.Marshal(storedAgentToken{OwnerID: owner, Digest: digest[:], ScopeVersion:1, AllowSend:allowSend, ChannelIDs:channelIDs})
	if err != nil { return "", err }
	if err := m.Store.Save(agentTokenKey(id), record); err != nil { return "", err }
	oldID, err := m.Store.Load(ownerTokenKey(owner))
	if err != nil { return "", err }
	if err := m.Store.Save(ownerTokenKey(owner), []byte(id)); err != nil { return "", err }
	if len(oldID) != 0 { _ = m.Store.Remove(agentTokenKey(string(oldID))) }
	return "tmt_" + id + "_" + hex.EncodeToString(secret), nil
}

func (m *TokenManager) Authenticate(raw string) (string, error) {
	claims,err:=m.AuthenticateClaims(raw)
	return claims.OwnerID,err
}

func (m *TokenManager) AuthenticateClaims(raw string) (AgentClaims, error) {
	parts := strings.Split(raw, "_")
	if len(parts) != 3 || parts[0] != "tmt" || len(parts[1]) != 24 || len(parts[2]) != 64 {
		return AgentClaims{}, ErrInvalidAgentToken
	}
	secret, err := hex.DecodeString(parts[2])
	if err != nil { return AgentClaims{}, ErrInvalidAgentToken }
	if _, err := hex.DecodeString(parts[1]); err != nil { return AgentClaims{}, ErrInvalidAgentToken }
	data, err := m.Store.Load(agentTokenKey(parts[1]))
	if err != nil { return AgentClaims{}, err }
	if len(data) == 0 { return AgentClaims{}, ErrInvalidAgentToken }
	var record storedAgentToken
	if err := json.Unmarshal(data, &record); err != nil || record.OwnerID == "" || len(record.Digest) != sha256.Size {
		return AgentClaims{}, ErrInvalidAgentToken
	}
	active, err := m.Store.Load(ownerTokenKey(record.OwnerID))
	if err != nil { return AgentClaims{}, err }
	digest := sha256.Sum256(secret)
	if subtle.ConstantTimeCompare(record.Digest, digest[:]) != 1 || subtle.ConstantTimeCompare(active, []byte(parts[1])) != 1 {
		return AgentClaims{}, ErrInvalidAgentToken
	}
	if record.ScopeVersion!=0 && record.ScopeVersion!=1 {return AgentClaims{},ErrInvalidAgentToken}
	if len(record.ChannelIDs)>32 {return AgentClaims{},ErrInvalidAgentToken}
	return AgentClaims{OwnerID:record.OwnerID,AllowSend:record.ScopeVersion==0 || record.AllowSend,ChannelIDs:record.ChannelIDs},nil
}

func (m *TokenManager) Revoke(owner string) error {
	oldID, err := m.Store.Load(ownerTokenKey(owner))
	if err != nil { return err }
	if err := m.Store.Remove(ownerTokenKey(owner)); err != nil { return err }
	if len(oldID) != 0 { _ = m.Store.Remove(agentTokenKey(string(oldID))) }
	return nil
}

func (m *TokenManager) HasToken(owner string) (bool, error) {
	id, err := m.Store.Load(ownerTokenKey(owner))
	if err != nil { return false, err }
	return len(id) != 0, nil
}

func (m *TokenManager) Status(owner string) (bool,AgentClaims,error) {
	id,err:=m.Store.Load(ownerTokenKey(owner))
	if err!=nil{return false,AgentClaims{},err}
	if len(id)==0{return false,AgentClaims{},nil}
	data,err:=m.Store.Load(agentTokenKey(string(id)))
	if err!=nil{return false,AgentClaims{},err}
	var record storedAgentToken
	if json.Unmarshal(data,&record)!=nil || record.OwnerID!=owner{return false,AgentClaims{},ErrInvalidAgentToken}
	return true,AgentClaims{OwnerID:owner,AllowSend:record.ScopeVersion==0||record.AllowSend,ChannelIDs:record.ChannelIDs},nil
}

type pluginAccessTokenStore struct { Plugin *Plugin }

func (s pluginAccessTokenStore) Load(key string) ([]byte, error) {
	value, err := s.Plugin.API.KVGet(key)
	if err != nil { return nil, err }
	return value, nil
}
func (s pluginAccessTokenStore) Save(key string, value []byte) error {
	if err := s.Plugin.API.KVSet(key, value); err != nil { return err }
	return nil
}
func (s pluginAccessTokenStore) Remove(key string) error {
	if err := s.Plugin.API.KVDelete(key); err != nil { return err }
	return nil
}

type TokenSettingsHandler struct { Manager *TokenManager }

func (h TokenSettingsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Type", "application/json")
	owner := r.Header.Get("Mattermost-User-Id")
	if owner == "" {
		http.Error(w, `{"error":"Mattermost session required"}`, http.StatusUnauthorized)
		return
	}
	if r.Method != http.MethodGet && r.Method != http.MethodPost && r.Method != http.MethodDelete {
		w.Header().Set("Allow", "GET, POST, DELETE")
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}
	if r.Method != http.MethodGet && !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
		http.Error(w, `{"error":"application/json required"}`, http.StatusUnsupportedMediaType)
		return
	}
	switch r.Method {
	case http.MethodGet:
		present, claims, err := h.Manager.Status(owner)
		if err != nil { http.Error(w, `{"error":"cannot load token"}`, http.StatusInternalServerError); return }
		_ = json.NewEncoder(w).Encode(struct { HasToken bool `json:"has_token"`; AllowSend bool `json:"allow_send"`; ChannelIDs []string `json:"channel_ids"` }{present,claims.AllowSend,claims.ChannelIDs})
	case http.MethodPost:
		var options struct {AllowSend bool `json:"allow_send"`; ChannelIDs []string `json:"channel_ids"`}
		dec:=json.NewDecoder(http.MaxBytesReader(w,r.Body,4096))
		dec.DisallowUnknownFields()
		if err:=dec.Decode(&options);err!=nil {http.Error(w,`{"error":"invalid token scope"}`,http.StatusBadRequest);return}
		var trailing any
		if err:=dec.Decode(&trailing);err!=io.EOF {http.Error(w,`{"error":"invalid token scope"}`,http.StatusBadRequest);return}
		token, err := h.Manager.RotateScoped(owner,options.AllowSend,options.ChannelIDs)
		if err!=nil {http.Error(w, `{"error":"invalid token scope"}`, http.StatusBadRequest);return}
		if err != nil { http.Error(w, `{"error":"cannot issue token"}`, http.StatusInternalServerError); return }
		_ = json.NewEncoder(w).Encode(struct { Token string `json:"token"` }{token})
	case http.MethodDelete:
		if err := h.Manager.Revoke(owner); err != nil { http.Error(w, `{"error":"cannot revoke token"}`, http.StatusInternalServerError); return }
		_, _ = fmt.Fprint(w, `{"ok":true}`)
	}
}
