package main

import (
	"testing"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin"
)

type provenanceWriteFailureAPI struct {
	plugin.API
	created int
	logged bool
}
func (f *provenanceWriteFailureAPI) CreatePost(post *model.Post)(*model.Post,*model.AppError) {
	f.created++
	post.Id="created-once"
	return post,nil
}
func (f *provenanceWriteFailureAPI) KVSet(string,[]byte)*model.AppError {
	return &model.AppError{Message:"temporary storage failure"}
}
func (f *provenanceWriteFailureAPI) LogError(string,...any) {f.logged=true}

func TestPublishedPostIsNotReportedAsFailedWhenOnlyBadgeProvenanceFails(t *testing.T) {
	api:=&provenanceWriteFailureAPI{}
	p:=&Plugin{MattermostPlugin:plugin.MattermostPlugin{API:api}}
	id,err:=p.createTwinPost("misha","channel","","work done","")
	if err!=nil || id!="created-once" || api.created!=1 || !api.logged {
		t.Fatalf("published post could be duplicated on retry: %q %v, created=%d logged=%t",id,err,api.created,api.logged)
	}
}

type openChannelProvenanceAPI struct {plugin.API; readable bool}
func (f *openChannelProvenanceAPI) KVGet(string)([]byte,*model.AppError) {
	return []byte(`{"owner_id":"misha","channel_id":"open-channel"}`),nil
}
func (f *openChannelProvenanceAPI) GetChannel(string)(*model.Channel,*model.AppError) {
	return &model.Channel{Id:"open-channel",Type:model.ChannelTypeOpen},nil
}
func (f *openChannelProvenanceAPI) HasPermissionToChannel(string,string,*model.Permission)bool {return f.readable}
func (f *openChannelProvenanceAPI) GetChannelMember(string,string)(*model.ChannelMember,*model.AppError) {
	panic("public-channel provenance must not require membership")
}
func (f *openChannelProvenanceAPI) GetPost(string)(*model.Post,*model.AppError) {
	return &model.Post{Id:"created-once",UserId:"misha",ChannelId:"open-channel",Message:twinmateLabel+"done"},nil
}

func TestVerifiedBadgeUsesReadPermissionForOpenChannels(t *testing.T) {
	f:=&openChannelProvenanceAPI{readable:true}
	p:=&Plugin{MattermostPlugin:plugin.MattermostPlugin{API:f}}
	verified,err:=p.verifiedPostForViewer("created-once","viewer")
	if err!=nil || !verified {t.Fatalf("readable open-channel badge hidden: %t %v",verified,err)}
	f.readable=false
	verified,err=p.verifiedPostForViewer("created-once","viewer")
	if err!=nil || verified {t.Fatalf("unreadable open-channel badge disclosed: %t %v",verified,err)}
}

func TestOnlyHumanPostsAreForwarded(t *testing.T) {
	human := &model.User{IsBot: false}
	bot := &model.User{IsBot: true}
	plain := &model.Post{Id: "human-1", UserId: "person", ChannelId: "dm", Message: "Hello"}
	if !shouldProcessPost(plain, human) {
		t.Fatal("human DM must be forwarded")
	}
	if shouldProcessPost(plain, bot) {
		t.Fatal("bot message must not be sent to personal agents")
	}
	system := &model.Post{Id: "system-1", UserId: "person", ChannelId: "dm", Type: "system_join_channel", Message: "Hello"}
	if shouldProcessPost(system, human) {
		t.Fatal("system messages must not be sent")
	}
	generated := &model.Post{Id: "twin-1", UserId: "person", ChannelId: "dm", Message: twinmateLabel + "AI reply"}
	if shouldProcessPost(generated, human) {
		t.Fatal("Twinmate replies must not recursively trigger agents")
	}
	forgedProps := &model.Post{Id: "human-2", UserId: "person", ChannelId: "dm", Message: "Human text", Props: model.StringInterface{"twinmate": map[string]any{"generated": true}}}
	if !shouldProcessPost(forgedProps, human) {
		t.Fatal("client-supplied props are not trusted provenance")
	}
	webhook := &model.Post{Id: "webhook-1", UserId: "person", ChannelId: "dm", Message: "Webhook notification", Props: model.StringInterface{"from_webhook": "true"}}
	if shouldProcessPost(webhook, human) {
		t.Fatal("integration-originated webhook posts must not trigger personal agents")
	}
}

func TestPostProvenanceRequiresMatchingServerRecord(t *testing.T) {
	post := &model.Post{Id: "twin-1", UserId: "misha", ChannelId: "dm-1", Message: twinmateLabel + "AI reply"}
	if !postMatchesProvenance(post, storedPostProvenance{OwnerID: "misha", ChannelID: "dm-1"}) {
		t.Fatal("server-recorded AI post was not recognized")
	}
	if postMatchesProvenance(post, storedPostProvenance{OwnerID: "petya", ChannelID: "dm-1"}) ||
		postMatchesProvenance(post, storedPostProvenance{OwnerID: "misha", ChannelID: "other-dm"}) {
		t.Fatal("provenance record must match post author and channel")
	}
	plain := &model.Post{Id: "human-2", UserId: "misha", ChannelId: "dm-1", Message: "Normal message"}
	if postMatchesProvenance(plain, storedPostProvenance{OwnerID: "misha", ChannelID: "dm-1"}) {
		t.Fatal("ordinary message must not be certified as an AI post")
	}
}
