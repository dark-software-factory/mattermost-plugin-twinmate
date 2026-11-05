## MODIFIED Requirements

### Requirement: Connection and conversation opt-in

Automatic DM forwarding and external connections MUST start disabled.
Owners MUST configure an HTTPS agent and opt into each one-to-one DM using
`/twinmate on`. Twinmate MUST publish a durable disclosure visible to both
participants and bind the opt-in to that exact post. If the notice is edited,
deleted, unverifiable or mismatched, forwarding MUST stop until a new opt-in.
`/twinmate off` and `/twinmate status` MUST manage or report forwarding.
DM opt-in MUST NOT restrict agent-initiated API calls.

#### Scenario: DM is disabled by default

- GIVEN an agent endpoint is configured but the DM is not opted in
- WHEN another human sends a new message to that DM
- THEN no message event is forwarded to the configured agent

#### Scenario: Disable forwarding after opt-in

- GIVEN the user's agent and a particular DM are enabled
- WHEN the user runs `/twinmate off` in that DM
- THEN later messages in that DM are not forwarded to the agent

#### Scenario: First opt-in notifies the other participant

- GIVEN forwarding is off for the conversation
- WHEN the owner uses `/twinmate on`
- THEN a persistent notice appears in the DM before forwarding becomes active

#### Scenario: Notice is edited or deleted

- GIVEN a DM has an active opt-in bound to its disclosure post
- WHEN the disclosure is edited or deleted before the next incoming message
- THEN no further message is forwarded to the external agent
- AND the owner must re-enable forwarding to publish a new notice
