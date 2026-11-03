# Twinmate real-Mattermost integration test design

## Intent

Catch failures hidden by mocked plugin APIs: install/activation incompatibility,
Mattermost auth and CSRF, real channel/DM permissions, post provenance, browser
registration, and stdio MCP interoperability. Running the test must not mutate
a developer's existing Mattermost instance.

## Patterns investigated

- **Mattermost Docs plugin:** Playwright global setup starts throwaway Mattermost
  and Postgres with Testcontainers, installs a freshly built plugin via
  `mmctl --local`, creates users/teams, and tears everything down.
- Mattermost GitHub and Boards plugins use Playwright for real web UI behavior.
- Running against an existing Mattermost is useful for opt-in debugging but
  must not be the default; a shared CI server would make tests non-hermetic.

**Selected:** One Playwright suite with Testcontainers for Mattermost 11 and
PostgreSQL, real Mattermost API calls for setup/assertions, and browser-driven
webapp assertions. Build the exact Linux plugin bundle and Linux MCP binary
from the current commit. No mocks of Mattermost or Twinmate in product E2E
tests. Temporary TLS proxy with a trusted test-only CA allows the Go MCP
binary's HTTPS-only client to connect to the local Mattermost server.

## Coverage

1. Start a clean Mattermost/Postgres; create admin and three people in a test
   team; upload/enable the built archive and verify plugin activation.
2. Exercise per-user Twinmate key issuance, status, rotation, revocation,
   anonymous rejection, and cross-user isolation over the real plugin HTTP API.
3. Create channel and DM messages with the Mattermost API. Verify permalink
   lookup, direction/scope/limit/cursor paging, forbidden private/DM reads,
   invalid references, and no content leak on failure.
4. Send normal channel posts and replies with an agent token; verify the
   owner identity, persisted `🤖 Twinmate · AI-generated` label, metadata,
   and provenance; reject owner spoofing and unauthorized writes.
5. Launch the actual compiled Go MCP process using stdio and call all three
   tools via JSON-RPC over a local trusted HTTPS proxy.
6. Verify web app registration and English/Russian localized AI badge and
   settings dialog in a Chromium browser.
7. Verify optional DM automation separately by configuring a reachable
   in-test HTTPS agent with an opt-in DM. A non-opted-in DM must not send
   agent events. If this requires a distinct container CA/SSRF setup, keep it
   as a clearly named scenario rather than hiding it behind fake plugin API.

## Failure and isolation rules

- No silent skips when Docker or browser dependencies are missing; the E2E job
  fails with an actionable error.
- Tests start no external servers except throwaway containers; never infer an
  existing Mattermost from `MM_SERVICESETTINGS_SITEURL` or developer config.
- Container images are pinned and overrideable; readiness and plugin activation
  have explicit timeouts. Setup failure also tears down allocated resources.
- Secrets are test-only, temporary, and never logged. No `insecureSkipVerify` or
  disabled TLS checks for the production MCP client.
- One CI job runs `make bundle`, `make build-mcp`, then the suite. Store traces,
  screenshots and server logs on failure. Existing unit/build CI stays intact.

## Known boundary

No native Mattermost mobile client can load webapp-plugin components; the E2E
asserts the server-persisted, locale-neutral post-body marker that mobile sees.
The standalone mobile application is outside this repository's test harness.
