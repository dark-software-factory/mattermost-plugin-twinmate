## Why

Security review of the unreleased Twinmate plugin found two failure paths:
owners can edit/delete the DM disclosure while forwarding remains enabled, and
a generated post whose provenance KV write fails can later lose its AI marker.
Both defeat assurances given to the other participant or readers.

## What Changes

- Bind each DM opt-in to the exact persisted notice post. Before forwarding,
  confirm that post still exists, belongs to the owner and DM, and has the
  original disclosure text; otherwise fail closed until a new opt-in.
- Prevent removing the existing AI label from a generated post even when its
  provenance KV record is missing. Continue to use KV provenance, not labels,
  for verified badges and loop suppression.

## Compatibility, Risks and Rollback

The plugin is unreleased. Old plain `"on"` DM state is treated as disabled so
unverifiable notices cannot authorize forwarding; `/twinmate on` re-enables
with a fresh notice. No messages are deleted. Additional GetPost reads on
active DM events are accepted to enforce disclosure integrity. Rollback to
the previous plugin loses this protection; organizations should disable
automatic forwarding if they cannot accept that risk.

Not in scope: mutual consent, immutable administrative audit records, key
expiry, outbound secret encryption or delivery guarantees.
