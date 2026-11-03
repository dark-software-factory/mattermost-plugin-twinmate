# Twinmate integration suite implementation plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Prove Twinmate works when installed in an actual Mattermost 11 server, accessed by real users, a browser and the compiled MCP client.

**Architecture:** Playwright starts and destroys Mattermost/Postgres Testcontainers. Setup installs the plugin with `mmctl`. API tests use Mattermost HTTP sessions, UI tests use Chromium, and MCP tests launch the real Go binary behind a temporary trusted TLS reverse proxy.

**Tech Stack:** Node 24, Playwright, Testcontainers, PostgreSQL 15, Mattermost 11, Go stdio MCP, GitHub Actions.

**Spec:** `docs/superpowers/specs/2026-11-04-twinmate-integration-design.md`

## Global constraints

- Use the exact plugin bundle and MCP binary produced by `make bundle build-mcp`.
- Use ephemeral Docker services; no modification of existing Mattermost sites.
- Retain the production `X-Twinmate-Token` authentication, permissions and HTTPS-only MCP behavior.
- Keep the suite independent from unit tests and require it as a separate CI job.

## Review focus

- Plugin install appears successful but its executable crashes: fail during setup, not hours later in a browser.
- Creating a token with a normal Mattermost session: exercise real CSRF behavior; do not inject `Mattermost-User-Id` manually.
- An outsider's agent token accesses a private channel/DM: verify 403 without private message data.
- A retry after a failed post or provenance write: avoid duplicate chat messages.
- E2E setup throws midway: stop all containers and retain failure logs.

### Task 1: Real Mattermost fixture and startup contract

**Files:** `e2e/package.json`, `e2e/playwright.config.js`, `e2e/helpers/server.js`, `e2e/helpers/setup.js`, `e2e/tests/installation.spec.js`.

**Interface:** Setup supplies a fresh baseURL/admin and user credentials; `server.start()` starts Mattermost/Postgres and activates the bundled plugin.

- [ ] Write a smoke spec asserting the plugin is active and the server responds.
- [ ] Verify it fails without a running fixture.
- [ ] Add Testcontainers setup with bounded startup and cleanup, using `mmctl --local plugin add/enable`.
- [ ] Run the smoke spec; verify it passes against a real container.

### Task 2: Token API, authorization, paging and posting

**Files:** `e2e/helpers/mm-api.js`, `e2e/tests/agent-api.spec.js`.

**Interface:** `mmAPI` performs real user sessions, Mattermost channel/post CRUD, and Twinmate API calls (no privileged mock methods).

- [ ] Add scenarios for token rotation/revocation, forbidden reads/writes, permalink resolution, before/after cursor paging, AI provenance and identity.
- [ ] Verify failing assertions against the real plugin, then resolve harness errors or production defects without weakening assertions.
- [ ] Run and verify the real API scenarios.

### Task 3: Compiled MCP and actual web app

**Files:** `e2e/helpers/mcp.js`, `e2e/tests/mcp.spec.js`, `e2e/tests/webapp.spec.js`.

**Interface:** MCP helper starts the compiled Linux binary and calls `tools/list`/`tools/call` over stdio using a local trusted TLS origin.

- [ ] Write failing MCP get/list/send and localized browser badge/settings tests.
- [ ] Implement only the test helpers and any production fixes their failures prove.
- [ ] Verify all scenarios, including no unsafe HTTP or credential leakage.

### Task 4: CI and docs

**Files:** `.github/workflows/e2e.yml`, `Makefile`, `README.md`, `.gitignore`.

- [ ] Add `make test-e2e` and a separate Actions job with Docker, build, Playwright, and failure artifacts.
- [ ] Run the full Actions E2E job and fix actual errors; do not claim green until observed.
- [ ] Request final code review and open a separate PR from `main`.
