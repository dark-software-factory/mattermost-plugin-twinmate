# Twinmate — product brief

**Status:** discovery, 2026-11-06. This is a product input, **not** an approved technical design or an implementation specification.

## Problem

People want to delegate some communication and work reporting to their own external AI agent, while staying in the Mattermost conversations they already use. Colleagues must know whether a message is from the human or an AI delegate.

## Confirmed outcomes

1. **Original personal conversation:** an agent can send a message into an existing direct conversation between exactly two humans, from its owner's user identity; no added bot member or replacement conversation.
2. **Visible delegation:** the receiver can recognize AI-delegated content, including in an unmodified native Mattermost mobile application.
3. **Owner-scoped access:** delegation is granted by the owner and never grants access or posting privileges beyond that person's current Mattermost permissions.
4. **Agent independence:** developers can use OpenClaw or other independently running agents. Mattermost should not dictate the agent runtime.

## Goals to validate during discovery

- Retrieve relevant Mattermost context without manual copy/paste.
- Allow an agent to publish results into appropriate conversations and threads.
- Allow optional automatic handling of incoming personal messages with explicit, understandable controls.
- Minimize custom integrations by reusing first-party Mattermost and MCP features where their security and license models fit.

## Non-goals of this phase

- Shipping, deploying, or packaging a plugin.
- Recreating the previous Go/MCP/webapp implementation.
- Choosing a licensing tier, transport, datastore, queue, or callback protocol before research.
- Preserving compatibility with the deleted prototype APIs.

## Unresolved product and policy questions

1. Is **Free/Team Edition** support mandatory, or is Enterprise acceptable?
2. Who may enable automatic DM forwarding, and must the other participant explicitly consent before content leaves Mattermost?
3. Which conversations may an agent read proactively, and what evidence of permission is required for each access?
4. Should the user be able to review/approve some generated messages before publication?
5. What should appear on mobile: a text label, bot-style badge where native, or another cross-client mechanism?
6. Must agents work exclusively from local machines behind NAT, without inbound network access?
7. What audit, retention, revocation, and incident-response behavior is necessary?
8. Which Mattermost versions and deployment modes are in scope?

Answers become requirements only after explicit review. Research hypotheses are not equivalent to user decisions.
