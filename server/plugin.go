package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/dark-software-factory/mattermost-plugin-twinmate/server/agent"
	"github.com/dark-software-factory/mattermost-plugin-twinmate/server/core"
	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin"
)

const twinmateLabel = "🤖 **Twinmate · AI-generated**\n\n"

const twinmateDMNotice = "🔔 **Twinmate notice:** This DM now allows a participant's personal AI agent to receive new messages outside Mattermost and send visibly labeled AI-generated replies. Ask the participant to use `/twinmate off` to stop forwarding."

type dmOptInRecord struct {
	NoticePostID string `json:"notice_post_id"`
}

type storedPostProvenance struct {
	OwnerID string `json:"owner_id"`
	ChannelID string `json:"channel_id"`
}

func postProvenanceKey(id string) string { return "provenance:" + id }

func postMatchesProvenance(post *model.Post, origin storedPostProvenance) bool {
	return post != nil && post.UserId == origin.OwnerID && post.ChannelId == origin.ChannelID &&
		strings.HasPrefix(post.Message, twinmateLabel)
}

type Plugin struct {
	plugin.MattermostPlugin
	allowPrivate atomic.Bool
	privateHosts atomic.Value // []string, changed atomically on config reload
	publishing sync.Map // server-created posts in flight, keyed by author/channel/body
	workSlots chan struct{}
}

func pendingTwinKey(post *model.Post) string {
	return post.UserId+"\x00"+post.ChannelId+"\x00"+post.RootId+"\x00"+post.Message
}

// Client-writable props and an arbitrary emoji prefix cannot prove provenance.
// A server-created post is suppressed while publishing and thereafter by KV.
func (p *Plugin) shouldProcessPost(post *model.Post, author *model.User) bool {
	if post==nil || post.UserId=="" || post.ChannelId=="" || strings.TrimSpace(post.Message)=="" || post.Type!="" || (author!=nil && (author.IsBot || author.DeleteAt!=0)) {return false}
	if post.GetProp("from_webhook")=="true" {return false}
	if _,pending:=p.publishing.Load(pendingTwinKey(post));pending {return false}
	if strings.HasPrefix(post.Message,twinmateLabel) {
		data,err:=p.API.KVGet(postProvenanceKey(post.Id))
		if err!=nil {return false} // fail closed on attribution store outage
		var origin storedPostProvenance
		if json.Unmarshal(data,&origin)==nil && postMatchesProvenance(post,origin) {return false}
	}
	return true
}

// MessageWillBeUpdated protects the native-mobile body label against edits.
// The record is server-owned; neither the client prop nor text by itself is proof.
func (p *Plugin) MessageWillBeUpdated(_ *plugin.Context, newPost, oldPost *model.Post) (*model.Post,string) {
	if oldPost==nil || newPost==nil || oldPost.Id=="" {return newPost,""}
	// The visible body label is also the fallback when provenance KVSet
	// failed after the post was published. This veto does not certify
	// provenance for badges or automatic forwarding.
	if strings.HasPrefix(oldPost.Message,twinmateLabel) && !strings.HasPrefix(newPost.Message,twinmateLabel) {
		return nil,"Twinmate AI attribution cannot be removed"
	}
	data,err:=p.API.KVGet(postProvenanceKey(oldPost.Id))
	if err!=nil {return nil,"Twinmate cannot verify AI attribution right now"}
	if len(data)==0 {return newPost,""}
	var origin storedPostProvenance
	if json.Unmarshal(data,&origin)!=nil || origin.OwnerID!=oldPost.UserId || origin.ChannelID!=oldPost.ChannelId {return newPost,""}
	if !strings.HasPrefix(newPost.Message,twinmateLabel) {return nil,"Twinmate AI attribution cannot be removed"}
	return newPost,""
}

func (p *Plugin) OnConfigurationChange() error {
	var config struct {
		AllowPrivateEndpoints bool `json:"AllowPrivateEndpoints"`
		AllowedPrivateHosts string `json:"AllowedPrivateHosts"`
	}
	if err := p.API.LoadPluginConfiguration(&config); err != nil {
		return err
	}
	approved := make([]string,0)
	for _,entry := range strings.FieldsFunc(config.AllowedPrivateHosts,func(c rune)bool {return c==',' || c=='\n'}) {
		entry = strings.TrimSpace(entry)
		host,port,err:=net.SplitHostPort(entry)
		if err!=nil || host=="" || port=="" || strings.ContainsAny(host,"*/@") || !validPrivatePort(port) {
			return fmt.Errorf("invalid AllowedPrivateHosts entry %q: expected exact host:port",entry)
		}
		approved=append(approved,net.JoinHostPort(strings.ToLower(host),port))
	}
	p.privateHosts.Store(approved)
	p.allowPrivate.Store(config.AllowPrivateEndpoints)
	return nil
}

func validPrivatePort(port string) bool {
	n,err:=strconv.Atoi(port)
	return err==nil && n>0 && n<=65535
}

