# Architecture workspace

**Status:** no selected implementation architecture.

After research, candidate architecture documents must describe:

- Components and dependencies: Mattermost server/plugin, optional Agents/MCP, external user-owned agent, optional local adapter.
- Ownership of authentication, authorization, provenance, messages, and consent state.
- How the agent acts **as the owner in an unchanged two-person DM** without bypassing Mattermost permissions.
- How AI attribution remains visible on native mobile and is verified independently of client-controlled fields.
- Data flows for pull-based history, outgoing publications and optional incoming-message events.
- Deployment/version/license boundaries, threat model, failure paths and migration/rollback.
- Contract tests and real Mattermost integration tests needed before shipping.

Compare at least the first-party-only, hybrid and standalone approaches in [the research backlog](../research/mattermost-building-blocks.md). Record the chosen approach as an accepted ADR and an OpenSpec `design.md` after the user's review; do not implement from this placeholder.
