package core

import (
	"context"
	"errors"
	"testing"

	"github.com/dark-software-factory/mattermost-plugin-twinmate/server/agent"
)

type fakePlatform struct {
	config Connection
	enabled bool
	recipient string
	published []publication
}

type publication struct{ owner, channel, root, text, source string }

func (f *fakePlatform) Recipient(context.Context, string, string) (string, error) { return f.recipient, nil }
func (f *fakePlatform) Connection(context.Context, string) (Connection, error) { return f.config, nil }
func (f *fakePlatform) DMEnabled(context.Context, string, string) (bool, error) { return f.enabled, nil }
func (f *fakePlatform) Publish(_ context.Context, owner, channel, root, text, source string) error {
	f.published = append(f.published, publication{owner, channel, root, text, source})
	return nil
}

type fakeAgent struct {
	events []agent.Event
	reply string
	err error
}

func (f *fakeAgent) Call(_ context.Context, endpoint, token string, event agent.Event) (string, error) {
	if endpoint != "https://agent.example.com/events" || token != "secret" {
		return "", errors.New("wrong endpoint or secret")
	}
	f.events = append(f.events, event)
	return f.reply, f.err
}

func TestRoutesHumanDMToRecipientsTwin(t *testing.T) {
	platform := &fakePlatform{config: Connection{Enabled: true, Endpoint: "https://agent.example.com/events", Token: "secret"}, recipient: "misha", enabled: true}
	remote := &fakeAgent{reply: "By Friday."}
	processor := Processor{Platform: platform, Agent: remote}
	message := Message{ID: "post-1", ChannelID: "dm-1", SenderID: "petya", Text: "When?", RootID: "thread-1"}
	if err := processor.Handle(context.Background(), message); err != nil {
		t.Fatal(err)
	}
	if len(remote.events) != 1 || remote.events[0].OwnerID != "misha" || remote.events[0].SenderID != "petya" || remote.events[0].Text != "When?" {
		t.Fatalf("unexpected delivery: %+v", remote.events)
	}
	if len(platform.published) != 1 || platform.published[0] != (publication{"misha", "dm-1", "thread-1", "By Friday.", "post-1"}) {
		t.Fatalf("unexpected publication: %+v", platform.published)
	}
}

func TestIgnoresGeneratedPostsToPreventTwinLoops(t *testing.T) {
	platform := &fakePlatform{recipient: "misha", enabled: true, config: Connection{Enabled: true}}
	remote := &fakeAgent{reply: "loop"}
	processor := Processor{Platform: platform, Agent: remote}
	if err := processor.Handle(context.Background(), Message{ID: "post-2", ChannelID: "dm-1", SenderID: "petya", Text: "AI", Generated: true}); err != nil {
		t.Fatal(err)
	}
	if len(remote.events) != 0 || len(platform.published) != 0 {
		t.Fatal("generated posts must never trigger an agent")
	}
}

func TestNeverCallsDisabledOrUnapprovedTwins(t *testing.T) {
	for _, tc := range []struct {
		name string
		enabled bool
		dmEnabled bool
	}{
		{"connection disabled", false, true},
		{"DM not opted in", true, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			platform := &fakePlatform{recipient: "misha", config: Connection{Enabled: tc.enabled, Endpoint: "https://agent.example.com/events", Token: "secret"}, enabled: tc.dmEnabled}
			remote := &fakeAgent{reply: "should not be sent"}
			if err := (&Processor{Platform: platform, Agent: remote}).Handle(context.Background(), Message{ID: "post-1", ChannelID: "dm-1", SenderID: "petya", Text: "hello"}); err != nil {
				t.Fatal(err)
			}
			if len(remote.events) != 0 || len(platform.published) != 0 {
				t.Fatal("message leaked to disabled twin")
			}
		})
	}
}

func TestEmptyAgentReplyIsNotPublished(t *testing.T) {
	platform := &fakePlatform{recipient: "misha", enabled: true, config: Connection{Enabled: true, Endpoint: "https://agent.example.com/events", Token: "secret"}}
	remote := &fakeAgent{reply: ""}
	if err := (&Processor{Platform: platform, Agent: remote}).Handle(context.Background(), Message{ID: "post-1", ChannelID: "dm-1", SenderID: "petya", Text: "hi"}); err != nil {
		t.Fatal(err)
	}
	if len(platform.published) != 0 {
		t.Fatal("empty reply must not produce a Mattermost post")
	}
}
