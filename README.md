# Twinmate

Twinmate is a **design-stage** project exploring personal AI delegation in [Mattermost](https://mattermost.com/).

**There is no plugin implementation, MCP server, runnable application, or release in this repository.** The previous proof of concept was deliberately removed on 2026-11-06. Historical design notes remain for research only; they are not an implementation commitment.

## Confirmed product direction

- A person's externally operated agent can write **as that person in an existing one-to-one Mattermost direct message**, without adding a bot as a third participant or opening a separate conversation.
- Recipients must be able to distinguish delegated AI messages from human messages, including in unmodified native mobile clients.
- The system must preserve the owner's permissions and must not silently grant access to other users' conversations.
- Agents are not tied to one vendor or runtime; OpenClaw is one example, not a required component.

Everything beyond these high-level constraints is open to design, including exact attribution, consent, agent connectivity, MCP integration, supported Mattermost editions, and data handling. See [the product brief](docs/product/brief.md).

## Design workflow

1. **Research** — gather current first-party sources, reproduce relevant Mattermost behavior where practical, and record evidence in [`docs/research/`](docs/research/).
2. **Requirements** — distinguish confirmed outcomes from assumptions and open questions in [`docs/product/`](docs/product/).
3. **Architecture** — compare viable approaches and trust boundaries in [`docs/architecture/`](docs/architecture/), with decisions recorded as [`ADRs`](docs/decisions/).
4. **OpenSpec proposal** — create `openspec/changes/<name>/` containing `proposal.md`, capability deltas in `specs/`, `design.md`, and `tasks.md`.
5. **Design review** — obtain explicit approval before product code or runtime dependencies are introduced.
6. **Implementation and verification** — only after design approval; build tests from the accepted scenarios and archive the finished OpenSpec change.

The source of truth for **approved** behavior will be [`openspec/specs/`](openspec/specs/). It is intentionally empty while the replacement architecture is being designed. The old specs and previous project notes are historical records, not current requirements.

## Contributor commands

```bash
make check       # Lightweight structure and status checks, no runtime dependencies
make spec-check # OpenSpec CLI validation; requires Node.js/npm for the development CLI
```

The repository's generated OpenSpec workflows for Codex (`.agents/skills/`) and Cursor (`.cursor/`) are retained. Node.js is only used for the **design-time OpenSpec CLI**, not chosen as a runtime dependency.

Do not infer a working Mattermost plugin from a successful CI run: at this stage CI validates design files only.
