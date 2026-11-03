package main

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strconv"
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
	ChannelID string `json:"channel_id"`
	UserID string `json:"user_id"`
	RootID string `json:"root_id,omitempty"`
	Message string `json:"message"`
	CreateAt int64 `json:"create_at"`
}

type MessagePage struct {
	ChannelID string `json:"channel_id"`
	AnchorPostID string `json:"anchor_post_id"`
	Posts []ContextPost `json:"posts"`
	HasMore bool `json:"has_more"`
	NextCursor string `json:"next_cursor,omitempty"`
}

type AgentTokenAuthenticator interface { Authenticate(raw string) (string,error) }
type AgentBridge interface {
	Get(owner, postID string) (ContextPost,error)
	List(owner, postID, direction, scope string, limit int) (MessagePage,error)
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

// Expose only message text and stable IDs, never props, attachments, or deleted posts.
func sanitizePosts(list *model.PostList) []ContextPost {
	if list==nil {return []ContextPost{}}
	posts:=make([]ContextPost,0,len(list.Order))
	for _,id:=range list.Order {
		post:=list.Posts[id]
		if post==nil || post.DeleteAt!=0 {continue}
		posts=append(posts,postToAgentPost(post))
	}
	sort.SliceStable(posts,func(i,j int)bool {if posts[i].CreateAt==posts[j].CreateAt{return posts[i].ID<posts[j].ID};return posts[i].CreateAt<posts[j].CreateAt})
	return posts
}

func postToAgentPost(post *model.Post) ContextPost {
	return ContextPost{ID:post.Id,ChannelID:post.ChannelId,UserID:post.UserId,RootID:post.RootId,Message:post.Message,CreateAt:post.CreateAt}
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
	case "/api/agent/posts/get", "/api/agent/posts/list":
		if r.Method!=http.MethodGet {writeAgentError(w,http.StatusMethodNotAllowed,"GET required");return}
		id,err:=parsePostReference(r.URL.Query().Get("reference"),h.SiteURL)
		if err!=nil {writeAgentError(w,http.StatusBadRequest,"invalid Mattermost permalink or post ID");return}
		if r.URL.Path=="/api/agent/posts/get" {
			post,err:=h.Bridge.Get(owner,id)
			if err!=nil {agentBridgeError(w,err);return}
			_ = json.NewEncoder(w).Encode(post)
			return
		}
		q:=r.URL.Query()
		direction:=q.Get("direction")
		if direction!="before" && direction!="after" {writeAgentError(w,http.StatusBadRequest,"direction must be before or after");return}
		scope:=q.Get("scope")
		if scope=="" {scope="channel"}
		if scope!="channel" && scope!="thread" {writeAgentError(w,http.StatusBadRequest,"scope must be channel or thread");return}
		limit:=20
		if q.Has("limit") {
			limit,err=strconv.Atoi(q.Get("limit"))
			if err!=nil || limit<1 || limit>50 {writeAgentError(w,http.StatusBadRequest,"limit must be 1..50");return}
		}
		page,err:=h.Bridge.List(owner,id,direction,scope,limit)
		if err!=nil {agentBridgeError(w,err);return}
		_ = json.NewEncoder(w).Encode(page)
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
