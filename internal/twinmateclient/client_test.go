package twinmateclient

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestClientGetsOnePostUsingTwinmateCredential(t *testing.T) {
	server:=httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request) {
		if r.URL.Path!="/plugins/com.darksoftwarefactory.twinmate/api/agent/posts/get" || r.Header.Get("X-Twinmate-Token")!="tmt_test" || r.URL.Query().Get("reference")!="https://mm.example.com/team/pl/post-1" {
			t.Errorf("unexpected client request: %s %s",r.URL.String(),r.Header.Get("X-Twinmate-Token"))
		}
		_,_=w.Write([]byte(`{"channel_id":"dm-1","id":"post-1","message":"deploy broken","user_id":"petya","create_at":123}`))
	}))
	defer server.Close()
	client,err:=New(server.URL,"tmt_test",server.Client())
	if err!=nil {t.Fatal(err)}
	result,err:=client.GetPost(context.Background(),"https://mm.example.com/team/pl/post-1")
	if err!=nil || result.ID!="post-1" || result.Message!="deploy broken" || result.ChannelID!="dm-1" {
		t.Fatalf("unexpected post: %+v %v",result,err)
	}
}

func TestClientListsMessagesBeforeOrAfterWithExplicitPagination(t *testing.T) {
	server:=httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request) {
		query:=r.URL.Query()
		if r.URL.Path!="/plugins/com.darksoftwarefactory.twinmate/api/agent/posts/list" || query.Get("reference")!="post-1" || query.Get("direction")!="after" || query.Get("scope")!="thread" || query.Get("limit")!="7" {t.Errorf("wrong page request: %s",r.URL.String())}
		_,_=w.Write([]byte(`{"channel_id":"dm-1","anchor_post_id":"post-1","posts":[{"id":"post-2","message":"found it"}],"has_more":true,"next_cursor":"post-2"}`))
	}))
	defer server.Close()
	client,err:=New(server.URL,"tmt_test",server.Client())
	if err!=nil {t.Fatal(err)}
	page,err:=client.ListPosts(context.Background(),"post-1","after","thread",7)
	if err!=nil || !page.HasMore || page.NextCursor!="post-2" || len(page.Posts)!=1 {t.Fatalf("unexpected page: %+v %v",page,err)}
}

func TestClientSendsAgentMessageWithoutMattermostPAT(t *testing.T) {
	server:=httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request) {
		if r.Method!=http.MethodPost || r.Header.Get("X-Twinmate-Token")!="tmt_test" || r.Header.Get("Authorization")!="" {
			t.Errorf("wrong authentication: %+v",r.Header)
		}
		var request SendRequest
		if err:=json.NewDecoder(r.Body).Decode(&request);err!=nil {t.Error(err)}
		if request.Message!="done" || request.ReplyToPostID!="root-1" {t.Errorf("wrong reply: %+v",request)}
		w.WriteHeader(http.StatusCreated)
		_,_=w.Write([]byte(`{"post_id":"post-2"}`))
	}))
	defer server.Close()
	client,err:=New(server.URL,"tmt_test",server.Client())
	if err!=nil {t.Fatal(err)}
	result,err:=client.SendMessage(context.Background(),SendRequest{ReplyToPostID:"root-1",Message:"done"})
	if err!=nil || result.PostID!="post-2" {t.Fatalf("unexpected receipt: %+v %v",result,err)}
}

func TestClientDoesNotFollowRedirectWithAgentSecret(t *testing.T) {
	server:=httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request) {
		http.Redirect(w,r,"https://external.example/steal",http.StatusTemporaryRedirect)
	}))
	defer server.Close()
	client,err:=New(server.URL,"tmt_secret",server.Client())
	if err!=nil {t.Fatal(err)}
	_,err=client.GetPost(context.Background(),"post-1")
	if err==nil || !strings.Contains(err.Error(),"redirect") {t.Fatalf("redirect followed or not reported: %v",err)}
}

func TestClientRejectsInsecureServerURL(t *testing.T) {
	if _,err:=New("http://mm.example.com","token",nil);err==nil {t.Fatal("plain HTTP may leak agent token")}
	if _,err:=New("https://mm.example.com","",nil);err==nil {t.Fatal("missing token accepted")}
}

func TestClientAcceptsLargeButValidPostPage(t *testing.T) {
	posts:=make([]ContextPost,50)
	// Encoding escapes '<', expanding a legitimate 15KB post to about
	// 90KB on the wire. Fifty such posts are still a valid paginated page.
	for i:=range posts {posts[i]=ContextPost{ID:"post",Message:strings.Repeat("<",15000)}}
	server:=httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){
		_ = json.NewEncoder(w).Encode(MessagePage{AnchorPostID:"post",Posts:posts})
	}))
	defer server.Close()
	client,err:=New(server.URL,"tmt_test",server.Client())
	if err!=nil {t.Fatal(err)}
	result,err:=client.ListPosts(context.Background(),"post","before","channel",50)
	if err!=nil || len(result.Posts)!=50 {t.Fatalf("valid page truncated: %d posts, %v",len(result.Posts),err)}
}
