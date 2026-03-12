.PHONY: help bootstrap test proto lint format docs

help:
	@echo "Available targets:"
	@echo "  bootstrap - verify basic repo structure"
	@echo "  test      - reserved for future Go test entrypoint"
	@echo "  proto     - reserved for future protocol generation"
	@echo "  lint      - reserved for future lint entrypoint"
	@echo "  format    - reserved for future format entrypoint"
	@echo "  docs      - show current documentation entrypoints"

bootstrap:
	@echo "Repository scaffold is in place."
	@echo "Read README.md and docs/plans/current.md to get started."

test:
	go test ./...

proto:
	@echo "Protocol generation pipeline is not configured yet. Define proto sources under api/proto first."

lint:
	@echo "Lint pipeline placeholder. Wire golangci-lint or equivalent when modules are added."

format:
	gofmt -w ./pkg ./services

docs:
	@echo "Current plan: docs/plans/current.md"
	@echo "Roadmap:      docs/plans/v1/roadmap.md"
	@echo "Architecture: docs/architecture/overview.md"
	@echo "Constraints:  docs/memory/constraints.md"
