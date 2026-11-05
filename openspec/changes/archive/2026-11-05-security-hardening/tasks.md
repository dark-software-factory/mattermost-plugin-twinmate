## Implementation

- [x] Introduce admin exact-host allowlist and safe per-dial checks, with tests.
- [x] Add scoped agent keys, API enforcement and UI controls, with tests.
- [x] Add durable DM notification and owner revalidation, with tests.
- [x] Preserve provenance on edits and trust server state for message routing.
- [x] Replace unbounded thread loading with budgeted pagination, with tests.
- [x] Update E2E fixtures, docs and OpenSpec source-of-truth spec.

## Verification

- [x] `make spec-check`
- [x] `go test ./...` and `go vet ./...` (CI if local Go unavailable)
- [x] `npm test --prefix webapp`
- [x] E2E Mattermost 11.0/11.11 CI green
- [x] Review security regression tests and archive change before merge
