## Context

Twinmate has established functionality, Go unit tests, browser tests and an
integration suite against Mattermost 11.0/11.11. Existing `docs/superpowers/`
records are historical design notes, not a machine-validated contract.

## Goals / Non-Goals

- Provide one current behavioral source of truth and a repeatable delta workflow.
- Make the standard CLI/AI commands usable in Codex and Cursor.
- Validate artifact structure without changing the runtime dependency graph.
- Do not introduce custom approval servers, adapters, or a custom OpenSpec schema.
- Do not migrate every past design note into an OpenSpec archive retroactively.

## Decisions

1. Keep the stock `spec-driven` schema and `core` command profile, generated
   by the official OpenSpec 1.16.0 initializer. Config rules capture Twinmate's
   trust boundaries and compatibility constraints; no custom templates.
2. Backfill three baseline capability specs, grouped by externally observable
   API/automation/attribution behaviors. Future code changes use delta specs
   and archive into the same paths.
3. Document the normal PR cycle as explore → propose → human review → apply
   → tests/validate → sync/archive → merge. Maintain task checkboxes.
4. Use `skip_specs: true` for this tooling-only adoption change. Complete
   all tasks and archive it before merging the PR.
5. Pin CLI version in `Makefile` and Actions. Validation runs both
   `openspec validate --all --strict` and `openspec validate --archived`.

## Risks / Trade-offs

- The generated Codex/Cursor skills contain some duplicated boilerplate but
  remain regenerable with `openspec update` and consistent with upstream.
- OpenSpec's validator cannot prove requirements match Go behavior. Reviewers
  must still compare deltas to implementation and run existing tests.
- Future agents may skip the process; `AGENTS.md` and PR review are the
  human policy layer. Avoid pretending CLI validation can enforce every diff.

## Migration / Rollback

There is no database migration, API change or runtime rollout. Remove the
workflow files and CI check to roll back the process adoption.
