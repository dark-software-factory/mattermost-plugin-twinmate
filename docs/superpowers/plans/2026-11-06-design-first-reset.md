# Design-First Repository Reset Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Remove the obsolete Twinmate prototype and leave a research-, ADR- and OpenSpec-first repository without a runtime implementation.

**Architecture:** The repository contains confirmed product constraints, dated external research, decision records, historical design notes and an empty active OpenSpec specification area. CI only checks design artifacts; future product code is blocked on approved architectural design.

**Tech Stack:** Markdown, YAML, Make, GitHub Actions and OpenSpec CLI 1.16.0 (design-time only).

**Spec:** `docs/product/brief.md`; accepted decision: `docs/decisions/0001-design-first-reset.md`.

## Global Constraints

- The user's explicit request is to remove old implementation code **without an archive**.
- Do not preserve old code by creating a backup commit or moving source into another directory.
- Preserve historical design documents but mark them non-authoritative.
- The existing two-person DM and cross-client AI attribution requirements remain product constraints, not settled APIs.
- No application runtime or packaging is created in this reset.

## Review Focus

- A developer might mistake historical OpenSpec specs for active requirements — move them out of `openspec/specs/`.
- A developer might assume the old plugin is still runnable — remove build/test commands and state design-only status in README.
- CI might accidentally claim that product tests passed — use a design-only workflow name and commands.
- Requirements might prematurely assume Enterprise, Go or existing protocol contracts — keep those as open questions.
- Existing tooling might still refer to deleted Go, npm or E2E targets — scan the active configuration and workflows.

---

### Task 1: Neutral research/design foundation

**Files:** `README.md`, `AGENTS.md`, `openspec/config.yaml`, `docs/product/brief.md`, `docs/research/`, `docs/architecture/`, `docs/decisions/`, `docs/history/`.

- [x] Replace legacy implementation instructions with an explicit design-only workflow.
- [x] Record confirmed needs separately from unconfirmed design hypotheses.
- [x] Record the accepted reset and dated Mattermost research evidence.
- [x] Verify all referenced documentation paths exist.

### Task 2: Remove only the obsolete implementation

**Files:** `cmd/`, `internal/`, `server/`, `webapp/`, `e2e/`, `go.mod`, `plugin.json`, `openspec/specs/*.md`.

- [x] Verify exact deletion targets are within the Twinmate repository.
- [x] Remove the runtime source/dependencies without an archive, as explicitly authorized.
- [x] Move the four previous active specs into `docs/history/legacy-specs/` as historical design text.
- [x] Verify the previous implementation directories/manifests no longer exist.

### Task 3: Replace CI with design checks

**Files:** `Makefile`, `.github/workflows/design.yml`, `.gitignore`.

- [x] Remove build, E2E and package workflows.
- [x] Add a local `make check` gate and a pinned `make spec-check` target.
- [x] Run the available checks and inspect the final project tree.
