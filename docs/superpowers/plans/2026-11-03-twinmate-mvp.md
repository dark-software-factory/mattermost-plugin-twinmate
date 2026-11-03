# Twinmate MVP Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Ship an installable Mattermost plugin that connects personal HTTPS agents and posts visibly attributed replies inside existing direct-message channels.

**Architecture:** Keep Mattermost-specific hooks and KV configuration at the Go plugin boundary. Put HTTP agent protocol and message routing into independently testable packages. Package a web UI and use text labeling for native mobile fallback.

**Tech Stack:** Go, Mattermost public plugin SDK, React, Webpack, GitHub Actions.

**Spec:** `docs/superpowers/specs/2026-11-03-twinmate-mvp-design.md`

## Global Constraints

- Start with direct messages only, no group chat.
- Never hand an agent a Mattermost PAT or whole conversation history.
- Do not post a twin message without an explicit owner configuration and DM opt-in.
- Include server-owned AI provenance and a mobile-visible Markdown label.
- Default-deny private/loopback agent endpoints unless admin explicitly enables them.

## Review Focus

- Agent endpoint SSRF, redirect, excessive body, invalid JSON and timeout.
- Both chat participants having a twin enabled: one human message must not cause infinite AI replies.
- Wrong/missing channel participants: never send data to a misidentified owner.
- Duplicate hook execution and plugin restart: document delivery semantics and avoid false exactly-once claims.
- Browser settings authorization and secret readback: no cross-user configuration access.

---

### Task 1: Agent protocol and safety

**Files:** `server/agent/client.go`, `server/agent/client_test.go`

**Interfaces:** `Client.Call(ctx, Config, Event) (string, error)`; HTTPS URL check, safe HTTP transport, bounded JSON response.

- [ ] Write tests for successful reply, blank reply, malicious URL, redirect, timeout, oversized/malformed response.
- [ ] Run failing tests and confirm missing implementation.
- [ ] Implement minimal client; run tests to green.
- [ ] Commit tested change.

### Task 2: DM routing and plugin adapter

**Files:** `server/core/processor.go`, `server/core/processor_test.go`, `server/plugin.go`, `server/main.go`

**Interfaces:** `Processor.Handle(ctx, Message) error`, platform interface for `Recipient`, `Settings`, `DMEnabled` and `Publish`; Mattermost hook translates the post and publishes labeled responses.

- [ ] Test incoming human DM, disabled owner, disabled DM, twin-generated messages, missing recipient, agent decline, and ownership.
- [ ] Run tests red, implement adapter and routing, run tests green.
- [ ] Commit tested change.

### Task 3: Self-service control and UI

**Files:** `server/http.go`, `server/config.go`, `server/command.go`, `server/http_test.go`, `webapp/src/index.jsx`, `webapp/package.json`, `webapp/webpack.config.js`

**Interfaces:** `GET/PUT /api/config` bound to Mattermost injected user ID; `/twinmate on|off|status` scoped to invoking user and current DM. Webapp main-menu settings and post-header badge.

- [ ] Test request authorization, secret redaction, endpoint validation, and enable/disable persistence.
- [ ] Run tests red, implement and run tests green.
- [ ] Build webapp and verify bundle.
- [ ] Commit tested change.

### Task 4: Distribution and CI

**Files:** `plugin.json`, `go.mod`, `Makefile`, `.github/workflows/ci.yml`, `README.md`, `.gitignore`

**Interfaces:** `make test`, `make bundle`; `.tar.gz` contains plugin manifest, server binary and web bundle.

- [ ] Add build and workflow checks; verify failure before the missing pieces exist.
- [ ] Package and build on GitHub Actions; inspect failures and correct them.
- [ ] Run full test suite, inspect artifact contents, commit and open a PR.
