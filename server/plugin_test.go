package main

import (
	"testing"

	"github.com/mattermost/mattermost/server/public/model"
)

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
