package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type memoryTokenKV struct { data map[string][]byte }

func (s *memoryTokenKV) Load(key string) ([]byte,error) { return s.data[key],nil }
func (s *memoryTokenKV) Save(key string,value []byte) error { s.data[key]=append([]byte(nil),value...);return nil }
func (s *memoryTokenKV) Remove(key string) error { delete(s.data,key);return nil }

func TestTokenManagerRotateAndRevoke(t *testing.T) {
	store:=&memoryTokenKV{data:map[string][]byte{}}
	manager:=TokenManager{Store:store}
	first,err:=manager.Rotate("misha")
	if err!=nil || !strings.HasPrefix(first,"tmt_") {t.Fatalf("unexpected initial token: %v",err)}
	if strings.Contains(string(bytes.Join(values(store.data),[]byte("/"))),first) {t.Fatal("raw token stored in KV")}
	owner,err:=manager.Authenticate(first)
	if err!=nil || owner!="misha" {t.Fatalf("token cannot authenticate owner: %s, %v",owner,err)}
	second,err:=manager.Rotate("misha")
	if err!=nil || first==second {t.Fatalf("rotation returned same token: %v",err)}
	if _,err:=manager.Authenticate(first);!errors.Is(err,ErrInvalidAgentToken) {t.Fatalf("old token still valid: %v",err)}
	if owner,err:=manager.Authenticate(second);err!=nil || owner!="misha" {t.Fatalf("new token invalid: %v",err)}
	if err:=manager.Revoke("misha");err!=nil {t.Fatal(err)}
	if _,err:=manager.Authenticate(second);!errors.Is(err,ErrInvalidAgentToken) {t.Fatalf("revoked token still valid: %v",err)}
}

func values(m map[string][]byte) [][]byte { out:=make([][]byte,0,len(m));for _,v:=range m {out=append(out,v)};return out }

func TestTokenManagerRejectsCorruptOrForgedTokens(t *testing.T) {
	manager:=TokenManager{Store:&memoryTokenKV{data:map[string][]byte{}}}
	valid,err:=manager.Rotate("misha")
	if err!=nil {t.Fatal(err)}
	for _,raw:=range []string{"", "fake", "tmt_x_y",valid+"tamper"} {
		if _,err:=manager.Authenticate(raw);!errors.Is(err,ErrInvalidAgentToken) {t.Fatalf("accepted invalid %q: %v",raw,err)}
	}
}

func TestPreScopeTokensRemainUsableUntilRotation(t *testing.T) {
    store:=&memoryTokenKV{data:map[string][]byte{}}
    manager:=&TokenManager{Store:store}
    token,err:=manager.Rotate("misha")
    if err!=nil {t.Fatal(err)}
    id:=strings.Split(token,"_")[1]
    var original storedAgentToken
    if err:=json.Unmarshal(store.data[agentTokenKey(id)],&original);err!=nil {t.Fatal(err)}
    legacy,_:=json.Marshal(storedAgentToken{OwnerID:original.OwnerID,Digest:original.Digest})
    store.data[agentTokenKey(id)]=legacy
    claims,err:=manager.AuthenticateClaims(token)
    if err!=nil || !claims.AllowSend || !claims.AllowsChannel("any-channel") {t.Fatalf("legacy key unexpectedly lost capabilities: %+v %v",claims,err)}
}

func TestUserTokenAPIRequiresSessionAndNeverReturnsSecretOnGET(t *testing.T) {
	manager:=&TokenManager{Store:&memoryTokenKV{data:map[string][]byte{}}}
	h:=TokenSettingsHandler{Manager:manager}
	unauth:=httptest.NewRecorder()
	h.ServeHTTP(unauth,httptest.NewRequest(http.MethodPost,"/api/agent/token",nil))
	if unauth.Code!=http.StatusUnauthorized {t.Fatalf("unauthenticated issuance: %d",unauth.Code)}
	post:=httptest.NewRequest(http.MethodPost,"/api/agent/token",strings.NewReader("{}"))
	post.Header.Set("Mattermost-User-Id","misha")
	post.Header.Set("Content-Type","application/json")
	issued:=httptest.NewRecorder()
	h.ServeHTTP(issued,post)
	if issued.Code!=http.StatusOK || !strings.Contains(issued.Body.String(),"tmt_") {t.Fatalf("token not issued: %d %s",issued.Code,issued.Body.String())}
	get:=httptest.NewRequest(http.MethodGet,"/api/agent/token",nil)
	get.Header.Set("Mattermost-User-Id","misha")
	status:=httptest.NewRecorder()
	h.ServeHTTP(status,get)
	if status.Code!=http.StatusOK || !strings.Contains(status.Body.String(),`"has_token":true`) || strings.Contains(status.Body.String(),"tmt_") {
		t.Fatalf("secret leaked in status: %d %s",status.Code,status.Body.String())
	}
	revoke:=httptest.NewRequest(http.MethodDelete,"/api/agent/token",strings.NewReader("{}"))
	revoke.Header.Set("Mattermost-User-Id","misha")
	revoke.Header.Set("Content-Type","application/json")
	deleted:=httptest.NewRecorder()
	h.ServeHTTP(deleted,revoke)
	if deleted.Code!=http.StatusOK {t.Fatalf("cannot revoke: %d",deleted.Code)}
}

type failedTokenStore struct{}
func (failedTokenStore) Load(string)([]byte,error) {return nil,nil}
func (failedTokenStore) Save(string,[]byte)error {return errors.New("storage offline")}
func (failedTokenStore) Remove(string)error {return nil}

func TestTokenIssuanceStorageFailureIsServerError(t *testing.T) {
    h:=TokenSettingsHandler{Manager:&TokenManager{Store:failedTokenStore{}}}
    req:=httptest.NewRequest(http.MethodPost,"/api/agent/token",strings.NewReader(`{"allow_send":true}`))
    req.Header.Set("Mattermost-User-Id","owner")
    req.Header.Set("Content-Type","application/json")
    result:=httptest.NewRecorder()
    h.ServeHTTP(result,req)
    if result.Code!=http.StatusInternalServerError {t.Fatalf("storage outage reported as caller error: %d",result.Code)}
}
