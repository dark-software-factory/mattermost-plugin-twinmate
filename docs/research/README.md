# Research

Research precedes OpenSpec proposals and architecture decisions. Start with first-party Mattermost documentation/source and reproducible behavior; include independent projects when comparing alternatives.

For each investigation, record:

1. **Question** and why its answer affects product behavior.
2. **Evidence** with URL, version, license tier and date checked.
3. **Verified observations** versus assumptions requiring a spike.
4. **Alternatives**, their restrictions and threat/privacy implications.
5. **Recommendation or unanswered question** and what would falsify it.

Research does not by itself authorize implementation. Transfer accepted conclusions to the product brief, an ADR and an OpenSpec proposal.

## Starting investigations

- [Mattermost building blocks](mattermost-building-blocks.md): native MCP, Agents, automation, server/plugin hooks and mobile limitations.
- Server/plugin authorization, identity attribution and post provenance in existing two-person DM channels.
- Protocol options for developer-owned local agents (pull-based MCP versus event callbacks).
- Mattermost edition/version compatibility and third-party plugin risk.
- Privacy model for granting an external agent access to personal conversations.
