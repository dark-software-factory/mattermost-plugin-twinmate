## Boundaries

- The Go plugin owns auth, permissions, egress and signed provenance decisions.
- The webapp only selects key scopes and shows status; a client cannot promote
  its own key after rotation.
- The local Go MCP adapter continues using the same REST API; an operation
  outside the selected key scope returns HTTP 403.

## Decisions

1. **Callback egress:** `AllowPrivateEndpoints` is no longer blanket permission.
   `AllowedPrivateHosts` holds exact lowercase `host:port` pairs; no wildcards.
   Recheck hostname and resolved addresses for each dial; never enable proxies
   or follow redirects. Public HTTPS targets need no allowlist. Admin changes
   take effect for future requests.
2. **Personal credentials:** a single active token per owner keeps a one-way
   digest plus immutable `allow_send` and `channel_ids` scopes. Empty channel
   list means all channels that Mattermost itself permits. New keys default
   to read-only. Legacy KV records predate scopes and retain compatibility
   until the owner rotates; GET status exposes scope metadata, never secret.
3. **DM disclosure:** on the first opt-in from an owner, publish a public
   informational post in the DM. Repeated `/on` does not spam. Opt-in occurs
   only if notification succeeds. The notice contains no secret or endpoint.
4. **Attribution:** reject editing a server-recorded generated post to remove
   its marker. When deciding which post to forward, trust server-owned
   provenance rather than treating arbitrary matching user text as proof.
5. **History:** paginate via `GetPostsBefore/After` with a hard cap of five
   pages of 100 posts. Never call `GetPostThread` (which loads entire threads).
   Return HTTP 422 if the requested thread page cannot be determined within
   the scan budget, rather than falsely claiming end of history. The client
   may switch to channel paging.
6. **Reply race:** the server verifies owner is active, member and allowed to
   post immediately before automatic delivery, independently of the checks
   already made for agent-initiated sends.

## Verification

Go unit tests cover each denied path, edge and credential transition. Existing
Playwright/Testcontainers E2E must cover send and read with explicitly scoped
tokens, DM disclosure, and Mattermost 11.0/11.11. CI runs Go vet/tests,
webapp tests and the real Mattermost E2E suite.
