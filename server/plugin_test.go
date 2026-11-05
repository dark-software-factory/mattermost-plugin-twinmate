package main

import (
	"testing"
	"encoding/json"

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
	origin,_:=json.Marshal(storedPostProvenance{OwnerID:"person",ChannelID:"dm"})
	p:=&Plugin{MattermostPlugin:plugin.MattermostPlugin{API:&perPostProvenanceAPI{records:map[string][]byte{"provenance:twin-1":origin}}}}
	human := &model.User{IsBot: false}
	bot := &model.User{IsBot: true}
	plain := &model.Post{Id: "human-1", UserId: "person", ChannelId: "dm", Message: "Hello"}
	if !p.shouldProcessPost(plain, human) {
		t.Fatal("human DM must be forwarded")
	}
	if p.shouldProcessPost(plain, bot) {
		t.Fatal("bot message must not be sent to personal agents")
	}
	system := &model.Post{Id: "system-1", UserId: "person", ChannelId: "dm", Type: "system_join_channel", Message: "Hello"}
	if p.shouldProcessPost(system, human) {
		t.Fatal("system messages must not be sent")
	}
	generated := &model.Post{Id: "twin-1", UserId: "person", ChannelId: "dm", Message: twinmateLabel + "AI reply"}
	if p.shouldProcessPost(generated, human) {
		t.Fatal("Twinmate replies must not recursively trigger agents")
	}
	forgedProps := &model.Post{Id: "human-2", UserId: "person", ChannelId: "dm", Message: "Human text", Props: model.StringInterface{"twinmate": map[string]any{"generated": true}}}
	if !p.shouldProcessPost(forgedProps, human) {
		t.Fatal("client-supplied props are not trusted provenance")
	}
	webhook := &model.Post{Id: "webhook-1", UserId: "person", ChannelId: "dm", Message: "Webhook notification", Props: model.StringInterface{"from_webhook": "true"}}
	if p.shouldProcessPost(webhook, human) {
		t.Fatal("integration-originated webhook posts must not trigger personal agents")
	}
}

type perPostProvenanceAPI struct {plugin.API; records map[string][]byte}
func (f *perPostProvenanceAPI) KVGet(key string)([]byte,*model.AppError) {return f.records[key],nil}

func TestHumanTextMimickingLabelIsNotTrustedAsProvenance(t *testing.T) {
    p:=&Plugin{MattermostPlugin:plugin.MattermostPlugin{API:&openChannelProvenanceAPI{readable:true}}}
    plain:=&model.Post{Id:"human-forged",UserId:"human",ChannelId:"dm",Message:twinmateLabel+"Actual human request"}
    if !p.shouldProcessPost(plain,&model.User{Id:"human"}) {
        t.Fatal("untrusted post body suppressed a legitimate human message")
    }
}

func TestGeneratedPostMarkerCannotBeEditedAway(t *testing.T) {
    api:=&openChannelProvenanceAPI{readable:true}
    p:=&Plugin{MattermostPlugin:plugin.MattermostPlugin{API:api}}
    oldPost:=&model.Post{Id:"created-once",UserId:"misha",ChannelId:"open-channel",Message:twinmateLabel+"done"}
    edited:=&model.Post{Id:oldPost.Id,UserId:oldPost.UserId,ChannelId:oldPost.ChannelId,Message:"done"}
    _,reason:=p.MessageWillBeUpdated(nil,edited,oldPost)
    if reason=="" {t.Fatal("author could edit away server-recorded AI attribution")}
}

type dmAuthorizationAPI struct {plugin.API; ownerDeleted,writeAllowed bool; posted bool}
func (f *dmAuthorizationAPI) GetChannel(string)(*model.Channel,*model.AppError) {return &model.Channel{Id:"dm",Type:model.ChannelTypeDirect},nil}
func (f *dmAuthorizationAPI) GetChannelMembers(string,int,int)(model.ChannelMembers,*model.AppError) {
    return model.ChannelMembers{{UserId:"sender"},{UserId:"owner"}},nil
}
func (f *dmAuthorizationAPI) GetUser(id string)(*model.User,*model.AppError) {
    u:=&model.User{Id:id}
    if id=="owner" && f.ownerDeleted {u.DeleteAt=3}
    return u,nil
}
func (f *dmAuthorizationAPI) HasPermissionToChannel(_, _ string, perm *model.Permission) bool {
    return perm==model.PermissionReadChannel || (perm==model.PermissionCreatePost && f.writeAllowed)
}
func (f *dmAuthorizationAPI) GetChannelMember(_, user string)(*model.ChannelMember,*model.AppError) {return &model.ChannelMember{UserId:user},nil}
func (f *dmAuthorizationAPI) CreatePost(post *model.Post)(*model.Post,*model.AppError) {f.posted=true;return post,nil}

func TestAutomaticRepliesRecheckOwnerActivityAndWritePermissions(t *testing.T) {
    api:=&dmAuthorizationAPI{ownerDeleted:true,writeAllowed:true}
    p:=&Plugin{MattermostPlugin:plugin.MattermostPlugin{API:api}}
    if owner,_:=p.Recipient(nil,"dm","sender");owner!="" {t.Fatalf("deactivated owner received external DM event: %s",owner)}
    if err:=p.Publish(nil,"owner","dm","","response","trigger");err==nil || api.posted {t.Fatalf("published as deactivated owner: %v",err)}
    api.ownerDeleted=false;api.writeAllowed=false
    if err:=p.Publish(nil,"owner","dm","","response","trigger");err==nil || api.posted {t.Fatalf("published without current permission: %v",err)}
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
