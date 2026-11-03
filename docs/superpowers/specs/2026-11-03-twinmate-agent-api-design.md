# Twinmate agent-initiated API and MCP design

## Intent

Let a developer's personal OpenClaw or any local/cloud coding agent read a Mattermost discussion by a message permalink and publish work results back into a channel or thread **as that developer's explicitly marked AI twin**. No incoming connection to the developer's machine is required.

## Approaches

- Give every agent a Mattermost PAT and let it use native APIs: fast, but broad credentials and no authoritative Twinmate origin marking.
- **Selected:** Give every developer a revocable Twinmate-specific token. The server plugin resolves owner identity, checks Mattermost permissions for every call, and creates AI-labeled posts. Provide a Go stdio MCP adapter over the same small HTTPS API.
- Outgoing webhooks / websocket pull: valuable for *autonomous reactions* but unnecessary for the two user-initiated workflows, and not the priority of this increment.

## MVP behavior

1. A signed-in Mattermost user opens Twinmate settings and generates/revokes a personal access credential, shown only once and stored as a hash, separate from any outbound agent endpoint secret.
2. An agent holding that credential calls `GET /api/agent/context?post_id=...` or `?url=<permalink>`. The plugin resolves the post, verifies that this user may read the channel, and returns the targeted thread (up to 50 non-deleted text posts, chronological), with channel/post IDs, authors, timestamps and the target ID. A permalink parser **never fetches the supplied URL over the network**. No global search or channel enumeration yet.
3. The agent calls `POST /api/agent/messages` with `channel_id` or `reply_to_post_id` plus `message`. The plugin verifies the user's channel membership and ability to create posts, derives `UserId` from the credential, and publishes through the existing `Twinmate` post provenance mechanism. The API cannot choose another owner.
4. Two stdio MCP tools, `twinmate_get_context` and `twinmate_send_message`, are implemented by a small Go CLI that calls the plugin over HTTPS. A local machine needs only outbound HTTPS to Mattermost. No Node/npm needed to run the MCP adapter.

## Boundaries and failure behavior

- Only authenticated Mattermost sessions can generate/revoke a token. Remote API requests use `X-Twinmate-Token` (not the Mattermost `Authorization` header) so that Mattermost does not confuse it with a PAT.
- Tokens have 256-bit random secret material, are verified using constant-time hash comparison, and may be rotated or revoked. Only one active token per owner in the first release. A token cannot authorize cross-user operations.
- Context lookup is limited to a direct link/post ID, and a maximum of 50 text messages. No attachments/downloads, global search, bulk exports, or ephemeral/deleted posts.
- Client-provided `channel_id` and `reply_to_post_id` must match. No write to channels outside the user's permissions. Incoming webhook, bot, or auto-reply paths retain their existing behavior.
- MCP must never print credentials or agent API data to stdout outside valid MCP protocol framing; errors use stderr.
- API errors are explicit JSON with appropriate 4xx/5xx responses. Sensitive responses are `Cache-Control: no-store`.
- Server requires Mattermost 11+, Go toolchain pinned in `go.mod`, and HTTPS for the external client in production.

## Product and security decisions

- `/twinmate on|off` controls **automatic DM replies only**. Manually invoked agent API writes are permitted in any channel the owner could normally write to, and always labeled AI.
- Read access covers the specific linked post/thread, even if the thread predates activation. It does not imply permission to search all of the owner's conversations.
- The API token grants both link-based reading and marked posting in v1. Separate scopes, confirmation workflows, multiple tokens, attachments and search can be added later.

## Success criteria

- OpenClaw on a developer laptop can resolve a Mattermost thread link and send a reply with no inbound network ports and no Mattermost PAT.
- A second user cannot read/post as the owner by swapping IDs, channels, links or a revoked token.
- Web and native mobile continue to show the persisted AI label; verified web badge uses server-owned post provenance.
