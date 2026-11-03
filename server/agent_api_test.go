package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/mattermost/mattermost/server/public/model"
)

const examplePostID = "abcdefghijklmnopqrstuvwx12"

type fakeAgentBridge struct {
	owner, postID, direction, scope string
	limit int
	sendOwner, sendChannel, sendReply, sendText string
	err error
}
func (b *fakeAgentBridge) Get(owner, postID string) (ContextPost,error) {
	b.owner,b.postID=owner,postID
	if b.err!=nil {return ContextPost{},b.err}
	return ContextPost{ID:postID,UserID:"petya",Message:"broken deploy",CreateAt:1234},nil
}
func (b *fakeAgentBridge) List(owner, postID, direction, scope string, limit int) (MessagePage,error) {
	b.owner,b.postID,b.direction,b.scope,b.limit=owner,postID,direction,scope,limit
	if b.err!=nil {return MessagePage{},b.err}
	return MessagePage{ChannelID:"channel-1",AnchorPostID:postID,Posts:[]ContextPost{{ID:"nearby",Message:"previous discussion"}}},nil
}
func (b *fakeAgentBridge) Send(owner, channel, reply, text string) (string,error) {
	b.sendOwner,b.sendChannel,b.sendReply,b.sendText=owner,channel,reply,text
	if b.err!=nil {return "",b.err}
	return "generated-post-id",nil
}

func authenticatedAgentHandler(t *testing.T,b *fakeAgentBridge) (AgentHandler,string) {
	t.Helper()
	tokens:=&TokenManager{Store:&memoryTokenKV{data:map[string][]byte{}}}
	token,err:=tokens.Rotate("misha")
	if err!=nil {t.Fatal(err)}
	return AgentHandler{Tokens:tokens,Bridge:b,SiteURL:"https://mm.example.com"},token
}

func TestGetPostByMattermostLinkDoesNotGuessContext(t *testing.T) {
	b:=&fakeAgentBridge{}
	h,token:=authenticatedAgentHandler(t,b)
	request:=httptest.NewRequest(http.MethodGet,"/api/agent/posts/get?reference="+url.QueryEscape("https://mm.example.com/team/pl/"+examplePostID),nil)
	request.Header.Set("X-Twinmate-Token",token)
	response:=httptest.NewRecorder()
	h.ServeHTTP(response,request)
	if response.Code!=http.StatusOK || b.owner!="misha" || b.postID!=examplePostID || !bytes.Contains(response.Body.Bytes(),[]byte("broken deploy")) {
		t.Fatalf("post lookup failed: %d %s",response.Code,response.Body.String())
	}
}

func TestAgentChoosesBeforeOrAfterAndPageLimit(t *testing.T) {
	b:=&fakeAgentBridge{}
	h,token:=authenticatedAgentHandler(t,b)
	request:=httptest.NewRequest(http.MethodGet,"/api/agent/posts/list?reference="+examplePostID+"&direction=before&scope=thread&limit=7",nil)
	request.Header.Set("X-Twinmate-Token",token)
	response:=httptest.NewRecorder()
	h.ServeHTTP(response,request)
	if response.Code!=http.StatusOK || b.direction!="before" || b.scope!="thread" || b.limit!=7 || !bytes.Contains(response.Body.Bytes(),[]byte("previous discussion")) {
		t.Fatalf("pagination failed: %d %s, %+v",response.Code,response.Body.String(),b)
	}
	invalid:=httptest.NewRequest(http.MethodGet,"/api/agent/posts/list?reference="+examplePostID+"&direction=sideways",nil)
	invalid.Header.Set("X-Twinmate-Token",token)
	bad:=httptest.NewRecorder()
	h.ServeHTTP(bad,invalid)
	if bad.Code!=http.StatusBadRequest {t.Fatalf("invalid direction accepted: %d",bad.Code)}
}

func TestAgentAPIRejectsWrongOriginAndNoCredentials(t *testing.T) {
	b:=&fakeAgentBridge{}
	h,token:=authenticatedAgentHandler(t,b)
	noAuth:=httptest.NewRecorder()
	h.ServeHTTP(noAuth,httptest.NewRequest(http.MethodGet,"/api/agent/posts/get?reference="+examplePostID,nil))
	if noAuth.Code!=http.StatusUnauthorized {t.Fatalf("unauthenticated post read: %d",noAuth.Code)}
	request:=httptest.NewRequest(http.MethodGet,"/api/agent/posts/get?reference="+url.QueryEscape("https://evil.test/team/pl/"+examplePostID),nil)
	request.Header.Set("X-Twinmate-Token",token)
	response:=httptest.NewRecorder()
	h.ServeHTTP(response,request)
	if response.Code!=http.StatusBadRequest || b.owner!="" {t.Fatalf("foreign permalink accepted: %d",response.Code)}
}

