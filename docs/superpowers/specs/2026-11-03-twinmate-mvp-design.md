# Twinmate MVP design

## Intent

Let Mattermost users attach their own HTTP-based digital twin. A twin may reply **in the existing one-to-one DM**, without creating a group conversation. Every generated message must be visibly attributed to an AI twin, including on native mobile clients.

## MVP goals

- Self-service endpoint and bearer-token configuration for each user.
- Explicit per-conversation enable/disable through `/twinmate on|off` in a DM. Default: off.
- After a human message in that DM, call the **recipient's** configured external agent, not the sender's.
- Agents receive only the triggering message and identifiers, not DM history.
- Agent may return a reply or decline (`{"reply":""}`); the plugin posts replies in the original channel as their owner with machine-readable provenance and an in-body AI label.
- Ignore plugin-generated posts to avoid agent-to-agent loops.
- Web/desktop badge and configuration page; mobile renders the plain message with the in-body AI label.
- Secure defaults: HTTPS endpoints, no redirects, finite request/response size and timeout, per-user secret never returned from GET, admin opt-in for private endpoint IPs.

## Explicit non-goals in MVP

- Local agents behind NAT connecting outbound; a reachable HTTPS endpoint is required.
- Background retries, durable queue, multi-node exactly-once delivery, attachments, RAG/history, group channels, arbitrary autonomous posting, administration of external runtimes.
- Custom native iOS/Android rendering; that requires a mobile app fork.

## Architecture

The Go server plugin listens to `MessageHasBeenPosted`, resolves the one-to-one recipient, checks its connection and DM opt-in, sends an event to that recipient's external HTTP agent asynchronously, and publishes a labeled response through `CreatePost`. Data lives in plugin KV. A small webapp bundle provides a settings page and optional post-header badge. The mobile-safe label is also included in the stored Markdown.

Agent request: `POST` with JSON `{event_id,owner_id,sender_id,channel_id,message_id,text}` and optional `Authorization: Bearer ...`; response JSON `{reply: string}`. The agent never receives a Mattermost access token.

## User-visible decisions

- Name: **Twinmate**, from the connected GitHub repository.
- A generated response is authored in Mattermost's data model by its human owner to keep the original two-person DM, while explicit provenance is recorded in post props. This is **not** a separate Mattermost bot identity.
- Label in every client: `🤖 **Twinmate · AI-двойник**` at the beginning of the message. Web clients may add a compact badge.
- Privacy: enabling an agent requires both personal connection opt-in and per-DM activation by the owner; participants should understand that messages sent to that owner may be shared with the configured external runtime. Mutual participant consent is a future policy decision, not silently claimed as delivered.

## Deployment

Target Mattermost v11+ for the webapp post-header extension. Build a standard Go server executable + Webpack React webapp bundle, package as a Mattermost `.tar.gz` plugin. GitHub Actions runs tests and produces build artifacts.
