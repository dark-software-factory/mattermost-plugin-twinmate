package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/mattermost/mattermost/server/public/model"
)

const examplePostID = "abcdefghijklmnopqrstuvwx12"

type fakeAgentBridge struct {
	contextOwner string
	contextPost string
	sendOwner string
	sendChannel string
	sendReply string
	sendText string
	contextError error
	sendError error
}

func (b *fakeAgentBridge) Context(owner, postID string) (ContextResult,error) {
	b.contextOwner,b.contextPost=owner,postID
	if b.contextError!=nil {return ContextResult{},b.contextError}
	return ContextResult{ChannelID:"channel-1",TargetPostID:postID,Posts:[]ContextPost{{ID:postID,UserID:"petya",Message:"broken deploy",CreateAt:1234}}},nil
}
func (b *fakeAgentBridge) Send(owner, channel, reply, message string) (string,error) {
	b.sendOwner,b.sendChannel,b.sendReply,b.sendText=owner,channel,reply,message
	if b.sendError!=nil {return "",b.sendError}
	return "generated-post-id",nil
}

func newAuthenticatedAgentHandler(t *testing.T, b *fakeAgentBridge) (AgentHandler,string) {
	t.Helper()
	tokens:=&TokenManager{Store:&memoryTokenKV{data:map[string][]byte{}}}
	token,err:=tokens.Rotate("misha")
	if err!=nil {t.Fatal(err)}
	return AgentHandler{Tokens:tokens,Bridge:b,SiteURL:"https://mm.example.com"},token
}

func TestAgentContextGetsLinkedThreadAsTokenOwner(t *testing.T) {
	b:=&fakeAgentBridge{}
	h,token:=newAuthenticatedAgentHandler(t,b)
	request:=httptest.NewRequest(http.MethodGet,"/api/agent/context?url="+url.QueryEscape("https://mm.example.com/team/pl/"+examplePostID),nil)
	request.Header.Set("X-Twinmate-Token",token)
	response:=httptest.NewRecorder()
	h.ServeHTTP(response,request)
	if response.Code!=http.StatusOK || b.contextOwner!="misha" || b.contextPost!=examplePostID {
		t.Fatalf("linked context failed: %d %s, owner=%s post=%s",response.Code,response.Body.String(),b.contextOwner,b.contextPost)
	}
	if !bytes.Contains(response.Body.Bytes(),[]byte("broken deploy")) {t.Fatal("missing context")}
}

func TestAgentAPIRejectsMissingCredentialsAndWrongOrigin(t *testing.T) {
	b:=&fakeAgentBridge{}
	h,token:=newAuthenticatedAgentHandler(t,b)
	unauth:=httptest.NewRecorder()
	h.ServeHTTP(unauth,httptest.NewRequest(http.MethodGet,"/api/agent/context?post_id="+examplePostID,nil))
	if unauth.Code!=http.StatusUnauthorized {t.Fatalf("missing credential accepted: %d",unauth.Code)}
	bad:=httptest.NewRequest(http.MethodGet,"/api/agent/context?url="+url.QueryEscape("https://attacker.test/team/pl/"+examplePostID),nil)
	bad.Header.Set("X-Twinmate-Token",token)
	badResponse:=httptest.NewRecorder()
	h.ServeHTTP(badResponse,bad)
	if badResponse.Code!=http.StatusBadRequest || b.contextOwner!="" {t.Fatalf("foreign permalink accepted: %d",badResponse.Code)}
}

func TestAgentAPISendPostsAsAuthenticatedOwner(t *testing.T) {
	b:=&fakeAgentBridge{}
	h,token:=newAuthenticatedAgentHandler(t,b)
	request:=httptest.NewRequest(http.MethodPost,"/api/agent/messages",bytes.NewBufferString(`{"channel_id":"channel-1","reply_to_post_id":"root-1","message":"investigation done"}`))
	request.Header.Set("X-Twinmate-Token",token)
	request.Header.Set("Content-Type","application/json")
	response:=httptest.NewRecorder()
	h.ServeHTTP(response,request)
	if response.Code!=http.StatusCreated || b.sendOwner!="misha" || b.sendChannel!="channel-1" || b.sendReply!="root-1" || b.sendText!="investigation done" {
		t.Fatalf("send failed: %d, %+v",response.Code,b)
	}
	var result struct { PostID string `json:"post_id"` }
	if err:=json.Unmarshal(response.Body.Bytes(),&result);err!=nil || result.PostID!="generated-post-id" {t.Fatalf("invalid receipt: %s %v",response.Body.String(),err)}
	forgery:=httptest.NewRequest(http.MethodPost,"/api/agent/messages",bytes.NewBufferString(`{"channel_id":"channel-1","message":"fake","owner_id":"petya"}`))
	forgery.Header.Set("X-Twinmate-Token",token)
	forgery.Header.Set("Content-Type","application/json")
	forgeryResult:=httptest.NewRecorder()
	h.ServeHTTP(forgeryResult,forgery)
	if forgeryResult.Code!=http.StatusBadRequest {t.Fatalf("owner spoofing accepted: %d",forgeryResult.Code)}
}

func TestAgentAPIMapsReadPermissionsWithoutLeakingContent(t *testing.T) {
	b:=&fakeAgentBridge{contextError:ErrAgentForbidden}
	h,token:=newAuthenticatedAgentHandler(t,b)
	request:=httptest.NewRequest(http.MethodGet,"/api/agent/context?post_id="+examplePostID,nil)
	request.Header.Set("X-Twinmate-Token",token)
	response:=httptest.NewRecorder()
	h.ServeHTTP(response,request)
	if response.Code!=http.StatusForbidden || bytes.Contains(response.Body.Bytes(),[]byte("broken deploy")) {
		t.Fatalf("unauthorized content leaked: %d %s",response.Code,response.Body.String())
	}
}

