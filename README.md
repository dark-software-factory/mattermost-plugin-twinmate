# Twinmate

Personal digital twins for [Mattermost](https://mattermost.com). Each person connects their own HTTPS AI agent. The agent may answer new messages **in the existing one-to-one direct message** as that person, with a visible AI attribution.

**Status:** v0.1.0 preview. Not production-hardened; test on a non-production Mattermost instance first.

## What works

- Personal HTTPS agent endpoint and optional bearer token configured in **Main menu → Twinmate · My digital twin**.
- Explicit per-DM opt-in: `/twinmate on`, `/twinmate off`, `/twinmate status` (also available in native mobile apps).
- Only the receiving participant's enabled agent is called after a human sends a new DM message. Posts from Mattermost bot users and system integrations are ignored. No group channels, background polling, or DM history extraction.
- Agent responses appear in the **same** DM, attributed to the owner, with persisted `post.props.twinmate` metadata and a mobile-visible `🤖 Twinmate · AI-двойник` prefix. Web/desktop has an additional `AI-двойник` header badge when a header is shown; this badge is verified against server-owned provenance, not trusted from client-supplied post props.
- Agents get a scoped event; they do **not** receive a Mattermost personal access token.

## Requirements

- Mattermost Server **v11+**, with plugin support enabled.
- A reachable HTTPS endpoint implementing the contract below. The server must be able to call it directly.
- Admin enables installation/upload of plugins. A Linux amd64 package is built by CI.

## Build and install

```bash
go mod tidy
npm ci --prefix webapp
make test
make bundle
```

The resulting `dist/com.darksoftwarefactory.twinmate-0.1.0.tar.gz` can be uploaded in the Mattermost System Console's plugin management page. The GitHub Actions workflow also publishes this archive as a build artifact when it succeeds.

1. Open the Mattermost main menu, then **Twinmate · My digital twin**.
2. Enter your agent's HTTPS endpoint and optional bearer token, enable the connection, and save.
3. Open an existing personal DM and type `/twinmate on` to allow that particular conversation to be processed. Use `/twinmate off` to pause.
4. When the other participant writes a message, Twinmate sends the event to your endpoint. The agent can respond or decline.

When changing the agent's hostname, the previously saved bearer token is cleared unless you explicitly enter a replacement. When modifying only the path on the same host, the token is preserved. The settings form also lets you remove the current token. Disabling a twin prevents a pending agent reply from being published once the external call finishes, but it cannot retract an already-transmitted event.

## Agent API

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
- AI post badges are shown only after checking a Mattermost-authenticated, DM-member-restricted provenance endpoint. A user can type a similar emoji label manually, but cannot obtain a verified Twinmate badge through post props alone.
- The plugin currently publishes posts with the owner's Mattermost user ID to preserve two-person DM membership. This is **not a separate bot user**. The AI label and custom post props disclose the origin; native mobile clients display the in-body label, not a custom native badge.
- Delivery is best-effort in memory with a 32-worker limit: a restart, an HTTP timeout or overload can lose an event. There is no outbox, retry queue or cross-node deduplication yet.

## Known gaps

- No outbound reverse tunnel for personal agents behind NAT.
- No durable queue, attachments, group chat or full conversation context.
- No mutual participant-consent flow or external data retention governance yet.
- No custom mobile client code. Native mobile clients cannot load webapp plugin React components.

## Development

The [design](docs/superpowers/specs/2026-11-03-twinmate-mvp-design.md) and [implementation plan](docs/superpowers/plans/2026-11-03-twinmate-mvp.md) capture current decisions and tradeoffs.
