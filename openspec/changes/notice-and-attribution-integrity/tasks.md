## Implementation

- [ ] Write failing Go tests for edited/deleted disclosure and failed-provenance edit.
- [ ] Persist and recheck DM notice ID/text/user/channel before forwarding.
- [ ] Harden attribution edit hook without treating text as verified provenance.
- [ ] Extend Mattermost E2E with edited/deleted notice regressions.

## Verification

- [ ] `make spec-check`
- [ ] `go test ./...` and `go vet ./...` in CI
- [ ] `npm test --prefix webapp`
- [ ] Mattermost 11.0 and 11.11 E2E green
- [ ] Archive OpenSpec change and revalidate current specs before merge
