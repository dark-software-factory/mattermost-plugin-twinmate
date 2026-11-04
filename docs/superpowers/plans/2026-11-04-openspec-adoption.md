# Twinmate OpenSpec Adoption Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Establish an actionable OpenSpec workflow and validated source of truth for Twinmate's existing behavior.

**Architecture:** Use official stock `spec-driven` + `core` profile, generated Codex/Cursor instructions, three capability specs, a tooling-only change archive, and a strict CI/Makefile check. The plugin runtime is unchanged.

**Tech Stack:** OpenSpec CLI 1.16.0, Markdown/YAML, GitHub Actions, Go, Node 24, Playwright.

**Spec:** `docs/superpowers/specs/2026-11-04-openspec-adoption-design.md`.

## Global Constraints

- Support Mattermost 11.0+; localized web badge exists on 11.9+ only.
- Do not change production Go/webapp/MCP logic in this adoption.
- CI validates both current OpenSpec specs and completed archives.
- Integrate official Codex and Cursor workflows, pin OpenSpec 1.16.0.

## Review Focus

- False spec claims about missing features: compare baseline requirements to README and integration tests.
- Empty spec changes passing unnoticed: ensure tooling-only changes use `skip_specs: true` explicitly.
- Incomplete archived tasks: run `openspec validate --archived` in CI.
- Local and CI CLI mismatch: use a single pinned version in commands.
- Accidentally treating schema validation as functional coverage: retain Go/webapp/E2E checks.

## Task 1: Initialize and configure OpenSpec

**Files:** `openspec/config.yaml`, `.agents/skills/`, `.cursor/`, `AGENTS.md`.

- [x] Generate official `core` profile files for Codex/Cursor using OpenSpec 1.16.0.
- [x] Add concrete project constraints and per-artifact/operation guidance.
- [x] Confirm `openspec list --json` resolves this repository as its root.

## Task 2: Baseline contracts and migration record

**Files:** `openspec/specs/{agent-access,automatic-dm,ai-attribution}/spec.md`, `openspec/changes/archive/`.

- [x] Document current externally visible behavior and denial scenarios.
- [x] Create tooling-only `adopt-openspec-workflow` proposal, design, tasks and `skip_specs: true`.
- [x] Validate and archive only after the implementation and tests are complete.

## Task 3: Developer experience and CI

**Files:** `README.md`, `Makefile`, `.github/workflows/ci.yml`.

- [x] Document Codex/Cursor usage, lifecycle and difference between CLI and AI commands.
- [x] Add `make spec-check` using pinned OpenSpec CLI.
- [x] Run actual OpenSpec validation locally, inspect status/archives, and add equivalent CI checks.
- [x] Verify no production files were changed; open a focused GitHub PR.
