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

type disclosureAPI struct {plugin.API; notices int; notice *model.Post; values map[string][]byte; failNotice bool}
func (f *disclosureAPI) GetChannel(string)(*model.Channel,*model.AppError) {return &model.Channel{Id:"dm",Type:model.ChannelTypeDirect},nil}
func (f *disclosureAPI) GetChannelMember(string,string)(*model.ChannelMember,*model.AppError) {return &model.ChannelMember{},nil}
func (f *disclosureAPI) KVGet(key string)([]byte,*model.AppError) {return f.values[key],nil}
func (f *disclosureAPI) KVSet(key string,data []byte)*model.AppError {
    if f.values==nil {f.values=make(map[string][]byte)}
    f.values[key]=append([]byte(nil),data...)
    return nil
}
func (f *disclosureAPI) GetPost(id string)(*model.Post,*model.AppError) {
    if f.notice==nil || f.notice.Id!=id {return nil,&model.AppError{Message:"notice unavailable"}}
    return f.notice,nil
}
func (f *disclosureAPI) CreatePost(post *model.Post)(*model.Post,*model.AppError) {
    if f.failNotice {return nil,&model.AppError{Message:"failed"}}
    if post.UserId!="owner" || post.ChannelId!="dm" || post.Message=="" {panic("invalid disclosure")}
    f.notices++
    f.notice=&model.Post{
        Id:model.NewId(), UserId:post.UserId, ChannelId:post.ChannelId,
        Message:post.Message, Props:post.Props,
    }
    return f.notice,nil
}

func TestOptInDisclosesExternalAgentBeforeForwarding(t *testing.T) {
    api:=&disclosureAPI{}
    p:=&Plugin{MattermostPlugin:plugin.MattermostPlugin{API:api}}
    args:=&model.CommandArgs{Command:"/twinmate on",UserId:"owner",ChannelId:"dm"}
    _,err:=p.ExecuteCommand(nil,args)
    enabled,readErr:=p.DMEnabled(nil,"owner","dm")
    if err!=nil || readErr!=nil || api.notices!=1 || !enabled {t.Fatalf("DM was enabled without public disclosure: notices=%d enabled=%t err=%v readErr=%v",api.notices,enabled,err,readErr)}
    _,_ = p.ExecuteCommand(nil,args)
    if api.notices!=1 {t.Fatal("repeated on posted duplicate disclosure")}
}

func TestDisclosureFailureKeepsDMDisabled(t *testing.T) {
    api:=&disclosureAPI{failNotice:true}
    p:=&Plugin{MattermostPlugin:plugin.MattermostPlugin{API:api}}
    _,_ = p.ExecuteCommand(nil,&model.CommandArgs{Command:"/twinmate on",UserId:"owner",ChannelId:"dm"})
    if enabled,_:=p.DMEnabled(nil,"owner","dm");enabled {t.Fatal("unsafe opt-in persisted despite disclosure failure")}
}

func TestEditedOrDeletedDMDisclosureStopsForwarding(t *testing.T) {
    for _,change:=range []struct{name string;mutate func(*model.Post)}{
        {"edited",func(post *model.Post){post.Message="No AI forwarding happens here"}},
        {"deleted",func(post *model.Post){post.DeleteAt=123}},
        {"wrong owner",func(post *model.Post){post.UserId="another-user"}},
    } {
        t.Run(change.name,func(t *testing.T){
            api:=&disclosureAPI{}
            p:=&Plugin{MattermostPlugin:plugin.MattermostPlugin{API:api}}
            _,_ = p.ExecuteCommand(nil,&model.CommandArgs{Command:"/twinmate on",UserId:"owner",ChannelId:"dm"})
            if enabled,_:=p.DMEnabled(nil,"owner","dm");!enabled {t.Fatal("fixture failed to enable forwarding")}
            change.mutate(api.notice)
            if enabled,_:=p.DMEnabled(nil,"owner","dm");enabled {t.Fatal("changed disclosure must not authorize external forwarding")}
        })
    }
}

func TestLegacyUnverifiableDMStateFailsClosed(t *testing.T) {
    api:=&disclosureAPI{values:map[string][]byte{dmKey("owner","dm"):[]byte("on")}}
    p:=&Plugin{MattermostPlugin:plugin.MattermostPlugin{API:api}}
    if enabled,_:=p.DMEnabled(nil,"owner","dm");enabled {t.Fatal("legacy opt-in has no verifiable notice")}
}