func (p *Plugin) approvedPrivateHosts() []string {
	if v:=p.privateHosts.Load();v!=nil{return v.([]string)}
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
	if !p.shouldProcessPost(post, nil) {
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
		if author == nil || author.IsBot || author.DeleteAt != 0 {
			return
		}
		processor := core.Processor{Platform: p, Agent: agent.NewClient(p.allowPrivate.Load(), p.approvedPrivateHosts()...)}
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
	if channel.DeleteAt!=0 {return "",nil}
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
	user,appErr:=p.API.GetUser(other)
	if appErr!=nil || user==nil || user.DeleteAt!=0 || user.IsBot || !p.API.HasPermissionToChannel(other,channelID,model.PermissionReadChannel) || !p.API.HasPermissionToChannel(other,channelID,model.PermissionCreatePost) {
		return "",nil
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
	var record dmOptInRecord
	// Old plain "on" states have no verifiable disclosure. Fail closed
	// and let the owner opt in again to publish a new notice.
	if json.Unmarshal(data,&record)!=nil || record.NoticePostID=="" {return false,nil}
	notice,appErr:=p.API.GetPost(record.NoticePostID)
	if appErr!=nil || notice==nil || notice.DeleteAt!=0 || notice.UserId!=userID || notice.ChannelId!=channelID || notice.Type!="" || notice.Message!=twinmateDMNotice {
		return false,nil
	}
	return true,nil
}

func (p *Plugin) setDM(userID, channelID, noticePostID string) error {
	state := []byte("off")
	if noticePostID!="" {
		var err error
		state,err=json.Marshal(dmOptInRecord{NoticePostID:noticePostID})
		if err!=nil {return err}
	}
	if appErr := p.API.KVSet(dmKey(userID, channelID), state); appErr != nil {
		return appErr
	}
	return nil
}

func (p *Plugin) Publish(_ context.Context, ownerID, channelID, rootID, text, sourceID string) error {
	if !((MattermostBridge{API:p.API}).canAccess(ownerID,channelID,model.PermissionCreatePost)) {return ErrAgentForbidden}
	_, err := p.createTwinPost(ownerID, channelID, rootID, text, sourceID)
	return err
}

func (p *Plugin) createTwinPost(ownerID, channelID, rootID, text, sourceID string) (string,error) {
	// Leave room for the label and stay under Mattermost's post size limit.
	if len([]byte(text)) > 15000 {
		return "",errors.New("agent reply exceeds the 15000-byte Twinmate limit")
	}
	post := &model.Post{
		UserId: ownerID, ChannelId: channelID, RootId: rootID,
		Message: twinmateLabel + text,
		Props: model.StringInterface{
			"twinmate": map[string]any{
				"generated": true, "owner_id": ownerID, "source_post_id": sourceID,
			},
		},
	}
	key:=pendingTwinKey(post)
	p.publishing.Store(key,struct{}{})
	defer p.publishing.Delete(key)
	created, appErr := p.API.CreatePost(post)
	if appErr != nil {
		return "",fmt.Errorf("publishing Twinmate response: %w", appErr)
	}
	if created == nil || created.Id == "" {
		return "",errors.New("Mattermost returned no post ID after publishing")
	}
	origin, err := json.Marshal(storedPostProvenance{OwnerID: ownerID, ChannelID: channelID})
	if err != nil {
		return "",err
	}
	if appErr := p.API.KVSet(postProvenanceKey(created.Id), origin); appErr != nil {
		// CreatePost has already succeeded. Reporting a failed send invites MCP
		// callers to retry and duplicate the message. The immutable body label
		// still discloses AI authorship even when the web badge is unavailable.
		p.API.LogError("Twinmate post published without verified badge provenance", "post_id", created.Id, "error", appErr.Error())
		return created.Id,nil
	}
	return created.Id,nil
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
	channel,appErr:=p.API.GetChannel(origin.ChannelID)
	if appErr!=nil || channel==nil || !p.API.HasPermissionToChannel(viewerID,origin.ChannelID,model.PermissionReadChannel) {
		return false,nil
	}
	if channel.Type!=model.ChannelTypeOpen {
		if _, appErr := p.API.GetChannelMember(origin.ChannelID, viewerID); appErr != nil {return false,nil}
	}
	if channel.DeleteAt!=0 {
		return false, nil
	}
	post, appErr := p.API.GetPost(postID)
	if appErr != nil {
		return false, nil
	}
	return postMatchesProvenance(post, origin), nil
}

func (p *Plugin) ServeHTTP(_ *plugin.Context, w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/api/agent/posts/get" || r.URL.Path == "/api/agent/posts/list" || r.URL.Path == "/api/agent/messages" {
		siteURL := ""
		if config := p.API.GetConfig(); config != nil && config.ServiceSettings.SiteURL != nil {
			siteURL = *config.ServiceSettings.SiteURL
		}
		(AgentHandler{
			Tokens: &TokenManager{Store: pluginAccessTokenStore{Plugin: p}},
			Bridge: MattermostBridge{API: p.API,Publisher: func(owner,channel,root,message string)(string,error) {
				return p.createTwinPost(owner,channel,root,message,"")
			}},
			SiteURL: siteURL,
		}).ServeHTTP(w,r)
		return
	}
	if r.URL.Path == "/api/agent/token" {
		(TokenSettingsHandler{Manager: &TokenManager{Store: pluginAccessTokenStore{Plugin: p}}}).ServeHTTP(w, r)
		return
	}
	if r.URL.Path == "/api/config" {
		(&SettingsHandler{Store: p, AllowPrivate: p.allowPrivate.Load(), ApprovedHosts:p.approvedPrivateHosts()}).ServeHTTP(w, r)
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
