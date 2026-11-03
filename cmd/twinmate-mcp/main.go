package main

import (
	"context"
	"log"
	"os"

	"github.com/dark-software-factory/mattermost-plugin-twinmate/internal/twinmateclient"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type GetContextInput struct {
	Reference string `json:"reference" jsonschema:"Mattermost message permalink or post ID"`
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
	mcp.AddTool(server,&mcp.Tool{Name:"twinmate_get_context",Description:"Read a Mattermost thread by permalink or post ID using the current user's Twinmate permissions"},
		func(ctx context.Context,_ *mcp.CallToolRequest,in GetContextInput)(*mcp.CallToolResult,twinmateclient.ContextResult,error) {
			result,err:=client.GetContext(ctx,in.Reference)
			return nil,result,err
		})
	mcp.AddTool(server,&mcp.Tool{Name:"twinmate_send_message",Description:"Publish a message or thread reply as the current user's visibly marked AI twin"},
		func(ctx context.Context,_ *mcp.CallToolRequest,in SendMessageInput)(*mcp.CallToolResult,twinmateclient.SendResult,error) {
			result,err:=client.SendMessage(ctx,twinmateclient.SendRequest{ChannelID:in.ChannelID,ReplyToPostID:in.ReplyToPostID,Message:in.Message})
			return nil,result,err
		})
	if err:=server.Run(context.Background(),&mcp.StdioTransport{});err!=nil {log.Fatal(err)}
}
