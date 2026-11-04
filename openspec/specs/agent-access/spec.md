# Agent Access

## Purpose

Define how a developer-owned local agent reads permitted Mattermost messages
and posts results as its owner's visibly identified Twinmate, without requiring
a Mattermost personal access token or an inbound connection to the developer's
computer.

## Requirements

### Requirement: Personal agent key

Each Mattermost user MUST be able to create, rotate, and revoke their own
Twinmate agent key. Twinmate MUST associate the key with its issuing user,
store only a non-reversible digest of that key, and MUST NOT return the
plaintext key through a later status request.

#### Scenario: Rotate a previously issued key

- GIVEN a user has a working Twinmate agent key
- WHEN that user rotates the key in settings
- THEN the previous key no longer authenticates agent requests
- AND the new key authenticates as the issuing user only

#### Scenario: Revoked or unknown key

- GIVEN an agent presents a revoked or unknown `X-Twinmate-Token` header
- WHEN the agent requests a post or tries to publish a message
- THEN the plugin denies the request without returning message contents

### Requirement: Read a permitted message by reference

An authenticated agent MUST be able to retrieve a message using a Mattermost
post ID or permalink only when the key owner currently has permission to read
that conversation. Returned data MUST NOT expose other post properties or file
attachment contents.

#### Scenario: Read a message from an accessible channel

- GIVEN the key owner may read a channel
- WHEN the agent requests a known post in that channel by permalink
- THEN the API returns the message text and its post/channel/thread identifiers

#### Scenario: Attempt to read a private conversation

- GIVEN the key owner is not a member of a private channel or DM
- WHEN the agent requests a post from that conversation
- THEN the API denies access without disclosing the private message text

### Requirement: Explicit chronological history pagination

An authenticated agent MUST be able to request older or newer messages around
a reference post, with `direction=before|after`, `scope=channel|thread`, and
an optional page limit from 1 to 50. The response MUST provide chronological
messages plus a continuation cursor when more data is available; the plugin
MUST enforce the owner's read permission for the requested scope.

#### Scenario: Fetch surrounding messages on demand

- GIVEN a readable message in a channel
- WHEN the agent asks for seven earlier messages in that message's thread
- THEN the API returns at most seven earlier messages from that thread
- AND the response indicates whether another page can be requested

#### Scenario: Invalid pagination arguments

- GIVEN an authenticated agent
- WHEN it asks for an unsupported scope or a limit greater than 50
- THEN the API rejects the request rather than returning an unbounded history

### Requirement: Authorized send-as-owner with AI attribution

An authenticated agent MUST be able to publish a new message to an allowed
channel or reply to an allowed post as the issuing user, subject to current
Mattermost membership and write permissions. The API MUST NOT allow the agent
to choose another sender identity. Successful writes MUST use the AI attribution
contract defined in `ai-attribution`.

#### Scenario: Publish a result in a thread

- GIVEN the key owner may post in the target thread
- WHEN the agent supplies a message and `reply_to_post_id`
- THEN Mattermost creates a reply authored by that owner and labeled as AI-generated

#### Scenario: Attempt to impersonate another user

- GIVEN a key associated with user A
- WHEN an agent tries to publish into an unauthorized channel or as user B
- THEN the request is rejected or any supplied sender ID is ignored
- AND no message is published under user B's identity

### Requirement: Local MCP adapter

Twinmate MUST expose the same authorized read, page, and send operations
through a standalone Go stdio MCP client. The client MUST use outbound HTTPS
to the server with the personal Twinmate key and MUST NOT require an inbound
port on the agent's machine.

#### Scenario: Local developer agent retrieves context and publishes results

- GIVEN a locally configured MCP client with a valid Twinmate URL and key
- WHEN the client invokes `twinmate_get_post`, `twinmate_list_posts`, or `twinmate_send_message`
- THEN the corresponding Twinmate HTTPS API operation runs with owner permissions