func TestAgentSendDerivesOwnerFromToken(t *testing.T) {
	b:=&fakeAgentBridge{}
	h,token:=authenticatedAgentHandler(t,b)
	request:=httptest.NewRequest(http.MethodPost,"/api/agent/messages",bytes.NewBufferString(`{"channel_id":"channel-1","reply_to_post_id":"root-1","message":"investigation done"}`))
	request.Header.Set("X-Twinmate-Token",token)
	request.Header.Set("Content-Type","application/json")
	response:=httptest.NewRecorder()
	h.ServeHTTP(response,request)
	if response.Code!=http.StatusCreated || b.sendOwner!="misha" || b.sendChannel!="channel-1" || b.sendReply!="root-1" || b.sendText!="investigation done" {
		t.Fatalf("wrong publication: %d %+v",response.Code,b)
	}
	var result struct{PostID string `json:"post_id"`}
	if err:=json.Unmarshal(response.Body.Bytes(),&result);err!=nil || result.PostID!="generated-post-id" {t.Fatalf("unexpected receipt: %s %v",response.Body.String(),err)}
	spoof:=httptest.NewRequest(http.MethodPost,"/api/agent/messages",bytes.NewBufferString(`{"channel_id":"channel-1","message":"fake","owner_id":"petya"}`))
	spoof.Header.Set("X-Twinmate-Token",token)
	spoof.Header.Set("Content-Type","application/json")
	bad:=httptest.NewRecorder()
	h.ServeHTTP(bad,spoof)
	if bad.Code!=http.StatusBadRequest {t.Fatalf("owner spoofing accepted: %d",bad.Code)}
}

func TestPostReferenceParserRestrictsMattermostOrigin(t *testing.T) {
	for _,raw:=range []string{examplePostID,"https://mm.example.com/team/pl/"+examplePostID,"https://mm.example.com/team/channels/town?post_id="+examplePostID} {
		id,err:=parsePostReference(raw,"https://mm.example.com")
		if err!=nil || id!=examplePostID {t.Fatalf("%s => %s %v",raw,id,err)}
	}
	for _,raw:=range []string{"https://evil.test/team/pl/"+examplePostID,"https://mm.example.com.evil.test/team/pl/"+examplePostID,"nonsense"} {
		if _,err:=parsePostReference(raw,"https://mm.example.com");err==nil {t.Fatalf("accepted %s",raw)}
	}
}

type fakeMattermostBridgeAPI struct {
	post *model.Post
	channel *model.Channel
	owner *model.User
	allowRead,allowWrite,member bool
	thread *model.PostList
	before,after *model.PostList
}
func (f *fakeMattermostBridgeAPI) GetUser(string)(*model.User,*model.AppError) {return f.owner,nil}
func (f *fakeMattermostBridgeAPI) GetPost(string)(*model.Post,*model.AppError) {return f.post,nil}
func (f *fakeMattermostBridgeAPI) GetPostThread(string)(*model.PostList,*model.AppError) {return f.thread,nil}
func (f *fakeMattermostBridgeAPI) GetPostsBefore(string,string,int,int)(*model.PostList,*model.AppError) {return f.before,nil}
func (f *fakeMattermostBridgeAPI) GetPostsAfter(string,string,int,int)(*model.PostList,*model.AppError) {return f.after,nil}
func (f *fakeMattermostBridgeAPI) GetChannel(string)(*model.Channel,*model.AppError) {return f.channel,nil}
func (f *fakeMattermostBridgeAPI) GetChannelMember(string,string)(*model.ChannelMember,*model.AppError) {
	if !f.member {return nil,&model.AppError{Message:"not a member"}}
	return &model.ChannelMember{UserId:"misha"},nil
}
func (f *fakeMattermostBridgeAPI) HasPermissionToChannel(_,_ string,perm *model.Permission) bool {
	if perm==model.PermissionReadChannel {return f.allowRead}
	if perm==model.PermissionCreatePost {return f.allowWrite}
	return false
}

func TestBridgeDeniesDeactivatedOwnerBeforeReadingOrPosting(t *testing.T) {
	post:=&model.Post{Id:examplePostID,ChannelId:"secret-dm",UserId:"petya",Message:"private"}
	f:=&fakeMattermostBridgeAPI{post:post,owner:&model.User{Id:"misha",DeleteAt:123},channel:&model.Channel{Id:"secret-dm",Type:model.ChannelTypeDirect},allowRead:true,allowWrite:true,member:true,thread:&model.PostList{Order:[]string{examplePostID},Posts:map[string]*model.Post{examplePostID:post}}}
	b:=MattermostBridge{API:f,Publisher:func(string,string,string,string)(string,error){t.Fatal("published as deactivated owner");return "",nil}}
	if _,err:=b.Get("misha",examplePostID);!errors.Is(err,ErrAgentForbidden) {t.Fatalf("deleted user read private context: %v",err)}
	if _,err:=b.Send("misha","secret-dm","","hi");!errors.Is(err,ErrAgentForbidden) {t.Fatalf("deleted user posted: %v",err)}
}

func TestBridgeDeniesArchivedChannelWrites(t *testing.T) {
	f:=&fakeMattermostBridgeAPI{owner:&model.User{Id:"misha"},channel:&model.Channel{Id:"archived",DeleteAt:123,Type:model.ChannelTypeOpen},allowWrite:true,member:true}
	b:=MattermostBridge{API:f,Publisher:func(string,string,string,string)(string,error){t.Fatal("published in archived channel");return "",nil}}
	if _,err:=b.Send("misha","archived","","hi");!errors.Is(err,ErrAgentForbidden) {t.Fatalf("archived channel accepted: %v",err)}
}

