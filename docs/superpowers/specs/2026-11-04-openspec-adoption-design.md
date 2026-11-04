# Twinmate OpenSpec Adoption — Design

## Intent

Make requirements reviewable and durable while AI agents implement Twinmate.
The plugin already has Go/MCP endpoints, optional automatic DM replies,
and browser/mobile attribution. This work formalizes their observable behavior
without changing plugin runtime functionality.

## Approaches considered

1. **README and AGENTS.md only.** Smallest footprint, but no structured
   behavioral deltas or deterministic validation; agents can silently drift.
2. **Stock OpenSpec `spec-driven` schema, `core` profile (chosen).** Keep
   baseline capabilities, proposal/specs/design/tasks for changes, official
   generated Codex/Cursor integrations, and strict CLI validation in CI.
   Uses established tooling and a simple change lifecycle.
3. **Custom schema with mandatory approval/verification artifacts.** Could
   encode team governance, but duplicates what proposal/tasks/tests already
   cover and creates a maintenance burden in this early preview.

## Lifecycle

1. Explore the request and baseline specs; `/opsx:explore` is optional.
2. Propose one kebab-case change: `proposal.md`, `specs/<capability>/spec.md`
   with ADDED/MODIFIED/REMOVED requirements, `design.md`, and `tasks.md`.
3. Human reviews the change; apply tasks and update deltas as needed.
4. Run tests and `openspec validate --all --strict`; use explicit verification
   of behavior, not only Markdown validation.
5. Sync and archive the change in the feature PR before merge so baseline
   `openspec/specs/` is updated atomically with code. Use `skip_specs: true`
   only when no observable contract changes.

## Artifacts and ownership

- `openspec/specs/agent-access/spec.md`: personal token, read-by-reference,
  pagination, authorized writes, outbound stdio MCP.
- `openspec/specs/automatic-dm/spec.md`: disabled defaults, DM opt-in, HTTPS
  callback, response and destination limits.
- `openspec/specs/ai-attribution/spec.md`: durable labels and server-verified
  localized web badge, including legacy/mobile fallback.
- `openspec/config.yaml`: Twinmate-specific AI guidance and artifact rules.
- `.agents/skills/` and `.cursor/`: official generated workflow instructions.
- `AGENTS.md` and `README.md`: human and generic-agent entry points.
- `.github/workflows/ci.yml` and `Makefile`: pin OpenSpec CLI and fail on
  invalid current specs or incomplete archives.

The existing `docs/superpowers/` records stay as historical decisions, not
competing authoritative descriptions. Baseline specs cover observed v0.1
behavior and must not imply missing features (search, attachments, outbox,
group replies, or native mobile plugin components) are already implemented.

## Quality, security and limitations

- No new dependency inside the Go plugin, MCP binary, or React bundle.
- OpenSpec CLI is developer/CI tooling, pinned to `1.16.0`, Node >=22.12.
- `openspec validate` checks document structure and deltas, not code/spec
  semantic equivalence. Existing Go, webapp and integration tests remain
  required; CI validation complements them.
- CI validates active changes and current specs strictly, and separately
  rejects archived changes with unfinished tasks. The CLI does not itself
  prove every code PR has a matching change, so `AGENTS.md` documents that
  review requirement.
- The adoption itself is a tooling-only OpenSpec change, recorded with
  `skip_specs: true`; newly backfilled baseline specs describe already
  implemented behavior rather than proposing new product features.
