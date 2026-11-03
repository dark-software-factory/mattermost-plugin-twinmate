package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"github.com/dark-software-factory/mattermost-plugin-twinmate/server/agent"
	"github.com/dark-software-factory/mattermost-plugin-twinmate/server/core"
	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin"
)

const twinmateLabel = "🤖 **Twinmate · AI-двойник**\n\n"

type storedPostProvenance struct {
	OwnerID string `json:"owner_id"`
	ChannelID string `json:"channel_id"`
}

func postProvenanceKey(id string) string { return "provenance:" + id }

func postMatchesProvenance(post *model.Post, origin storedPostProvenance) bool {
	return post != nil && post.UserId == origin.OwnerID && post.ChannelId == origin.ChannelID &&
		strings.HasPrefix(post.Message, twinmateLabel)
}

// Post props are writable by clients, so they are never authorization or
// provenance evidence. The stored text prefix is a loop-safety fallback.
func shouldProcessPost(post *model.Post, author *model.User) bool {
	if post == nil || post.UserId == "" || post.ChannelId == "" || strings.TrimSpace(post.Message) == "" || post.Type != "" {
		return false
	}
	if strings.HasPrefix(post.Message, twinmateLabel) {
		return false
	}
	if post.GetProp("from_webhook") == "true" {
		return false
	}
	return author == nil || !author.IsBot
}

type Plugin struct {
	plugin.MattermostPlugin
	allowPrivate atomic.Bool
	workSlots chan struct{}
}

func (p *Plugin) OnConfigurationChange() error {
	var config struct {
		AllowPrivateEndpoints bool `json:"AllowPrivateEndpoints"`
	}
	if err := p.API.LoadPluginConfiguration(&config); err != nil {
		return err
	}
	p.allowPrivate.Store(config.AllowPrivateEndpoints)
	return nil
}

func (p *Plugin) OnActivate() error {
	p.workSlots = make(chan struct{}, 32)
	return p.API.RegisterCommand(&model.Command{
		Trigger: "twinmate", AutoComplete: true,
		AutoCompleteDesc: "Enable or disable your personal AI twin in this direct message",
		AutoCompleteHint: "on | off | status",
	})
}

func (p *Plugin) MessageHasBeenPosted(_ *plugin.Context, post *model.Post) {
	if !shouldProcessPost(post, nil) {
		return
	}
	// Copy hook-owned data before launching a goroutine. Never retain the
	// Mattermost SDK's Post pointer beyond this callback.
	message := core.Message{
		ID: post.Id, ChannelID: post.ChannelId, SenderID: post.UserId,
		RootID: post.RootId, Text: post.Message,
	}
	// Mattermost's post hook must return quickly; never block chat delivery
	// while waiting for a third-party model.
	select {
	case p.workSlots <- struct{}{}:
	default:
		p.API.LogWarn("Twinmate worker limit reached", "post_id", post.Id)
		return
	}
	go func() {
		defer func() { <-p.workSlots }()
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		author, appErr := p.API.GetUser(message.SenderID)
		if appErr != nil {
			p.API.LogWarn("Twinmate cannot verify sender", "user_id", message.SenderID, "error", appErr.Error())
			return
		}
		if author == nil || author.IsBot {
			return
		}
		processor := core.Processor{Platform: p, Agent: agent.NewClient(p.allowPrivate.Load())}
		if err := processor.Handle(ctx, message); err != nil {
			p.API.LogError("Twinmate event failed", "post_id", message.ID, "error", err.Error())
		}
	}()
}

// Recipient resolves a genuine two-member DM. Group channels and unexpected
// memberships do not get sent to an external agent.
func (p *Plugin) Recipient(_ context.Context, channelID, senderID string) (string, error) {
	channel, appErr := p.API.GetChannel(channelID)
	if appErr != nil {
		return "", appErr
	}
	if channel.Type != model.ChannelTypeDirect {
		return "", nil
	}
	members, appErr := p.API.GetChannelMembers(channelID, 0, 3)
	if appErr != nil {
		return "", appErr
	}
	if len(members) != 2 {
		return "", nil
	}
	seenSender, other := false, ""
	for _, member := range members {
		if member.UserId == senderID {
			seenSender = true
		} else {
			other = member.UserId
		}
	}
	if !seenSender {
		return "", nil
	}
	return other, nil
}

func connectionKey(userID string) string { return "connection:" + userID }
func dmKey(userID, channelID string) string { return "dm:" + userID + ":" + channelID }

