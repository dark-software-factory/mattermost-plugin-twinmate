# Mattermost building blocks — initial reconnaissance

**Checked:** 2026-11-06. **Status:** evidence to verify against the selected Mattermost versions; **no architecture is approved**.

## First-party building blocks

### Mattermost Agents and embedded MCP

- The official Agents plugin provides embedded Mattermost MCP tools for reading and searching posts, channels and users, as well as tools for writing and file operations.
- Its optional external Streamable HTTP endpoint can aggregate eligible plugin-registered MCP tools, but external MCP access and state-changing native tools are licensed at Enterprise and above in the documented version.
- The external endpoint requires Mattermost Server 11.2+ and supports OAuth 2.0 or personal access tokens.
- Research whether the external MCP alone can preserve verified *delegated* provenance; ordinary `create_post` is not assumed equivalent to a Twinmate action.

Sources:
- [Mattermost Agents admin guide — MCP](https://docs.mattermost.com/administration-guide/configure/agents-admin-guide)
- [Mattermost Agents repository](https://github.com/mattermost/mattermost-plugin-agents)

### Channel Automation

- The [official Channel Automation plugin](https://github.com/mattermost/mattermost-plugin-channel-automation) reacts to events in configured channels, can invoke an Agents-backed AI prompt, and publishes messages as a bot.
- **Unverified for this need:** whether any suitable trigger addresses an existing, human-only one-to-one DM. Do not assume channel automation replaces server hooks.

### Server and webapp plugins

- The [Mattermost plugin framework](https://docs.mattermost.com/developers/integrate/plugins/overview) allows Go server hooks and webapp UI extensions.
- [Native mobile plugins](https://docs.mattermost.com/developers/integrate/plugins/components/mobile) do not offer general webapp-style React customization; interactive dialogs/messages are supported cross-platform.
- Consequently, any trustable mobile attribution mechanism must be evaluated on real native clients before selecting a design.

### DM identity boundary

- A conversation between two human participants must remain their **existing one-to-one DM**.
- A distinct bot account may provide visible bot identity but changes conversation topology if introduced as a third participant. The [Mattermost import documentation](https://docs.mattermost.com/administration-guide/onboard/bulk-loading-data) distinguishes two-member direct channels from multi-member group channels.
- The feasibility and exact permission implications of server-mediated `post.UserId` delegation must be confirmed against official source and real tests, not inferred from the previous prototype.

## Candidate compositions to compare

1. **First-party-only:** Agents + MCP + automation + bot accounts. Likely fails the unchanged two-person DM identity constraint; explicitly test.
2. **Hybrid:** first-party MCP for general read/search plus a minimal Twinmate plugin for per-owner delegation, permissions, and mobile-visible provenance. May impose Enterprise licensing.
3. **Standalone:** minimal Twinmate plugin and independent per-user local MCP client, with optional interoperable generic Mattermost MCP. Avoids licensing dependency but duplicates some capabilities.

No candidate is selected yet. The next design review must compare functionality, licensing, threats and maintenance cost with testable evidence.
