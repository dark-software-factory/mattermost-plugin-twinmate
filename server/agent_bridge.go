package main

import (
	"errors"

	"github.com/mattermost/mattermost/server/public/model"
)

type MattermostAgentAPI interface {
	GetUser(userID string) (*model.User,*model.AppError)
	GetPost(id string) (*model.Post,*model.AppError)
	GetPostThread(id string) (*model.PostList,*model.AppError)
	GetPostsBefore(channelID,postID string,page,perPage int) (*model.PostList,*model.AppError)
	GetPostsAfter(channelID,postID string,page,perPage int) (*model.PostList,*model.AppError)
	GetChannel(id string) (*model.Channel,*model.AppError)
	GetChannelMember(channelID,userID string) (*model.ChannelMember,*model.AppError)
	HasPermissionToChannel(userID,channelID string,permission *model.Permission) bool
}

type MattermostBridge struct {
	API MattermostAgentAPI
	Publisher func(owner,channel,root,message string)(string,error)
}

func (b MattermostBridge) canAccess(owner,channelID string,permission *model.Permission) bool {
	user,err:=b.API.GetUser(owner)
	if err!=nil || user==nil || user.DeleteAt!=0 || user.IsBot {return false}
	channel,err:=b.API.GetChannel(channelID)
	if err!=nil || channel==nil {return false}
	if permission==model.PermissionCreatePost && channel.DeleteAt!=0 {return false}
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

func (b MattermostBridge) readablePost(owner,postID string) (*model.Post,error) {
	post,err:=b.API.GetPost(postID)
	if err!=nil || post==nil || post.DeleteAt!=0 {return nil,ErrAgentNotFound}
	if !b.canAccess(owner,post.ChannelId,model.PermissionReadChannel) {return nil,ErrAgentForbidden}
	return post,nil
}

func (b MattermostBridge) Get(owner,postID string) (ContextPost,error) {
	post,err:=b.readablePost(owner,postID)
	if err!=nil {return ContextPost{},err}
	return postToAgentPost(post),nil
}

// List returns one page relative to an anchor. The agent controls the direction,
// scope, and page size; no server-side context window is inferred.
func (b MattermostBridge) List(owner,postID,direction,scope string,limit int) (MessagePage,error) {
	if (direction!="before" && direction!="after") || (scope!="channel" && scope!="thread") || limit<1 || limit>50 {return MessagePage{},ErrAgentBadInput}
	anchor,err:=b.readablePost(owner,postID)
	if err!=nil {return MessagePage{},err}
	var list *model.PostList
	var appErr *model.AppError
	if scope=="thread" {
		root:=anchor.RootId
		if root=="" {root=anchor.Id}
		list,appErr=b.API.GetPostThread(root)
	} else if direction=="before" {
		list,appErr=b.API.GetPostsBefore(anchor.ChannelId,anchor.Id,0,limit+1)
	} else {
		list,appErr=b.API.GetPostsAfter(anchor.ChannelId,anchor.Id,0,limit+1)
	}
	if appErr!=nil || list==nil {return MessagePage{},ErrAgentNotFound}
	posts:=sanitizePosts(list)
	// Some server API responses include the anchor itself. Keep the ordering
	// and cursor semantics deterministic for both channel and thread scopes.
	filtered:=make([]ContextPost,0,len(posts))
	for _,post:=range posts {
		if post.ChannelID!=anchor.ChannelId {continue}
		if scope=="thread" {
			root:=anchor.RootId
			if root=="" {root=anchor.Id}
			if post.ID!=root && post.RootID!=root {continue}
		}
		if direction=="before" && (post.CreateAt<anchor.CreateAt || (post.CreateAt==anchor.CreateAt && post.ID<anchor.Id)) {
			filtered=append(filtered,post)
		} else if direction=="after" && (post.CreateAt>anchor.CreateAt || (post.CreateAt==anchor.CreateAt && post.ID>anchor.Id)) {
			filtered=append(filtered,post)
		}
	}
	page:=MessagePage{ChannelID:anchor.ChannelId,AnchorPostID:anchor.Id,Posts:filtered}
	if len(filtered)>limit {
		page.HasMore=true
		if direction=="before" {page.Posts=filtered[len(filtered)-limit:]} else {page.Posts=filtered[:limit]}
	}
	if len(page.Posts)!=0 {
		if direction=="before" {page.NextCursor=page.Posts[0].ID} else {page.NextCursor=page.Posts[len(page.Posts)-1].ID}
	}
	return page,nil
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