func TestPostReferenceParserRestrictsMattermostSite(t *testing.T) {
	for _,raw:=range []string{examplePostID,"https://mm.example.com/team/pl/"+examplePostID,"https://mm.example.com/team/channels/town?post_id="+examplePostID} {
		id,err:=parsePostReference(raw,"https://mm.example.com")
		if err!=nil || id!=examplePostID {t.Fatalf("%s => %s %v",raw,id,err)}
	}
	for _,raw:=range []string{"https://evil.test/team/pl/"+examplePostID,"https://mm.example.com.evil.test/team/pl/"+examplePostID,"nonsense"} {
		if _,err:=parsePostReference(raw,"https://mm.example.com");err==nil {t.Fatalf("accepted %s",raw)}
	}
}

func TestThreadContextFiltersDeletedAndCapsMessages(t *testing.T) {
	posts:=&model.PostList{Order:[]string{},Posts:map[string]*model.Post{}}
	for i:=0;i<60;i++ {
		id:=model.NewId()
		posts.Order=append(posts.Order,id)
		posts.Posts[id]=&model.Post{Id:id,UserId:"person",ChannelId:"dm",Message:"hi",CreateAt:int64(i)}
	}
	posts.Posts[posts.Order[1]].DeleteAt=1
	result:=sanitizeThread(posts,posts.Order[0],50)
	if len(result)!=50 {t.Fatalf("expected bounded 50 messages, got %d",len(result))}
	if result[0].ID!=posts.Order[0] {t.Fatalf("thread root was lost")}
	for _,post:=range result {if post.ID==posts.Order[1] {t.Fatal("deleted post leaked")}}
	if !sortIsChronological(result) {t.Fatal("thread posts not chronological")}
}

func sortIsChronological(posts []ContextPost) bool {for i:=1;i<len(posts);i++ {if posts[i].CreateAt<posts[i-1].CreateAt {return false}};return true}

func TestAgentAPIBridgeErrorsHaveStableSentinels(t *testing.T) {
	if errors.Is(ErrAgentForbidden,ErrAgentNotFound) {t.Fatal("access-denied must not be confused with absence")}
}

type fakeMattermostBridgeAPI struct {
	post *model.Post
	channel *model.Channel
	allowRead bool
	allowWrite bool
	member bool
	thread *model.PostList
}

func (f *fakeMattermostBridgeAPI) GetPost(string) (*model.Post,*model.AppError) {return f.post,nil}
func (f *fakeMattermostBridgeAPI) GetPostThread(string) (*model.PostList,*model.AppError) {return f.thread,nil}
func (f *fakeMattermostBridgeAPI) GetChannel(string) (*model.Channel,*model.AppError) {return f.channel,nil}
func (f *fakeMattermostBridgeAPI) GetChannelMember(string,string) (*model.ChannelMember,*model.AppError) {
	if !f.member {return nil,&model.AppError{Message:"not a member"}}
	return &model.ChannelMember{UserId:"misha"},nil
}
func (f *fakeMattermostBridgeAPI) HasPermissionToChannel(_,_ string,perm *model.Permission) bool {
	if perm==model.PermissionReadChannel {return f.allowRead}
	if perm==model.PermissionCreatePost {return f.allowWrite}
	return false
}

func TestMattermostBridgeChecksPermissionsBeforeReturningThread(t *testing.T) {
	post:=&model.Post{Id:examplePostID,ChannelId:"secret-dm",UserId:"petya",Message:"private"}
	f:=&fakeMattermostBridgeAPI{post:post,channel:&model.Channel{Id:"secret-dm",Type:model.ChannelTypeDirect},allowRead:false,member:true,thread:&model.PostList{Order:[]string{examplePostID},Posts:map[string]*model.Post{examplePostID:post}}}
	bridge:=MattermostBridge{API:f}
	if _,err:=bridge.Context("misha",examplePostID);!errors.Is(err,ErrAgentForbidden) {t.Fatalf("private context returned: %v",err)}
	f.allowRead=true
	f.member=false
	if _,err:=bridge.Context("misha",examplePostID);!errors.Is(err,ErrAgentForbidden) {t.Fatalf("nonmember context returned: %v",err)}
	f.member=true
	result,err:=bridge.Context("misha",examplePostID)
	if err!=nil || len(result.Posts)!=1 {t.Fatalf("authorized context failed: %+v %v",result,err)}
}

func TestMattermostBridgePreventsCrossChannelReplyAndUnauthorizedWrite(t *testing.T) {
	f:=&fakeMattermostBridgeAPI{post:&model.Post{Id:examplePostID,ChannelId:"channel-A"},channel:&model.Channel{Id:"channel-A",Type:model.ChannelTypeOpen},allowWrite:true,member:true}
	called:=false
	bridge:=MattermostBridge{API:f,Publisher:func(owner,channel,root,message string)(string,error) {called=true;return "posted-id",nil}}
	if _,err:=bridge.Send("misha","channel-B",examplePostID,"hello");!errors.Is(err,ErrAgentBadInput) {t.Fatalf("cross-channel reply accepted: %v",err)}
	if called {t.Fatal("published cross-channel reply")}
	f.allowWrite=false
	if _,err:=bridge.Send("misha","channel-A",examplePostID,"hello");!errors.Is(err,ErrAgentForbidden) {t.Fatalf("write without permission: %v",err)}
	f.allowWrite=true
	id,err:=bridge.Send("misha","channel-A",examplePostID,"hello")
	if err!=nil || id!="posted-id" || !called {t.Fatalf("authorized write failed: %s %v",id,err)}
}
