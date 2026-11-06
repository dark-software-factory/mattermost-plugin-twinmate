# Twinmate — agent collaboration rules

## Current phase: research and design only

There is **no active product implementation**. Do not create server, webapp, MCP, dependency manifests, packaging, or deployment pipelines before the user approves an architecture and its OpenSpec design artifacts. Do not treat historical examples as a request to rebuild them.

Read `README.md`, `docs/product/brief.md`, `openspec/config.yaml`, and the relevant research/ADRs before proposing a design. Prior work in `docs/superpowers/`, `docs/history/`, and `openspec/changes/archive/` is historical evidence, not a current specification.

## Fixed product constraints

1. A delegated agent must be able to publish in the **existing one-to-one DM**, as its human owner, without adding a third participant.
2. A recipient must be able to recognize an AI-delegated message on **native mobile**, not only through a webapp plugin component.
3. No agent may read or write beyond its owner's actual Mattermost permissions. Trust, credential handling, and provenance must be reasoned about server-side.
4. The external agent runtime is interchangeable; do not assume Mattermost Agents, OpenClaw, or a particular MCP implementation is mandatory.

These describe desired outcomes, not an approved implementation. Other choices (Free vs Enterprise, callback vs pull, consent, identity marking, data retention, supported versions) must be researched before being frozen into specs.

## Workflow

1. **Research first.** Start with vendor docs, source code, and small reproducible experiments. Record a dated source, observed behavior, version/license assumptions, and unresolved questions in `docs/research/`.
2. **Separate evidence and decisions.** Put confirmed product needs in `docs/product/`; compare alternatives and trust boundaries in `docs/architecture/`; use `docs/decisions/` for accepted ADRs. Never present an unverified technical assumption as a settled constraint.
3. **OpenSpec propose.** After research, produce a named change with `proposal.md` (why/scope/non-goals/rollback), `specs/` (observable behavior and GIVEN/WHEN/THEN scenarios), `design.md` (interfaces, permissions, failure modes), and `tasks.md` (verifiable steps).
4. **Review gate.** Get explicit user approval of the design **before implementation**. Update the proposal and ADRs whenever conclusions change.
5. **Implement only after approval.** Follow the approved OpenSpec change, add focused and end-to-end tests at relevant boundaries, then sync/validate/archive the accepted change.

Use `.agents/skills/openspec-propose/` in Codex or `.cursor/commands/opsx-propose.md` in Cursor for proposal generation. Explore/update/sync/archive workflows are installed in both environments. `make check` validates the design-first repository shape; `make spec-check` runs the pinned OpenSpec CLI.

## Design review checklist

- Reusing official Mattermost functionality versus owning a custom plugin, including version/license constraints.
- Existing one-to-one DM semantics and correct author identity.
- Visible, unforgeable AI provenance and native-mobile fallback.
- Per-user grants, identity propagation, revocation, access checks, and auditability.
- Explicit privacy/consent decisions for automatic forwarding, sensitive messages, and external retention.
- Connectivity for local agents behind NAT; delivery, retries, idempotency, and failures.
- Integration tests against real supported Mattermost versions before release.
