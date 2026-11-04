# Twinmate agent instructions

Use OpenSpec for feature work, behavior changes, API contracts, and security-sensitive fixes.

1. Read `openspec/config.yaml`, the affected `openspec/specs/*/spec.md`, and any active change before proposing code.
2. Capture the request with **OpenSpec propose**: `proposal.md` (why/scope), capability delta `specs/` (what), `design.md` (how), and `tasks.md` (verification checklist). Review the artifacts before implementation.
3. Implement via **OpenSpec apply**. Update the artifacts when discoveries change the agreed behavior. Mark tasks complete only after checking results.
4. Run `make spec-check` and relevant Go, webapp, and E2E tests. For cross-boundary behavior run `make test-e2e` (Docker required); CI tests Mattermost 11.0 and 11.11.
5. **Sync and archive** the completed change on the same PR before merging, so `openspec/specs/` always describes the merged behavior. CI validates current specs and completed archives.

Use `$openspec-propose`, `$openspec-apply-change`, `$openspec-archive-change` in Codex; use `/opsx-propose`, `/opsx-apply`, `/opsx-archive` in Cursor. The optional explore, update and sync workflows are installed too. CLI equivalents include `openspec new change`, `openspec status`, `openspec instructions`, `openspec validate`, and `openspec archive`.

For pure tooling/refactors that do **not** affect observable behavior, explicitly set `skip_specs: true` in the change's `.openspec.yaml`. Do not invent a fake behavioral delta. Minor typo-only corrections need no new change.

Project invariants:

- Mattermost **11.0+** server plugin in Go; React webapp; Go stdio MCP client; no Node.js dependency at plugin runtime.
- A personal Twinmate key never authorizes another person's messages. Recheck Mattermost permissions on each read/write and keep provenance server-owned.
- AI messages must remain visibly attributed on native mobile via the persisted body marker; web badges must not trust client-supplied props.
- Automatic one-to-one DM forwarding starts **off**, requires explicit opt-in, and never silently extends to group channels.
- Preserve existing REST/MCP behavior unless the change's delta spec intentionally changes it.

Current contracts: `openspec/specs/`. Historical implementation decisions: `docs/superpowers/`.
