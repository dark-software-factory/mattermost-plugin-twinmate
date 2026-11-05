## Implementation

- [x] Write failing Go tests for edited/deleted disclosure and failed-provenance edit.
- [x] Persist and recheck DM notice ID/text/user/channel before forwarding.
- [x] Harden attribution edit hook without treating text as verified provenance.
- [x] Preserve Mattermost's normal edit-hook contract for allowed post edits.
- [x] Extend Mattermost E2E with edited/deleted notice regressions.

## Verification

- [x] `make spec-check`
- [x] `go test ./...` and `go vet ./...` in CI
- [x] `npm test --prefix webapp`
- [x] Mattermost 11.0 and 11.11 E2E green
- [x] Archive OpenSpec change and revalidate current specs before merge
