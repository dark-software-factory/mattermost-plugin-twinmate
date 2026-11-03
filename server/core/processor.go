package core

import (
	"context"
	"strings"

	"github.com/dark-software-factory/mattermost-plugin-twinmate/server/agent"
)

type Connection struct {
	Enabled  bool   `json:"enabled"`
	Endpoint string `json:"endpoint"`
	Token    string `json:"token"`
}

type Message struct {
	ID        string
	ChannelID string
	SenderID  string
	RootID    string
	Text      string
	Generated bool
}

type Platform interface {
	Recipient(ctx context.Context, channelID, senderID string) (string, error)
	Connection(ctx context.Context, ownerID string) (Connection, error)
	DMEnabled(ctx context.Context, ownerID, channelID string) (bool, error)
	Publish(ctx context.Context, ownerID, channelID, rootID, text, sourceID string) error
}

type Agent interface {
	Call(ctx context.Context, endpoint, token string, event agent.Event) (string, error)
}

type Processor struct {
	Platform Platform
	Agent    Agent
}

func (p *Processor) Handle(ctx context.Context, message Message) error {
	if message.Generated || message.ID == "" || strings.TrimSpace(message.Text) == "" {
		return nil
	}
	ownerID, err := p.Platform.Recipient(ctx, message.ChannelID, message.SenderID)
	if err != nil || ownerID == "" || ownerID == message.SenderID {
		return err
	}
	config, err := p.Platform.Connection(ctx, ownerID)
	if err != nil {
		return err
	}
	if !config.Enabled || config.Endpoint == "" {
		return nil
	}
	enabled, err := p.Platform.DMEnabled(ctx, ownerID, message.ChannelID)
	if err != nil || !enabled {
		return err
	}
	reply, err := p.Agent.Call(ctx, config.Endpoint, config.Token, agent.Event{
		EventID:   message.ID + ":" + ownerID,
		OwnerID:   ownerID,
		SenderID:  message.SenderID,
		ChannelID: message.ChannelID,
		MessageID: message.ID,
		Text:      message.Text,
	})
	if err != nil {
		return err
	}
	if strings.TrimSpace(reply) == "" {
		return nil
	}
	return p.Platform.Publish(ctx, ownerID, message.ChannelID, message.RootID, reply, message.ID)
}
