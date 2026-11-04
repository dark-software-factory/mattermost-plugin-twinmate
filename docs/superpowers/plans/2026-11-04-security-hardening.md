# Twinmate security hardening plan

**Goal:** Implement `openspec/changes/security-hardening` without weakening
local MCP or native-mobile AI attribution.

**Source:** `openspec/changes/security-hardening/specs/security-boundaries/spec.md`

## Task 1: Callback network boundary

- Write failing tests in `server/agent/client_test.go` and `server/settings_test.go` for denied blanket private access, explicit allowlist, unsafe DNS and port mismatch.
- Implement allowlisted validation and DNS-aware transport; wire plugin settings and E2E fixture.
- Verify `go test ./server/agent ./server` (CI if Go unavailable locally).

## Task 2: Personal key authorization

- Write failing tests for read-only default and channel-scoped read/write in `server/access_tokens_test.go` and `server/agent_api_test.go`.
- Implement token capability records, issue-time validation, endpoint enforcement and web UI selection; adapt E2E issued keys.
- Verify Go and webapp unit tests.

## Task 3: DM privacy and reply reauthorization

- Write failing command and permission tests.
- Add public disclosure on first opt-in, check active recipient, and gate auto-publish with write permission checks.
- Verify processor and E2E DM scenarios.

## Task 4: Attribution integrity and bounded history

- Write failing tests for edit bypass, label spoof, sparse thread and fixed scan budget.
- Add hook based on server provenance and bounded paged thread reads.
- Verify Go unit and E2E history tests.

## Task 5: Close lifecycle

- Update README, baseline capability specs, verify OpenSpec, all tests, CI.
- Archive completed OpenSpec change, push feature branch and open reviewable PR.
- Review final changed files against security boundaries; never claim Go/E2E green without CI evidence.
