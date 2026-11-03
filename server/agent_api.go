package main

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"

	"github.com/mattermost/mattermost/server/public/model"
)

var (
	ErrAgentForbidden = errors.New("agent owner lacks channel permission")
	ErrAgentNotFound = errors.New("Mattermost message or channel not found")
	ErrAgentBadInput = errors.New("invalid Twinmate agent request")
)

type ContextPost struct {
	ID string `json:"id"`
	UserID string `json:"user_id"`
	RootID string `json:"root_id,omitempty"`
	Message string `json:"message"`
	CreateAt int64 `json:"create_at"`
}

type ContextResult struct {
	ChannelID string `json:"channel_id"`
	TargetPostID string `json:"target_post_id"`
	Posts []ContextPost `json:"posts"`
}

type AgentTokenAuthenticator interface { Authenticate(raw string) (string,error) }
type AgentBridge interface {
	Context(owner, postID string) (ContextResult,error)
	Send(owner, channel, reply, message string) (string,error)
}

type AgentHandler struct {
	Tokens AgentTokenAuthenticator
	Bridge AgentBridge
	SiteURL string
}

func parsePostReference(input, siteURL string) (string,error) {
	if model.IsValidId(input) { return input,nil }
	if siteURL == "" {return "",ErrAgentBadInput}
	u, err := url.Parse(input)
	if err!=nil || u.Host=="" || u.User!=nil || u.Fragment!="" {return "",ErrAgentBadInput}
	site, err := url.Parse(siteURL)
	if err!=nil || site.Host=="" || !strings.EqualFold(u.Host,site.Host) || !strings.EqualFold(u.Scheme,site.Scheme) {return "",ErrAgentBadInput}
	id:=u.Query().Get("post_id")
	if id=="" {
		parts:=strings.Split(strings.Trim(u.Path,"/"),"/")
		for i:=1;i<len(parts);i++ {if parts[i-1]=="pl" {id=parts[i];break}}
	}
	if !model.IsValidId(id) {return "",ErrAgentBadInput}
	return id,nil
}

// Limit a possibly large Mattermost thread and never export deleted posts or
// post Props. Preserve the root plus the recent conversation when truncating.
func sanitizeThread(list *model.PostList, target string, limit int) []ContextPost {
	if list==nil || limit<=0 {return []ContextPost{}}
	posts:=make([]ContextPost,0,len(list.Order))
	for _,id:=range list.Order {
		post:=list.Posts[id]
		if post==nil || post.DeleteAt!=0 {continue}
		posts=append(posts,ContextPost{ID:post.Id,UserID:post.UserId,RootID:post.RootId,Message:post.Message,CreateAt:post.CreateAt})
	}
	sort.SliceStable(posts,func(i,j int)bool {if posts[i].CreateAt==posts[j].CreateAt{return posts[i].ID<posts[j].ID};return posts[i].CreateAt<posts[j].CreateAt})
	if len(posts)<=limit {return posts}
	// Include the earliest root for orientation; keep the latest messages and
	// ensure the linked target survives truncation.
	selected:=append([]ContextPost{posts[0]},posts[len(posts)-(limit-1):]...)
	hasTarget:=false
	for _,p:=range selected {if p.ID==target {hasTarget=true;break}}
	if !hasTarget {
		for _,p:=range posts {if p.ID==target {selected[1]=p;break}}
	}
	sort.SliceStable(selected,func(i,j int)bool {if selected[i].CreateAt==selected[j].CreateAt{return selected[i].ID<selected[j].ID};return selected[i].CreateAt<selected[j].CreateAt})
	return selected
}

func writeAgentError(w http.ResponseWriter, status int, message string) {
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error":message})
}

func agentBridgeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err,ErrAgentForbidden): writeAgentError(w,http.StatusForbidden,"not allowed to access this channel")
	case errors.Is(err,ErrAgentNotFound): writeAgentError(w,http.StatusNotFound,"message or channel not found")
	case errors.Is(err,ErrAgentBadInput): writeAgentError(w,http.StatusBadRequest,"invalid message or channel")
	default: writeAgentError(w,http.StatusInternalServerError,"Mattermost operation failed")
	}
}

func (h AgentHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control","no-store")
	w.Header().Set("Content-Type","application/json")
	owner,err:=h.Tokens.Authenticate(r.Header.Get("X-Twinmate-Token"))
	if errors.Is(err,ErrInvalidAgentToken) {writeAgentError(w,http.StatusUnauthorized,"invalid Twinmate token");return}
	if err!=nil {writeAgentError(w,http.StatusInternalServerError,"authentication failed");return}
	switch r.URL.Path {
	case "/api/agent/context":
		if r.Method!=http.MethodGet {writeAgentError(w,http.StatusMethodNotAllowed,"GET required");return}
		input:=r.URL.Query().Get("post_id")
		if r.URL.Query().Has("url") {
			if input!="" {writeAgentError(w,http.StatusBadRequest,"choose post_id or url");return}
			input=r.URL.Query().Get("url")
		}
		id,err:=parsePostReference(input,h.SiteURL)
		if err!=nil {writeAgentError(w,http.StatusBadRequest,"invalid Mattermost permalink or post ID");return}
		result,err:=h.Bridge.Context(owner,id)
		if err!=nil {agentBridgeError(w,err);return}
		_ = json.NewEncoder(w).Encode(result)
	case "/api/agent/messages":
		if r.Method!=http.MethodPost {writeAgentError(w,http.StatusMethodNotAllowed,"POST required");return}
		if !strings.HasPrefix(r.Header.Get("Content-Type"),"application/json") {writeAgentError(w,http.StatusUnsupportedMediaType,"application/json required");return}
		var input struct {
			ChannelID string `json:"channel_id"`
			ReplyToPostID string `json:"reply_to_post_id"`
			Message string `json:"message"`
		}
		dec:=json.NewDecoder(http.MaxBytesReader(w,r.Body,20*1024))
		dec.DisallowUnknownFields()
		if err:=dec.Decode(&input);err!=nil {writeAgentError(w,http.StatusBadRequest,"invalid JSON");return}
		var trailing any
		if err:=dec.Decode(&trailing);err!=io.EOF {writeAgentError(w,http.StatusBadRequest,"multiple JSON objects");return}
		if strings.TrimSpace(input.Message)=="" || len([]byte(input.Message))>15000 || (input.ChannelID=="" && input.ReplyToPostID=="") {
			writeAgentError(w,http.StatusBadRequest,"message and channel or reply target required");return
		}
		id,err:=h.Bridge.Send(owner,input.ChannelID,input.ReplyToPostID,input.Message)
		if err!=nil {agentBridgeError(w,err);return}
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(struct {PostID string `json:"post_id"`}{id})
	default: http.NotFound(w,r)
	}
}
