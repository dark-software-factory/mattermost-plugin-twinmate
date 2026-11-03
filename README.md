# Twinmate

Personal digital twins for [Mattermost](https://mattermost.com). A developer's local OpenClaw or another agent can retrieve a linked discussion and publish work results as the developer's visibly marked AI twin. The plugin also supports optional automatic responses to new one-to-one DMs.

**Status:** v0.1.0 preview. Not production-hardened; test on a non-production Mattermost instance first.

## What works

- Agent-initiated API: issue a personal Twinmate key, read any accessible message by permalink, explicitly page messages before/after it within a channel or thread, and write a new message or reply as the owner. Local agents need only outbound HTTPS, without a Mattermost PAT.
- Local Go MCP adapter: stdio tools `twinmate_get_post`, `twinmate_list_posts`, and `twinmate_send_message`, buildable for macOS arm64/amd64 and Linux amd64.
- Personal HTTPS agent endpoint and optional bearer token configured in **Main menu → Twinmate · My digital twin**.
- Explicit per-DM opt-in: `/twinmate on`, `/twinmate off`, `/twinmate status` (also available in native mobile apps).
- Only the receiving participant's enabled agent is called after a human sends a new DM message. Posts from Mattermost bot users and system integrations are ignored. No group channels, background polling, or DM history extraction.
- Agent responses appear in the **same** DM, attributed to the owner, with persisted `post.props.twinmate` metadata and a mobile-visible, locale-neutral `🤖 Twinmate · AI-generated` prefix. Web/desktop additionally shows a server-verified localized badge (e.g. `AI twin` in English, `ИИ-двойник` in Russian, `KI-Zwilling` in German). The badge is verified against server-owned provenance, not trusted from client-supplied post props. The settings UI supports the main Mattermost locales with English fallback.
- Agents get a scoped event; they do **not** receive a Mattermost personal access token.

## Requirements

- Mattermost Server **v11+**, with plugin support enabled.
- A reachable HTTPS endpoint is necessary **only for optional automatic DM replies**; the agent-initiated API requires no publicly reachable agent machine.
- Admin enables installation/upload of plugins. A Linux amd64 package is built by CI.

## Build and install

```bash
go mod tidy
npm ci --prefix webapp
make test
make bundle
make build-mcp
```

The resulting `dist/com.darksoftwarefactory.twinmate-0.1.0.tar.gz` can be uploaded in the Mattermost System Console's plugin management page. The GitHub Actions workflow publishes this plugin archive plus separate local MCP binaries as build artifacts when it succeeds.

## Connect OpenClaw or another local MCP client

1. Install the Twinmate plugin. Open **Main menu → Twinmate · My digital twin** and click **Generate agent key**. Copy it immediately: the server stores only its hash. Rotating/revoking the key invalidates the old one.
2. Build the stdio MCP binary using `make build-mcp` or download the `twinmate-mcp-binaries` build artifact. For Apple Silicon use `dist/twinmate-mcp-darwin-arm64`; for Intel macOS use `dist/twinmate-mcp-darwin-amd64`.
3. Point your MCP host at the binary and pass the Mattermost URL and the new Twinmate key. The config format below is common for stdio MCP hosts; adapt the keys to your host if necessary:

```json
{
  "mcpServers": {
    "twinmate": {
      "command": "/path/to/twinmate-mcp-darwin-arm64",
      "env": {
        "TWINMATE_MATTERMOST_URL": "https://mattermost.example.com",
        "TWINMATE_AGENT_TOKEN": "<your-one-time-Twinmate-key>"
      }
    }
  }
}
```

The local MCP adapter exposes three tools:

- `twinmate_get_post(reference)` — fetch one post by ID or Mattermost permalink, including its channel/thread identifiers.
- `twinmate_list_posts(reference, direction, scope?, limit?)` — fetch messages `before` or `after` that post in `channel` (default) or `thread` scope, 1–50 at a time (default 20). Responses include `has_more` and `next_cursor`; pass the latter as the next `reference` to continue. The agent chooses how much context to request.
- `twinmate_send_message(message, channel_id?, reply_to_post_id?)` — post a new message or reply as the token owner with verified AI provenance.

No `/twinmate on` command is necessary for **agent-initiated** calls; that command controls *automatic* DM responses only.

### Direct HTTPS API

The MCP adapter calls the same API any developer tool can use. Supply `X-Twinmate-Token`, not a Mattermost personal access token:

```bash
curl -H "X-Twinmate-Token: $TWINMATE_AGENT_TOKEN" \
  'https://mattermost.example.com/plugins/com.darksoftwarefactory.twinmate/api/agent/posts/get?reference=<post-id>'
```

Fetch the seven messages preceding a post in its thread (set `scope=channel` to read nearby channel messages instead):

```bash
curl -H "X-Twinmate-Token: $TWINMATE_AGENT_TOKEN" \
  'https://mattermost.example.com/plugins/com.darksoftwarefactory.twinmate/api/agent/posts/list?reference=<post-id>&direction=before&scope=thread&limit=7'
```

Post a result under a thread:

```bash
curl -X POST -H 'Content-Type: application/json' \
  -H "X-Twinmate-Token: $TWINMATE_AGENT_TOKEN" \
  -d '{"reply_to_post_id":"<post-id>","message":"Investigation complete: the race was fixed."}' \
  'https://mattermost.example.com/plugins/com.darksoftwarefactory.twinmate/api/agent/messages'
```

Both read endpoints accept `reference` as either a Mattermost permalink or post ID and never fetch arbitrary URLs. Responses contain chronological, non-deleted text posts, never post props or attachment contents. The plugin checks the current owner's user status and channel permissions for every request; publishing also checks channel membership and rejects archived channels. **The personal key grants read-by-link and send-as-owner capabilities:** keep it secret, and revoke it if your machine is compromised.

## Optional automatic DM responses

1. Open the Mattermost main menu, then **Twinmate · My digital twin**.
2. Enter your agent's HTTPS endpoint and optional bearer token, enable the connection, and save.
3. Open an existing personal DM and type `/twinmate on` to allow that particular conversation to be processed. Use `/twinmate off` to pause.
4. When the other participant writes a message, Twinmate sends the event to your endpoint. The agent can respond or decline.

When changing the agent's hostname, the previously saved bearer token is cleared unless you explicitly enter a replacement. When modifying only the path on the same host, the token is preserved. The settings form also lets you remove the current token. Disabling a twin prevents a pending agent reply from being published once the external call finishes, but it cannot retract an already-transmitted event.

## External push-agent callback contract

Twinmate sends `POST https://your-agent.example/events` with `Content-Type: application/json`, an `X-Twinmate-Event-ID` header, and `Authorization: Bearer <configured-token>` if provided.

```json
{
  "event_id": "source-post-id:recipient-user-id",
  "owner_id": "recipient-user-id",
  "sender_id": "sender-user-id",
  "channel_id": "existing-dm-channel-id",
  "message_id": "source-post-id",
  "text": "When do we finish the migration?"
}
```

The agent returns HTTP 2xx and JSON:

```json
{"reply": "We expect to finish by the end of this sprint."}
```

Return `{"reply":""}` when the agent chooses not to speak. Calls time out after 25 seconds; response bodies must be at most 64 KiB and posted messages at most 15,000 UTF-8 bytes. Redirects are never followed.

## Security and privacy

- All connections and per-DM permissions start disabled. **Enabling a twin means the external agent can receive new messages from that DM.** Communicate this to the other participant and obtain any consent your organization requires.
- Private, loopback and link-local destinations are blocked by default, including DNS-resolved IPs at connection time; an administrator may enable private-network endpoints in plugin settings. This carries server-side request forgery risk and is for trusted corporate networks only.
- Connection tokens are stored in the Mattermost plugin KV store. They are withheld from the browser's GET endpoint, but **they are not encrypted independently by this plugin**. Use a dedicated rotatable token and protect database backups.
- By contrast, the inbound **Twinmate agent API key** is random, stored only as a SHA-256 digest, never returned on status GET, and can be rotated/revoked by its Mattermost owner. It authorizes reading linked messages and paginating within authorized conversations, and posting as that user; keep the key out of logs and untrusted configuration files.
- AI post badges are shown only after checking a Mattermost-authenticated, DM-member-restricted provenance endpoint. A user can type a similar emoji label manually, but cannot obtain a verified Twinmate badge through post props alone.
- The plugin currently publishes posts with the owner's Mattermost user ID to preserve two-person DM membership. This is **not a separate bot user**. The AI label and custom post props disclose the origin; native mobile clients display the in-body label, not a custom native badge.
- Delivery is best-effort in memory with a 32-worker limit: a restart, an HTTP timeout or overload can lose an event. There is no outbox, retry queue or cross-node deduplication yet.

## Known gaps

- No outbound reverse tunnel for **automatic DM push** agents behind NAT. Agent-initiated API/MCP calls already work from local machines without incoming connections.
- No durable auto-reply queue, attachment downloads or automatic group-chat replies. The agent-initiated API can post to authorized channels and fetch channel/thread pages around a reference post.
- No mutual participant-consent flow or external data retention governance yet.
- No custom mobile client code. Native mobile clients cannot load webapp plugin React components.
- Mattermost hides post headers for consecutive posts by the same author, so its `registerPostHeaderComponent` badge is also suppressed on those posts. The in-body `🤖 Twinmate · AI-generated` marker remains visible in all clients, including those grouped posts and native mobile clients.
- Thread-scoped paging currently uses Mattermost's unpaginated plugin `GetPostThread` API internally; extremely long threads can be expensive to read, despite the 50-post response cap. This needs a bounded server-side pagination strategy before production rollout.
- The agent-initiated API deliberately has no global search, attachment downloads, per-token fine-grained scopes or multiple simultaneous keys yet. Each read page is capped at 50 posts, but the agent may request more pages using the returned cursor.

## Development

The [original DM design](docs/superpowers/specs/2026-11-03-twinmate-mvp-design.md) and [agent-initiated API design](docs/superpowers/specs/2026-11-03-twinmate-agent-api-design.md), with their implementation plans, capture current decisions and tradeoffs.

### Full integration tests

The `e2e/` Playwright suite follows the disposable Mattermost + PostgreSQL Testcontainers pattern used in the [Mattermost Docs plugin](https://github.com/mattermost/mattermost-plugin-docs). It builds and uploads the **actual** plugin archive to a newly created Mattermost 11 server, creates four real users, and runs seven browser/API/MCP/automatic-DM scenarios:

- Startup and plugin activation through Mattermost's official API.
- Per-user key issuance, rotation/revocation, private-channel isolation, real permission checks, chronological paging, thread scope, AI post ownership and server-verified provenance.
- A compiled Go stdio MCP process calling the real server through a disposable, certificate-validated HTTPS proxy.
- Chromium rendering of the AI badge in English and Russian, plus the settings dialog.
- Actual automatic-DM delivery over HTTPS to a disposable test agent container, validating opt-in/opt-out, event authentication and generated reply provenance.

With a local Docker daemon, Go, Node.js 24, Chromium dependencies and OpenSSL:

```bash
make test-e2e
```

The separate GitHub Actions **Mattermost integration E2E** workflow runs the same suite on pull requests and `main`, preserving Playwright traces, JUnit output and Mattermost logs on failure. The suite always creates its own server and refuses the `MM_E2E_USE_EXISTING_SERVER` escape hatch; it never seeds a team Mattermost instance. No mock Mattermost service is involved.

