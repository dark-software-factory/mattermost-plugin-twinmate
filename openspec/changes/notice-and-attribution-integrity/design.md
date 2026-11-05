## Boundaries and Decisions

1. A successful `/twinmate on` first creates the visible notice, then stores
   an opt-in record containing its post ID. A failed create or KV write leaves
   forwarding off. Repeated `/on` with an intact notice adds no duplicate.
2. The Go server re-reads the notice with `GetPost` whenever `DMEnabled` is
   checked, including before delivery and again before reply publication.
   Missing, deleted, edited, mismatched-user or mismatched-channel notices
   disable forwarding. `off` clears the opt-in state. Old `"on"` values fail
   closed and require a new notice.
3. The update hook treats the persisted Twinmate body marker as a mandatory
   marker on an existing post regardless of KV availability. This fallback
   protects even a published post whose provenance write failed. It does NOT
   classify the post as trusted AI for badges or forwarding. A human who
   deliberately uses the exact reserved prefix also cannot edit that prefix
   away; they can still delete their own message.
4. Mattermost's `MessageWillBeUpdated` requires returning the updated post
   pointer on an allowed edit (`nil` rejects it). Keep all ordinary edits and
   label-preserving AI corrections working while rejecting label removal.

## Verification

- Unit tests simulate editing/deleting the notice after opt-in, missing IDs,
  a failed provenance KV write followed by attempted marker removal, and
  allowed human/AI message corrections.
- Playwright E2E verifies edits/deletions of the real Mattermost notice stop
  subsequent automatic callback events on both supported server versions.
- Run Go/vet, webapp, OpenSpec and Mattermost 11.0/11.11 CI before archive.
