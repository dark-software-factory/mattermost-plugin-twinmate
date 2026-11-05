## Why

The unreleased Twinmate plugin makes privileged reads and send-as-owner writes
available to developer agents. Security review found a global private-network
callback toggle, excessive token privileges, silent transmission of DM text to
third parties, editable AI attribution and unbounded thread-history reads.

## What Changes

- Require an explicit administrator allowlist for private callback hosts.
- Issue scoped agent keys with read/send permissions and optional channel IDs;
  default new keys to read-only, with existing legacy keys retaining their
  permissions until rotation.
- Disclose automatic DM forwarding in the conversation and validate the
  recipient's active membership/permissions before external and post callbacks.
- Reject attempts to remove AI attribution from generated posts and prevent
  forged prefixes from suppressing ordinary user-message processing.
- Bound thread pagination work, failing clearly when a thread is too sparse
  to page within a fixed server-side budget.

## Compatibility, Risks and Rollback

Token issuance without options becomes read-only; callers requiring send must
request `allow_send: true`. Existing tokens keep working until rotated.
Private network callbacks now require **both** the existing admin toggle and
an exact admin host:port allowlist entry. Existing private endpoints must be
explicitly approved before they resume working.

Automatic DM opt-in now produces a visible channel notice that discloses
external processing to the other participant. No existing data is deleted.
Rollback is the prior plugin bundle and settings; already-posted disclosures
remain in Mattermost history.

Not in this change: a new enterprise secret store for callback credentials,
durable queues or migration of plaintext callback secrets; these remain
deployment risks documented in the README.
