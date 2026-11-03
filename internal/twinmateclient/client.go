// Package twinmateclient is the HTTPS transport used by local MCP agents.
// It never needs a Mattermost personal access token.
package twinmateclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const agentAPIPath = "/plugins/com.darksoftwarefactory.twinmate/api/agent"

// JSON escaping can expand a 15KiB message substantially. A page of 50
// posts may exceed 256KiB even though every message is valid.
const maxAgentResponseBytes = 8 * 1024 * 1024

type ContextPost struct {
	ID string `json:"id"`
	ChannelID string `json:"channel_id"`
	UserID string `json:"user_id"`
	RootID string `json:"root_id,omitempty"`
	Message string `json:"message"`
	CreateAt int64 `json:"create_at"`
}
type MessagePage struct {
	ChannelID string `json:"channel_id"`
	AnchorPostID string `json:"anchor_post_id"`
	Posts []ContextPost `json:"posts"`
	HasMore bool `json:"has_more"`
	NextCursor string `json:"next_cursor,omitempty"`
}
type SendRequest struct {
	ChannelID string `json:"channel_id,omitempty"`
	ReplyToPostID string `json:"reply_to_post_id,omitempty"`
	Message string `json:"message"`
}
type SendResult struct { PostID string `json:"post_id"` }

type Client struct {
	baseURL string
	token string
	httpClient *http.Client
}

func New(rawBaseURL, token string, client *http.Client) (*Client,error) {
	u,err:=url.Parse(rawBaseURL)
	if err!=nil || u.Scheme!="https" || u.Host=="" || u.User!=nil || u.Fragment!="" || u.RawQuery!="" {
		return nil,errors.New("TWINMATE_MATTERMOST_URL must be an HTTPS Mattermost origin")
	}
	if token=="" {return nil,errors.New("TWINMATE_AGENT_TOKEN is required")}
	if client==nil {client=&http.Client{Timeout:15*time.Second}}
	return &Client{baseURL:strings.TrimRight(u.String(),"/"),token:token,httpClient:client},nil
}

func (c *Client) request(ctx context.Context, method, path string, query url.Values, input any, output any) error {
	var body io.Reader
	if input!=nil {
		data,err:=json.Marshal(input)
		if err!=nil {return err}
		body=bytes.NewReader(data)
	}
	endpoint:=c.baseURL+agentAPIPath+path
	if len(query)!=0 {endpoint+="?"+query.Encode()}
	req,err:=http.NewRequestWithContext(ctx,method,endpoint,body)
	if err!=nil {return err}
	req.Header.Set("X-Twinmate-Token",c.token)
	if input!=nil {req.Header.Set("Content-Type","application/json")}
	noRedirect:=*c.httpClient
	noRedirect.CheckRedirect=func(*http.Request,[]*http.Request)error{return errors.New("Twinmate redirect rejected")}
	resp,err:=noRedirect.Do(req)
	if err!=nil {return err}
	defer resp.Body.Close()
	if resp.StatusCode<200 || resp.StatusCode>=300 {return fmt.Errorf("Twinmate HTTP %d",resp.StatusCode)}
	data,err:=io.ReadAll(io.LimitReader(resp.Body,maxAgentResponseBytes+1))
	if err!=nil {return err}
	if len(data)>maxAgentResponseBytes {return errors.New("Twinmate response exceeds 8 MiB")}
	if err:=json.Unmarshal(data,output);err!=nil {return fmt.Errorf("invalid Twinmate response: %w",err)}
	return nil
}

func (c *Client) GetPost(ctx context.Context, reference string) (ContextPost,error) {
	var result ContextPost
	if strings.TrimSpace(reference)=="" {return result,errors.New("Mattermost post ID or permalink required")}
	query:=url.Values{"reference":{reference}}
	err:=c.request(ctx,http.MethodGet,"/posts/get",query,nil,&result)
	return result,err
}

func (c *Client) ListPosts(ctx context.Context, reference, direction, scope string, limit int) (MessagePage,error) {
	var result MessagePage
	if strings.TrimSpace(reference)=="" {return result,errors.New("Mattermost post ID or permalink required")}
	if direction!="before" && direction!="after" {return result,errors.New("direction must be before or after")}
	if scope=="" {scope="channel"}
	if scope!="channel" && scope!="thread" {return result,errors.New("scope must be channel or thread")}
	if limit==0 {limit=20}
	if limit<1 || limit>50 {return result,errors.New("limit must be between 1 and 50")}
	query:=url.Values{"reference":{reference},"direction":{direction},"scope":{scope},"limit":{fmt.Sprint(limit)}}
	err:=c.request(ctx,http.MethodGet,"/posts/list",query,nil,&result)
	return result,err
}

func (c *Client) SendMessage(ctx context.Context, input SendRequest) (SendResult,error) {
	var result SendResult
	if strings.TrimSpace(input.Message)=="" || (input.ChannelID=="" && input.ReplyToPostID=="") {
		return result,errors.New("message and channel_id or reply_to_post_id are required")
	}
	err:=c.request(ctx,http.MethodPost,"/messages",nil,input,&result)
	return result,err
}
