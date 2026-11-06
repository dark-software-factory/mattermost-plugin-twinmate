OPENSPEC_VERSION := 1.16.0
OPENSPEC := npm exec --yes --package=@fission-ai/openspec@$(OPENSPEC_VERSION) -- openspec

.PHONY: check spec-check

# Design-stage gate: intentionally does not compile or package an application.
check:
	@test -f README.md
	@test -f AGENTS.md
	@test -f docs/product/brief.md
	@test -f docs/architecture/README.md
	@test -f docs/research/README.md
	@test -f docs/decisions/README.md
	@test -f openspec/config.yaml
	@test -d openspec/specs
	@test ! -e go.mod
	@test ! -e plugin.json
	@test ! -d server
	@test ! -d internal
	@test ! -d webapp
	@test ! -d cmd
	@test ! -d e2e
	@echo "Design-only repository checks passed"

spec-check:
	OPENSPEC_TELEMETRY=0 OPENSPEC_NO_UPDATE_CHECK=1 $(OPENSPEC) validate --all --strict --no-interactive
	OPENSPEC_TELEMETRY=0 OPENSPEC_NO_UPDATE_CHECK=1 $(OPENSPEC) validate --archived --no-interactive
