## Implementation

- [ ] Introduce admin exact-host allowlist and safe per-dial checks, with tests.
- [ ] Add scoped agent keys, API enforcement and UI controls, with tests.
- [ ] Add durable DM notification and owner revalidation, with tests.
- [ ] Preserve provenance on edits and trust server state for message routing.
- [ ] Replace unbounded thread loading with budgeted pagination, with tests.
- [ ] Update E2E fixtures, docs and OpenSpec source-of-truth spec.

## Verification

- [ ] `make spec-check`
- [ ] `go test ./...` and `go vet ./...` (CI if local Go unavailable)
- [ ] `npm test --prefix webapp`
- [ ] E2E Mattermost 11.0/11.11 CI green
- [ ] Review security regression tests and archive change before merge