func (p *Plugin) Get(userID string) (core.Connection, error) {
	var config core.Connection
	data, appErr := p.API.KVGet(connectionKey(userID))
	if appErr != nil {
		return config, appErr
	}
	if len(data) == 0 {
		return config, nil
	}
	return config, json.Unmarshal(data, &config)
}

func (p *Plugin) Put(userID string, config core.Connection) error {
	data, err := json.Marshal(config)
	if err != nil {
		return err
	}
	if appErr := p.API.KVSet(connectionKey(userID), data); appErr != nil {
		return appErr
	}
	return nil
}

func (p *Plugin) Connection(_ context.Context, userID string) (core.Connection, error) {
	return p.Get(userID)
}

func (p *Plugin) DMEnabled(_ context.Context, userID, channelID string) (bool, error) {
	data, appErr := p.API.KVGet(dmKey(userID, channelID))
	if appErr != nil {
		return false, appErr
	}
	return string(data) == "on", nil
}

func (p *Plugin) setDM(userID, channelID string, enabled bool) error {
	state := []byte("off")
	if enabled {
		state = []byte("on")
	}
	if appErr := p.API.KVSet(dmKey(userID, channelID), state); appErr != nil {
		return appErr
	}
	return nil
}

func (p *Plugin) Publish(_ context.Context, ownerID, channelID, rootID, text, sourceID string) error {
	// Leave room for the label and stay under Mattermost's post size limit.
	if len([]byte(text)) > 15000 {
		return errors.New("agent reply exceeds the 15000-byte Twinmate limit")
	}
	created, appErr := p.API.CreatePost(&model.Post{
		UserId: ownerID, ChannelId: channelID, RootId: rootID,
		Message: twinmateLabel + text,
		Props: model.StringInterface{
			"twinmate": map[string]any{
				"generated": true, "owner_id": ownerID, "source_post_id": sourceID,
			},
		},
	})
	if appErr != nil {
		return fmt.Errorf("publishing Twinmate response: %w", appErr)
	}
	if created == nil || created.Id == "" {
		return errors.New("Mattermost returned no post ID after publishing")
	}
	origin, err := json.Marshal(storedPostProvenance{OwnerID: ownerID, ChannelID: channelID})
	if err != nil {
		return err
	}
	if appErr := p.API.KVSet(postProvenanceKey(created.Id), origin); appErr != nil {
		return fmt.Errorf("recording Twinmate post provenance: %w", appErr)
	}
	return nil
}

func (p *Plugin) verifiedPostForViewer(postID, viewerID string) (bool, error) {
	if len(postID) == 0 || len(postID) > 64 || viewerID == "" {
		return false, nil
	}
	data, appErr := p.API.KVGet(postProvenanceKey(postID))
	if appErr != nil {
		return false, appErr
	}
	if len(data) == 0 {
		return false, nil
	}
	var origin storedPostProvenance
	if err := json.Unmarshal(data, &origin); err != nil {
		return false, err
	}
	if origin.OwnerID == "" || origin.ChannelID == "" {
		return false, nil
	}
	if _, appErr := p.API.GetChannelMember(origin.ChannelID, viewerID); appErr != nil {
		return false, nil
	}
	post, appErr := p.API.GetPost(postID)
	if appErr != nil {
		return false, nil
	}
	return postMatchesProvenance(post, origin), nil
}

func (p *Plugin) ServeHTTP(_ *plugin.Context, w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/api/agent/token" {
		(TokenSettingsHandler{Manager: &TokenManager{Store: pluginAccessTokenStore{Plugin: p}}}).ServeHTTP(w, r)
		return
	}
	if r.URL.Path == "/api/config" {
		(&SettingsHandler{Store: p, AllowPrivate: p.allowPrivate.Load()}).ServeHTTP(w, r)
		return
	}
	if r.URL.Path != "/api/provenance" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}
	viewer := r.Header.Get("Mattermost-User-Id")
	if viewer == "" {
		http.Error(w, `{"error":"authentication required"}`, http.StatusUnauthorized)
		return
	}
	verified, err := p.verifiedPostForViewer(r.URL.Query().Get("post_id"), viewer)
	if err != nil {
		http.Error(w, `{"error":"cannot check provenance"}`, http.StatusInternalServerError)
		return
	}
	_ = json.NewEncoder(w).Encode(struct { Verified bool `json:"verified"` }{verified})
}
