## MODIFIED Requirements

### Requirement: Connection and conversation opt-in

An external AI connection and automatic forwarding MUST start disabled.
Each owner MUST enable an HTTPS agent connection and explicitly opt in to
each individual one-to-one DM using `/twinmate on`. Before enabling a new
conversation the plugin MUST post a durable notice visible to both
participants that new messages may be transmitted to an external AI service.
If the notice cannot be posted, forwarding MUST remain off. The commands
`/twinmate off` and `/twinmate status` MUST manage or report opt-in state.
The opt-in MUST NOT govern agent-initiated read/write API calls.

#### Scenario: DM is disabled by default

- GIVEN a user configures an agent endpoint but has not opted into a DM
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

### Requirement: Bounded HTTPS callback

Twinmate MUST send a scoped JSON message event to the configured HTTPS
endpoint, with a stable event ID and the owner's optional bearer token.
It MUST reject redirects and unsafe destinations; connections to private,
loopback, link-local or special-use networks MUST be blocked unless an
administrator enables private endpoints **and** allowlists the exact host
and port. It MUST recheck resolved IP addresses when dialing and enforce
bounded timeouts and response sizes.

#### Scenario: Agent receives a scoped event

- GIVEN a valid opted-in DM and permitted HTTPS destination
- WHEN a new human message arrives
- THEN the callback contains the sender, owner, channel, message ID and text
- AND the callback has a stable event identifier
- AND the request includes the configured bearer token when one exists

#### Scenario: Private destination without administrative allowance

- GIVEN private-network callback endpoints are not enabled or the host is not allowlisted
- WHEN the owner configures an agent using a loopback or private-network address
- THEN the plugin refuses to contact that destination

### Requirement: Optional reply in the same DM

The callback MAY return an empty reply to stay silent. For a non-empty reply,
Twinmate MUST recheck the connection, DM opt-in, active owner status,
conversation membership and write permissions immediately before publishing
a bounded, AI-attributed message in the same one-to-one conversation. Delivery
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

#### Scenario: Owner is deactivated during callback

- GIVEN the reply is pending and the owner is deactivated
- WHEN the callback returns
- THEN Twinmate does not publish under the deactivated account
