# Security Boundaries

## Purpose

Define the trust boundaries for external callbacks, personal agent keys,
message visibility and AI attribution. The server enforces these boundaries
independently of user-controlled messages, properties and client settings.

## Requirements

### Requirement: Restricted callback egress

Twinmate MUST NOT let an ordinary user reach a private, loopback, link-local
or special-use network through its callback unless the administrator enables
private callbacks and explicitly approves the destination's exact host and
port. DNS resolutions MUST be checked again on every connection.

#### Scenario: Global opt-in without a host allowance

- GIVEN the administrator enables private callbacks but leaves its host allowlist empty
- WHEN a user configures `https://127.0.0.1/agent`
- THEN Twinmate refuses the endpoint

### Requirement: Least-privilege personal agent keys

Newly issued keys MUST default to read-only. Owners MAY explicitly grant
send-as-owner and optionally restrict the key to specified channel IDs.
Every read and write MUST enforce token scope in addition to current
Mattermost permissions. Existing keys MUST remain usable until rotated.

#### Scenario: Agent attempts a write with read-only key

- GIVEN an owner issues a new key without opting into writes
- WHEN the agent attempts to send a channel message
- THEN Twinmate rejects the write with HTTP 403 and creates no post

#### Scenario: Agent requests a disallowed channel

- GIVEN a key scoped to channel A
- WHEN an agent requests a post from channel B or tries to publish there
- THEN Twinmate rejects access even if the owner has membership in B

### Requirement: DM participant disclosure

An owner enabling automatic DM forwarding MUST cause a visible notice in
the DM describing that incoming messages may be sent to their external
AI service. The opt-in MUST NOT silently succeed if the notice fails.

#### Scenario: DM opt-in

- GIVEN an opted-out one-to-one conversation
- WHEN an owner runs `/twinmate on`
- THEN both participants can see a durable AI forwarding notice
- AND only subsequently posted human messages may be forwarded

### Requirement: Verified and durable generated-message label

Twinmate MUST reject edits that remove the visible AI attribution from
server-recorded generated messages. A human-authored post that only mimics
the label MUST NOT be treated as trusted provenance for skipping DM events.

#### Scenario: Attempt to erase AI attribution

- GIVEN a server-recorded AI post
- WHEN its author edits away the label
- THEN Mattermost rejects the edit, leaving the label visible

### Requirement: Bounded historical scans

Twinmate MUST NOT load unbounded whole-thread contents when providing
agent history. It MUST constrain upstream fetch sizes and total work,
and report an explicit error when a complete thread page cannot be
determined within that bound.

#### Scenario: Extremely sparse thread

- GIVEN a long channel with a sparse thread
- WHEN an agent pages that thread
- THEN the plugin fetches no more than a fixed bounded number of posts
- AND returns an explicit limit error if it cannot resolve the page

### Requirement: Reauthorization of automatic replies

Automatic Twinmate replies MUST recheck the owner is active, is a member
of the direct message and can still write immediately before publication.

#### Scenario: Owner removed while AI is thinking

- GIVEN an automatic reply is in flight
- WHEN the owner's account is deactivated or DM membership revoked
- THEN the plugin does not publish a reply as that owner
