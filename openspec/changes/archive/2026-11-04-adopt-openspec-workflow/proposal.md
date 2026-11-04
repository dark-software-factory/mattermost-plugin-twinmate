## Why

Twinmate already has working REST/MCP, DM forwarding and provenance features,
but a developer or coding agent has to reconstruct their contracts from the
README, source and historical design docs. Future changes risk unintentionally
loosening privacy requirements or losing mobile AI attribution.

## What Changes

- Initialize the stock OpenSpec `spec-driven`/`core` lifecycle for Codex and Cursor.
- Backfill baseline specs documenting the **existing** agent access, DM opt-in,
  and AI-attribution contracts. No new product behavior is requested.
- Add guidance to `AGENTS.md`, project context/rules to `openspec/config.yaml`,
  and a practical propose/apply/verify/archive workflow to `README.md`.
- Add a strict OpenSpec validation check to the existing CI and Makefile.

## Capabilities

No capability delta: this is a tooling and documentation migration only.
The three baseline capability specs describe the existing v0.1 behavior.
The change explicitly declares `skip_specs: true` rather than inventing
an ADDED requirement for a feature that is already implemented.

## Impact

- No changes to Go plugin, React bundle, MCP protocol, or deployed dependencies.
- Developer/CI tool: pinned OpenSpec 1.16.0 (Node.js 22.12+).
- Rollback: remove the OpenSpec CI step and generated workflow artifacts;
  runtime and user data need no migration or rollback.
