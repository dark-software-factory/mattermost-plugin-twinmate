# Automatic Direct Messages

## Purpose

Define the optional, explicitly authorized forwarding of new one-to-one
Mattermost messages to a user's external AI agent and the delivery of a
generated reply in the original conversation.

## Requirements

### Requirement: Connection and conversation opt-in

An external AI connection and automatic forwarding MUST start disabled.
Each owner MUST enable an HTTPS agent connection and explicitly opt in to
each individual one-to-one DM using `/twinmate on`. The commands
`/twinmate off` and `/twinmate status` MUST manage or report the opt-in state
for that conversation. The opt-in state MUST NOT govern agent-initiated
read/write API calls.

#### Scenario: DM is disabled by default

- GIVEN a user configures an agent endpoint but has not opted into a DM
- WHEN another human sends a new message to that DM
- THEN no message event is forwarded to the configured agent

#### Scenario: Disable forwarding after opt-in

- GIVEN the user's agent and a particular DM are enabled
- WHEN the user runs `/twinmate off` in that DM
- THEN later messages in that DM are not forwarded to the agent

### Requirement: Only new human one-to-one messages trigger forwarding

Twinmate MUST forward only new messages from the other human participant
of an opted-in one-to-one DM. It MUST NOT forward a group conversation,
historical messages, bot-generated messages, or its own AI replies through
this automatic pathway.

#### Scenario: Human participant asks an opted-in recipient

- GIVEN the recipient enabled their connection and this one-to-one DM
- WHEN the other participant publishes a new human-authored message
- THEN the recipient's configured agent receives an event for that message
- AND the same event is not sent to an unrelated user's agent

#### Scenario: Bot or group message

- GIVEN an enabled agent connection
- WHEN a bot post or a message in a group channel is created
- THEN Twinmate does not invoke the automatic DM callback

### Requirement: Bounded HTTPS callback

Twinmate MUST send a scoped JSON message event to the configured HTTPS
endpoint, with a stable event ID and the owner's optional bearer token.
It MUST reject redirects and unsafe destinations; connections to private,
loopback, or link-local networks MUST be blocked unless an administrator
explicitly allows private agent endpoints. It MUST enforce bounded timeouts
and response sizes.

#### Scenario: Agent receives a scoped event

- GIVEN a valid opted-in DM and permitted HTTPS destination
- WHEN a new human message arrives
- THEN the callback contains the sender, owner, channel, message ID and text
- AND the callback has a stable event identifier
- AND the request includes the configured bearer token when one exists

#### Scenario: Private destination without administrative allowance

- GIVEN private-network callback endpoints are not enabled by the administrator
- WHEN the owner configures an agent using a loopback or private-network address
- THEN the plugin refuses to contact that destination

### Requirement: Optional reply in the same DM

The callback MAY return an empty reply to stay silent. For a non-empty reply,
Twinmate MUST recheck the connection and DM opt-in before publishing a
bounded, AI-attributed message in the same one-to-one conversation. Delivery
is best-effort; Twinmate does not guarantee a durable retry queue.

#### Scenario: Agent returns a reply

- GIVEN the callback returns a valid non-empty reply while the DM remains enabled
- WHEN Twinmate publishes the response
- THEN the response appears in that same DM under the agent owner's user ID
- AND the response carries Twinmate AI attribution

#### Scenario: Owner disables the DM while callback is in flight

- GIVEN a callback is already running
- WHEN the owner disables the DM before its reply is published
- THEN Twinmate does not publish that pending reply
