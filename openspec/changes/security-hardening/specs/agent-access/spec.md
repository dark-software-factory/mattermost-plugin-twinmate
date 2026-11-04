## MODIFIED Requirements

### Requirement: Personal agent key

Each Mattermost user MUST be able to create, rotate, and revoke their own
Twinmate agent key. Twinmate MUST associate the key with its issuing user,
store only a non-reversible digest of that key, and MUST NOT return the
plaintext key through a later status request. Newly issued keys MUST be
read-only unless the user explicitly grants send-as-owner permission. The
owner MAY restrict the key to specified Mattermost channel IDs; these scopes
MUST be enforced on every operation in addition to Mattermost permissions.
Existing keys without scope metadata retain their former privileges until
rotated or revoked.

#### Scenario: Rotate a previously issued key

- GIVEN a user has a working Twinmate agent key
- WHEN that user rotates the key in settings
- THEN the previous key no longer authenticates agent requests
- AND the new key authenticates as the issuing user only

#### Scenario: Revoked or unknown key

- GIVEN an agent presents a revoked or unknown `X-Twinmate-Token` header
- WHEN the agent requests a post or tries to publish a message
- THEN the plugin denies the request without returning message contents

#### Scenario: New read-only or channel-scoped key

- GIVEN a user issues a read-only key limited to channel A
- WHEN the agent requests channel B or attempts to send any message
- THEN the plugin rejects the operation with HTTP 403

### Requirement: Explicit chronological history pagination

An authenticated agent MUST be able to request older or newer messages around
a reference post, with `direction=before|after`, `scope=channel|thread`, and
an optional page limit from 1 to 50. The response MUST provide chronological
messages plus a continuation cursor when more data is available; the plugin
MUST enforce the owner's read permission for the requested scope. Thread
pagination MUST NOT load unbounded threads; if it cannot determine a page
within the fixed upstream scan budget, it MUST return HTTP 422.

#### Scenario: Fetch surrounding messages on demand

- GIVEN a readable message in a channel
- WHEN the agent asks for seven earlier messages in that message's thread
- THEN the API returns at most seven earlier messages from that thread
- AND the response indicates whether another page can be requested

#### Scenario: Invalid pagination arguments

- GIVEN an authenticated agent
- WHEN it asks for an unsupported scope or a limit greater than 50
- THEN the API rejects the request rather than returning an unbounded history

#### Scenario: Sparse thread exceeds scan budget

- GIVEN a sparse thread in a busy channel
- WHEN its next page cannot be found by scanning at most 500 posts
- THEN the API returns a specific HTTP 422 error instead of claiming the thread ends

### Requirement: Authorized send-as-owner with AI attribution

An authenticated agent with an explicitly send-enabled key MUST be able to
publish a new message to an allowed channel or reply to an allowed post as
the issuing user, subject to current Mattermost membership, write permissions
and the key's channel scope. A read-only key MUST NOT publish. The API MUST
NOT allow the agent to choose another sender identity. Successful writes MUST
use the AI attribution contract defined in `ai-attribution`.

#### Scenario: Publish a result in a thread

- GIVEN a send-enabled key whose owner may post in the target thread
- WHEN the agent supplies a message and `reply_to_post_id`
- THEN Mattermost creates a reply authored by that owner and labeled as AI-generated

#### Scenario: Attempt to impersonate another user

- GIVEN a key associated with user A
- WHEN an agent tries to publish into an unauthorized channel or as user B
- THEN the request is rejected or any supplied sender ID is ignored
- AND no message is published under user B's identity
