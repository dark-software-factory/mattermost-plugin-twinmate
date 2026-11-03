package main

import (
	"context"
	"log"
	"os"

	"github.com/dark-software-factory/mattermost-plugin-twinmate/internal/twinmateclient"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type GetPostInput struct {
	Reference string `json:"reference" jsonschema:"Mattermost message permalink or post ID"`
}
type ListPostsInput struct {
	Reference string `json:"reference" jsonschema:"Mattermost message permalink or post ID used as the pagination anchor"`
	Direction string `json:"direction" jsonschema:"before or after the reference post"`
	Scope string `json:"scope,omitempty" jsonschema:"channel or thread; defaults to channel"`
	Limit int `json:"limit,omitempty" jsonschema:"Number of messages to fetch from 1 to 50; defaults to 20"`
}
type SendMessageInput struct {
	ChannelID string `json:"channel_id,omitempty" jsonschema:"Mattermost channel ID; optional if replying to a post"`
	ReplyToPostID string `json:"reply_to_post_id,omitempty" jsonschema:"Reply target post ID; optional for a new message"`
	Message string `json:"message" jsonschema:"Work result text to publish as the user's clearly labeled AI twin"`
}

func main() {
	client,err:=twinmateclient.New(os.Getenv("TWINMATE_MATTERMOST_URL"),os.Getenv("TWINMATE_AGENT_TOKEN"),nil)
	if err!=nil {log.Fatal(err)} // log writes to stderr, never stdout
	server:=mcp.NewServer(&mcp.Implementation{Name:"twinmate",Version:"0.1.0"},nil)
	mcp.AddTool(server,&mcp.Tool{Name:"twinmate_get_post",Description:"Read one Mattermost message by permalink or post ID using the user's own channel permissions"},
		func(ctx context.Context,_ *mcp.CallToolRequest,in GetPostInput)(*mcp.CallToolResult,twinmateclient.ContextPost,error) {
			result,err:=client.GetPost(ctx,in.Reference)
			return nil,result,err
		})
	mcp.AddTool(server,&mcp.Tool{Name:"twinmate_list_posts",Description:"Read messages before or after any post, with channel/thread scope and cursor pagination controlled by the agent"},
		func(ctx context.Context,_ *mcp.CallToolRequest,in ListPostsInput)(*mcp.CallToolResult,twinmateclient.MessagePage,error) {
			result,err:=client.ListPosts(ctx,in.Reference,in.Direction,in.Scope,in.Limit)
			return nil,result,err
		})
	mcp.AddTool(server,&mcp.Tool{Name:"twinmate_send_message",Description:"Publish a message or thread reply as the current user's visibly marked AI twin"},
		func(ctx context.Context,_ *mcp.CallToolRequest,in SendMessageInput)(*mcp.CallToolResult,twinmateclient.SendResult,error) {
			result,err:=client.SendMessage(ctx,twinmateclient.SendRequest{ChannelID:in.ChannelID,ReplyToPostID:in.ReplyToPostID,Message:in.Message})
			return nil,result,err
		})
	if err:=server.Run(context.Background(),&mcp.StdioTransport{});err!=nil {log.Fatal(err)}
}
