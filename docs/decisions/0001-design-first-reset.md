# ADR 0001 — Reset Twinmate to design-first development

**Status:** Accepted 2026-11-06

## Context

The existing preview plugin and MCP client were built before confirming how official Mattermost plugins might be composed. The user requested deleting the implementation without retaining a code archive, and retaining a clean process for research, architecture and OpenSpec reviews.

## Alternatives

- **Continue extending the prototype:** preserves work, but biases future design toward its undocumented integration choices.
- **Preserve the prototype in a backup branch:** enables rollback to code, but adds retention and cleanup that the user explicitly declined.
- **Remove runtime code and reset the project to design artifacts:** gives research and design an independent foundation, but loses local implementation and tests.

## Decision

Use the third option. Remove product source, dependencies, packaging and runtime CI without a code archive. Retain prior design/decision documents as clearly marked historical references only. Current OpenSpec specifications will be empty until replaced through an approved design.

## Consequences

- No runnable plugin or backward compatibility is claimed.
- Existing source is not recoverable from this local checkout, which has no commits or configured remote.
- Future work starts with research and explicit design review, including native-mobile behavior, two-person DM identity and security analysis.
- CI must indicate design-validation status, never application fitness.

**Evidence:** user's explicit approval in the project conversation on 2026-11-06.