func TestBridgeChecksReadMembershipAndWriteReplyChannel(t *testing.T) {
	post:=&model.Post{Id:examplePostID,ChannelId:"secret-dm",UserId:"petya",Message:"private"}
	f:=&fakeMattermostBridgeAPI{post:post,owner:&model.User{Id:"misha"},channel:&model.Channel{Id:"secret-dm",Type:model.ChannelTypeDirect},allowRead:true,allowWrite:true,member:false,thread:&model.PostList{Order:[]string{examplePostID},Posts:map[string]*model.Post{examplePostID:post}}}
	b:=MattermostBridge{API:f,Publisher:func(string,string,string,string)(string,error){return "posted",nil}}
	if _,err:=b.Get("misha",examplePostID);!errors.Is(err,ErrAgentForbidden) {t.Fatalf("nonmember read private DM: %v",err)}
	f.member=true
	if _,err:=b.Send("misha","different-channel",examplePostID,"hello");!errors.Is(err,ErrAgentBadInput) {t.Fatalf("cross-channel reply accepted: %v",err)}
	f.allowWrite=false
	if _,err:=b.Send("misha","secret-dm",examplePostID,"hello");!errors.Is(err,ErrAgentForbidden) {t.Fatalf("no write permission accepted: %v",err)}
}

func TestBridgePaginatesBeforeAndAfterInChannel(t *testing.T) {
	anchor:=&model.Post{Id:examplePostID,ChannelId:"public",CreateAt:100}
	prev:=&model.Post{Id:"previous",ChannelId:"public",Message:"older",CreateAt:90}
	next:=&model.Post{Id:"next",ChannelId:"public",Message:"newer",CreateAt:110}
	f:=&fakeMattermostBridgeAPI{post:anchor,owner:&model.User{Id:"misha"},channel:&model.Channel{Id:"public",Type:model.ChannelTypeOpen},allowRead:true,member:true,
		before:&model.PostList{Order:[]string{"previous"},Posts:map[string]*model.Post{"previous":prev}},
		after:&model.PostList{Order:[]string{"next"},Posts:map[string]*model.Post{"next":next}}}
	b:=MattermostBridge{API:f}
	before,err:=b.List("misha",examplePostID,"before","channel",10)
	if err!=nil || len(before.Posts)!=1 || before.Posts[0].ID!="previous" {t.Fatalf("before failed: %+v %v",before,err)}
	after,err:=b.List("misha",examplePostID,"after","channel",10)
	if err!=nil || len(after.Posts)!=1 || after.Posts[0].ID!="next" {t.Fatalf("after failed: %+v %v",after,err)}
}

func TestBridgePagesThreadWithoutGuessingAContextWindow(t *testing.T) {
	root:=&model.Post{Id:"root",ChannelId:"public",CreateAt:1}
	anchor:=&model.Post{Id:examplePostID,ChannelId:"public",RootId:"root",CreateAt:100}
	old:=&model.Post{Id:"old",ChannelId:"public",RootId:"root",Message:"first",CreateAt:2}
	near:=&model.Post{Id:"near",ChannelId:"public",RootId:"root",Message:"second",CreateAt:99}
	newer:=&model.Post{Id:"newer",ChannelId:"public",RootId:"root",Message:"third",CreateAt:101}
	other:=&model.Post{Id:"other-thread",ChannelId:"public",RootId:"other",Message:"irrelevant",CreateAt:105}
	deleted:=&model.Post{Id:"deleted",ChannelId:"public",RootId:"root",Message:"deleted",DeleteAt:12,CreateAt:102}
	f:=&fakeMattermostBridgeAPI{post:anchor,owner:&model.User{Id:"misha"},channel:&model.Channel{Id:"public",Type:model.ChannelTypeOpen},allowRead:true,
		thread:&model.PostList{Order:[]string{"deleted","newer","anchor","near","root","old","other-thread"},Posts:map[string]*model.Post{"root":root,"anchor":anchor,"near":near,"old":old,"newer":newer,"deleted":deleted,"other-thread":other}}}
	b:=MattermostBridge{API:f}
	page,err:=b.List("misha",examplePostID,"before","thread",1)
	if err!=nil || len(page.Posts)!=1 || page.Posts[0].ID!="near" || !page.HasMore || page.NextCursor!="near" {t.Fatalf("wrong before page: %+v %v",page,err)}
	page,err=b.List("misha",examplePostID,"after","thread",2)
	if err!=nil || len(page.Posts)!=1 || page.Posts[0].ID!="newer" || page.HasMore {t.Fatalf("wrong after page: %+v %v",page,err)}
}

func TestPersistedTwinmateLabelIsLanguageNeutral(t *testing.T) {
	if !strings.Contains(twinmateLabel,"AI-generated") || strings.Contains(twinmateLabel,"двойник") {t.Fatalf("mobile-visible label must work across language settings: %q",twinmateLabel)}
}
