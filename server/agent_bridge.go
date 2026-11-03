package main

import (
	"errors"

	"github.com/mattermost/mattermost/server/public/model"
)

type MattermostAgentAPI interface {
	GetPost(id string) (*model.Post,*model.AppError)
	GetPostThread(id string) (*model.PostList,*model.AppError)
	GetChannel(id string) (*model.Channel,*model.AppError)
	GetChannelMember(channelID,userID string) (*model.ChannelMember,*model.AppError)
	HasPermissionToChannel(userID,channelID string,permission *model.Permission) bool
}

type MattermostBridge struct {
	API MattermostAgentAPI
	Publisher func(owner,channel,root,message string)(string,error)
}

func (b MattermostBridge) canAccess(owner,channelID string,permission *model.Permission) bool {
	channel,err:=b.API.GetChannel(channelID)
	if err!=nil || channel==nil {return false}
	if !b.API.HasPermissionToChannel(owner,channelID,permission) {return false}
	// Mattermost's server-side plugin API operates with plugin authority,
	// not the caller's channel visibility. Private/DM channels must always
	// verify the actual membership before returning private content.
	if channel.Type!=model.ChannelTypeOpen || permission==model.PermissionCreatePost {
		member,err:=b.API.GetChannelMember(channelID,owner)
		if err!=nil || member==nil {return false}
	}
	return true
}

func (b MattermostBridge) Context(owner,postID string) (ContextResult,error) {
	post,err:=b.API.GetPost(postID)
	if err!=nil || post==nil || post.DeleteAt!=0 {return ContextResult{},ErrAgentNotFound}
	if !b.canAccess(owner,post.ChannelId,model.PermissionReadChannel) {return ContextResult{},ErrAgentForbidden}
	list,err:=b.API.GetPostThread(postID)
	if err!=nil || list==nil {return ContextResult{},ErrAgentNotFound}
	result:=ContextResult{ChannelID:post.ChannelId,TargetPostID:postID,Posts:sanitizeThread(list,postID,50)}
	return result,nil
}

func (b MattermostBridge) Send(owner,channelID,replyID,text string) (string,error) {
	root:=""
	if replyID!="" {
		reply,err:=b.API.GetPost(replyID)
		if err!=nil || reply==nil || reply.DeleteAt!=0 {return "",ErrAgentNotFound}
		if channelID!="" && channelID!=reply.ChannelId {return "",ErrAgentBadInput}
		channelID=reply.ChannelId
		root=reply.RootId
		if root=="" {root=reply.Id}
	}
	if channelID=="" {return "",ErrAgentBadInput}
	if !b.canAccess(owner,channelID,model.PermissionCreatePost) {return "",ErrAgentForbidden}
	if b.Publisher==nil {return "",errors.New("no Twinmate publisher configured")}
	return b.Publisher(owner,channelID,root,text)
}
