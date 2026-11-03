# Twinmate Agent API and MCP Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Enable personal local agents to retrieve a Mattermost thread by URL and publish AI-labeled posts through a scoped Twinmate API and Go MCP adapter.

**Architecture:** Server-side Go token issuance/auth verifies a self-owned, revocable credential; a narrow REST bridge delegates channel permission checks and post writes to Mattermost. A local Go stdio MCP process wraps the REST client without a Mattermost PAT.

**Tech Stack:** Mattermost server/public Go v0.6.0, Go 1.26.10, modelcontextprotocol/go-sdk v1.8.0, React webapp and GitHub Actions.

**Spec:** `docs/superpowers/specs/2026-11-03-twinmate-agent-api-design.md`

## Global Constraints

- Preserve current plugin ID `com.darksoftwarefactory.twinmate`, existing auto-DM behavior, mobile label, and server-verified web provenance.
- Only one active Twinmate API token per owner, hash-at-rest, shown once; no Mattermost PAT passed to agents.
- Context is thread-scoped and limited to at most 50 posts; only owner-authorized channels.
- Token API authenticated by verified Mattermost session, agent API by `X-Twinmate-Token`.
- Feature stays on existing PR branch and is verified by GitHub CI until a local Go runtime is available.

## Review Focus

- A token for user A requests a post/channel visible only to B: must return 403 and no content.
- An attacker guesses `owner_id` or supplies another user's ID in POST body: API must reject/ignore it and never impersonate.
- A revoked or rotated token attempts another read/write: must return 401.
- A reply target is in a different channel from `channel_id`: must reject, no post.
- A malicious permalink contains an external host: parser must not fetch it or accept it for a different Mattermost instance.

### Task 1: Personal token issuance and authentication

**Files:** Create `server/access_tokens.go`, `server/access_tokens_test.go`; Modify `server/plugin.go` to route `/api/agent/token`.

**Interfaces:** Produce `TokenStore` (Get/Put by key), `TokenManager` with `Rotate(ownerID) (string,error)`, `Revoke(ownerID) error`, `Authenticate(raw) (ownerID,error)`, `HasToken(ownerID) (bool,error)`; HTTP handler for Mattermost-authenticated GET/POST/DELETE.

- [ ] Write tests for no-session 401, token shown once, hash-not-secret storage, authenticate, rotation revokes old token, and revoke.
- [ ] Push test-only changes; GitHub CI must fail because implementation is missing (RED).
- [ ] Implement token manager with cryptographic randomness and constant-time verification, then session-only HTTP handler.
- [ ] Push and verify relevant Go tests and `go vet` pass (GREEN).

### Task 2: Context and write REST agent API

**Files:** Create `server/agent_api.go`, `server/agent_api_test.go`; Modify `server/plugin.go` to route `/api/agent/context` and `/api/agent/messages`, and reuse provenance post creation.

**Interfaces:** `Bridge` implements `Context(ownerID, postID string) (ContextResult,error)` and `Send(ownerID, channelID, replyID, text string) (postID string,error)`; `AgentHandler{Tokens,Bridge,SiteURL}` verifies token and request shape.

- [ ] Write tests for token auth, URL/post ID extraction, read permission, thread ordering/limit, reply-channel mismatch, no impersonation, and successful AI-marked write.
- [ ] Push test changes and verify CI fails for the expected undefined symbols (RED).
- [ ] Implement a narrow REST handler and permission-checked Mattermost bridge using `GetPost`, `GetPostThread`, `GetChannelMember` and `HasPermissionToChannel`.
- [ ] Push and verify Go tests/vet pass (GREEN).

### Task 3: Local MCP client, settings UI and instructions

**Files:** Create `cmd/twinmate-mcp/main.go`, `internal/twinmateclient/client.go`, `internal/twinmateclient/client_test.go`; Modify `go.mod`, `webapp/src/index.js`, `README.md`, `Makefile`, `.github/workflows/ci.yml` as needed.

**Interfaces:** `twinmateclient.Client{BaseURL,Token,HTTP}` exposes `GetContext(ctx,input)`, `SendMessage(ctx,request)`, stdio MCP tool schemas `twinmate_get_context` and `twinmate_send_message` use this client. User settings display an issue/revoke token control.

- [ ] Write client tests for correctly authenticated context/read and write requests, 4xx propagation and refused redirects.
- [ ] Push test-only changes and verify they fail (RED).
- [ ] Implement Go MCP SDK stdio CLI and secure HTTPS client; add settings button for token issue/revoke, show secret only at creation.
- [ ] Update README with REST/MCP usage and examples; verify `go test ./...`, `go vet ./...`, `npm ci/test/build`, and packaged plugin via GitHub CI (GREEN).

### Task 4: Final review and installation artifact

**Files:** Adjust code only for Critical/Important review issues, with reproducing tests.

- [ ] Request a fresh read-only code review covering authentication, Mattermost permission enforcement, URL parsing, AI provenance and MCP stdout safety.
- [ ] Address Critical/Important findings with RED→GREEN tests, run full CI, and keep PR draft until live Mattermost 11+ smoke test.
- [ ] Check GitHub Actions archive and document exact install/test steps.
