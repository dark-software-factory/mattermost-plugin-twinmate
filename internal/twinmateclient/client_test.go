package twinmateclient

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestClientGetsThreadUsingTwinmateCredential(t *testing.T) {
	server:=httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request) {
		if r.URL.Path!="/plugins/com.darksoftwarefactory.twinmate/api/agent/context" || r.Header.Get("X-Twinmate-Token")!="tmt_test" || r.URL.Query().Get("url")!="https://mm.example.com/team/pl/post-1" {
			t.Errorf("unexpected client request: %s %s",r.URL.String(),r.Header.Get("X-Twinmate-Token"))
		}
		_,_=w.Write([]byte(`{"channel_id":"dm-1","target_post_id":"post-1","posts":[{"id":"post-1","message":"deploy broken","user_id":"petya","create_at":123}]}`))
	}))
	defer server.Close()
	client,err:=New(server.URL,"tmt_test",server.Client())
	if err!=nil {t.Fatal(err)}
	result,err:=client.GetContext(context.Background(),"https://mm.example.com/team/pl/post-1")
	if err!=nil || result.TargetPostID!="post-1" || len(result.Posts)!=1 || result.Posts[0].Message!="deploy broken" {
		t.Fatalf("unexpected context: %+v %v",result,err)
	}
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
	_,err=client.GetContext(context.Background(),"post-1")
	if err==nil || !strings.Contains(err.Error(),"redirect") {t.Fatalf("redirect followed or not reported: %v",err)}
}

func TestClientRejectsInsecureServerURL(t *testing.T) {
	if _,err:=New("http://mm.example.com","token",nil);err==nil {t.Fatal("plain HTTP may leak agent token")}
	if _,err:=New("https://mm.example.com","",nil);err==nil {t.Fatal("missing token accepted")}
}
