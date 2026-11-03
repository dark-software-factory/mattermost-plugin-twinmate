package main

import (
	"fmt"
	"strings"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin"
)

func parseTwinmateCommand(raw string) string {
	parts := strings.Fields(raw)
	if len(parts) == 1 && parts[0] == "/twinmate" {
		return "status"
	}
	if len(parts) != 2 || parts[0] != "/twinmate" {
		return "invalid"
	}
	switch parts[1] {
	case "on", "off", "status":
		return parts[1]
	default:
		return "invalid"
	}
}

func (p *Plugin) ExecuteCommand(_ *plugin.Context, args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	respond := func(message string) (*model.CommandResponse, *model.AppError) {
		return &model.CommandResponse{ResponseType: model.CommandResponseTypeEphemeral, Text: message}, nil
	}
	command := parseTwinmateCommand(args.Command)
	if command == "invalid" {
		return respond("Usage: `/twinmate on`, `/twinmate off` or `/twinmate status`.")
	}
	channel, appErr := p.API.GetChannel(args.ChannelId)
	if appErr != nil {
		return respond("Unable to resolve this channel.")
	}
	if channel.Type != model.ChannelTypeDirect {
		return respond("Twinmate currently supports one-to-one direct messages only.")
	}
	if _, appErr := p.API.GetChannelMember(channel.Id, args.UserId); appErr != nil {
		return respond("You are not a member of this conversation.")
	}
	if command == "status" {
		enabled, err := p.DMEnabled(nil, args.UserId, args.ChannelId)
		if err != nil {
			return respond("Unable to read Twinmate settings.")
		}
		if enabled {
			return respond("Your Twinmate is **enabled** in this DM. Use `/twinmate off` to pause it.")
		}
		return respond("Your Twinmate is **off** in this DM. Configure your agent in Twinmate settings, then use `/twinmate on`.")
	}
	if err := p.setDM(args.UserId, args.ChannelId, command == "on"); err != nil {
		return respond(fmt.Sprintf("Failed to update Twinmate: %v", err))
	}
	if command == "on" {
		return respond("Your Twinmate can now receive new messages from this DM and answer as your explicitly labeled AI twin. Use `/twinmate off` to stop it.")
	}
	return respond("Your Twinmate is now off in this DM.")
}
