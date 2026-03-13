.PHONY: help bootstrap test proto lint format docs run-gateway run-identity run-social run-invite run-chat run-party run-guild run-presence run-ops run-worker

help:
	@echo "Available targets:"
	@echo "  bootstrap - verify basic repo structure"
	@echo "  test      - reserved for future Go test entrypoint"
	@echo "  proto     - reserved for future protocol generation"
	@echo "  lint      - reserved for future lint entrypoint"
	@echo "  format    - reserved for future format entrypoint"
	@echo "  docs      - show current documentation entrypoints"
	@echo "  run-gateway  - start the gateway starter service"
	@echo "  run-identity - start the identity starter service"
	@echo "  run-social   - start the social starter service"
	@echo "  run-invite   - start the invite starter service"
	@echo "  run-chat     - start the chat starter service"
	@echo "  run-party    - start the party starter service"
	@echo "  run-guild    - start the guild starter service"
	@echo "  run-presence - start the presence starter service"
	@echo "  run-ops      - start the ops starter service"
	@echo "  run-worker   - start the worker starter service"

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

run-gateway:
	go run ./services/gateway/cmd/gateway

run-identity:
	go run ./services/identity/cmd/identity

run-social:
	go run ./services/social/cmd/social

run-invite:
	go run ./services/invite/cmd/invite

run-chat:
	go run ./services/chat/cmd/chat

run-party:
	go run ./services/party/cmd/party

run-guild:
	go run ./services/guild/cmd/guild

run-presence:
	go run ./services/presence/cmd/presence

run-ops:
	go run ./services/ops/cmd/ops

run-worker:
	go run ./services/worker/cmd/worker
