package main

import (
    "testing"
    "github.com/mattermost/mattermost/server/public/model"
    "github.com/mattermost/mattermost/server/public/plugin"
)

func TestParseTwinmateCommand(t *testing.T) {
	for _, tc := range []struct{ input, want string }{
		{"/twinmate", "status"}, {"/twinmate on", "on"}, {"/twinmate off", "off"}, {"/twinmate status", "status"},
		{"/twinmate delete all", "invalid"}, {"/twinmate unknown", "invalid"},
	} {
		if got := parseTwinmateCommand(tc.input); got != tc.want {
			t.Errorf("%q: got %q, want %q", tc.input, got, tc.want)
		}
	}
}

type disclosureAPI struct {plugin.API; notices int; enabled bool; failNotice bool}
func (f *disclosureAPI) GetChannel(string)(*model.Channel,*model.AppError) {return &model.Channel{Id:"dm",Type:model.ChannelTypeDirect},nil}
func (f *disclosureAPI) GetChannelMember(string,string)(*model.ChannelMember,*model.AppError) {return &model.ChannelMember{},nil}
func (f *disclosureAPI) KVGet(string)([]byte,*model.AppError) {if f.enabled{return []byte("on"),nil};return nil,nil}
func (f *disclosureAPI) KVSet(_ string,data []byte)*model.AppError {f.enabled=string(data)=="on";return nil}
func (f *disclosureAPI) CreatePost(post *model.Post)(*model.Post,*model.AppError) {
    if f.failNotice {return nil,&model.AppError{Message:"failed"}}
    if post.UserId!="owner" || post.ChannelId!="dm" || post.Message=="" {panic("invalid disclosure")}
    f.notices++
    return post,nil
}

func TestOptInDisclosesExternalAgentBeforeForwarding(t *testing.T) {
    api:=&disclosureAPI{}
    p:=&Plugin{MattermostPlugin:plugin.MattermostPlugin{API:api}}
    args:=&model.CommandArgs{Command:"/twinmate on",UserId:"owner",ChannelId:"dm"}
    _,err:=p.ExecuteCommand(nil,args)
    if err!=nil || api.notices!=1 || !api.enabled {t.Fatalf("DM was enabled without public disclosure: notices=%d enabled=%t err=%v",api.notices,api.enabled,err)}
    _,_ = p.ExecuteCommand(nil,args)
    if api.notices!=1 {t.Fatal("repeated on posted duplicate disclosure")}
}

func TestDisclosureFailureKeepsDMDisabled(t *testing.T) {
    api:=&disclosureAPI{failNotice:true}
    p:=&Plugin{MattermostPlugin:plugin.MattermostPlugin{API:api}}
    _,_ = p.ExecuteCommand(nil,&model.CommandArgs{Command:"/twinmate on",UserId:"owner",ChannelId:"dm"})
    if api.enabled {t.Fatal("unsafe opt-in persisted despite disclosure failure")}
}
