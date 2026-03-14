.PHONY: help bootstrap test proto proto-check proto-lint check-contracts check-dev lint format docs run-api-gateway run-social-core run-ops-worker run-gateway run-identity run-social run-invite run-chat run-party run-guild run-presence run-ops run-worker run-identity-mysql run-social-mysql run-invite-mysql run-chat-mysql run-party-mysql run-guild-mysql run-presence-redis run-worker-redis run-gateway-redis run-ops-durable test-local-durable bootstrap-local-mysql verify-local-mysql-migrations check-local-durable-status release-dry-run load-hot-paths fault-drill

help:
	@echo "Available targets:"
	@echo "  bootstrap - verify basic repo structure"
	@echo "  test      - reserved for future Go test entrypoint"
	@echo "  proto     - lint and generate Go bindings from api/proto via buf"
	@echo "  proto-check - run lightweight proto smoke tests"
	@echo "  proto-lint - run proto smoke tests and require buf lint"
	@echo "  check-contracts - print and validate current HTTP/proto/TCP contract inventory"
	@echo "  check-dev - run go test, proto checks, and contract inventory checks"
	@echo "  lint      - reserved for future lint entrypoint"
	@echo "  format    - reserved for future format entrypoint"
	@echo "  docs      - show current documentation entrypoints"
	@echo "  run-api-gateway - start the rebuilt client ingress runtime"
	@echo "  run-social-core - start the rebuilt core social runtime"
	@echo "  run-ops-worker  - start the rebuilt ops/worker runtime"
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
	@echo "  run-identity-mysql - start identity against local MySQL"
	@echo "  run-social-mysql   - start social against local MySQL"
	@echo "  run-invite-mysql   - start invite against local MySQL"
	@echo "  run-chat-mysql     - start chat against local MySQL"
	@echo "  run-party-mysql    - start party against local MySQL"
	@echo "  run-guild-mysql    - start guild against local MySQL"
	@echo "  run-presence-redis - start presence against local Redis"
	@echo "  run-worker-redis   - start worker against local Redis"
	@echo "  run-gateway-redis  - start gateway against local Redis"
	@echo "  run-ops-durable    - start ops with MySQL and Redis status readers enabled"
	@echo "  test-local-durable - run opt-in durable integration tests against local MySQL and Redis"
	@echo "  bootstrap-local-mysql - bootstrap owned MySQL schemas without serving traffic"
	@echo "  verify-local-mysql-migrations - inspect required schema_migrations rows on local MySQL"
	@echo "  check-local-durable-status - query ops for local MySQL and Redis durable status"
	@echo "  release-dry-run - run the single-region release checklist without deploying"
	@echo "  load-hot-paths - execute benchmark-style hot path smoke load tests"
	@echo "  fault-drill - run local dependency fault drill helper"

bootstrap:
	@echo "Repository scaffold is in place."
	@echo "Read README.md and docs/plans/current.md to get started."

test:
	go test ./...

proto:
	powershell -ExecutionPolicy Bypass -File ./scripts/dev/proto-generate.ps1

proto-check:
	powershell -ExecutionPolicy Bypass -File ./scripts/dev/proto-check.ps1

proto-lint:
	powershell -ExecutionPolicy Bypass -File ./scripts/dev/proto-check.ps1 -RequireBuf -RunBufLint

check-contracts:
	go run ./scripts/dev/cmd/check_contract_inventory

check-dev:
	powershell -ExecutionPolicy Bypass -File ./scripts/dev/dev-check.ps1

lint:
	@echo "Lint pipeline placeholder. Wire golangci-lint or equivalent when modules are added."

format:
	gofmt -w ./pkg ./services

docs:
	@echo "Current plan: docs/plans/current.md"
	@echo "Roadmap:      docs/plans/product/roadmap.md"
	@echo "Architecture: docs/architecture/overview.md"
	@echo "Constraints:  docs/memory/constraints.md"

run-api-gateway:
	go run ./services/api-gateway/cmd/api-gateway

run-social-core:
	go run ./services/social-core/cmd/social-core

run-ops-worker:
	go run ./services/ops-worker/cmd/ops-worker

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

run-identity-mysql:
	powershell -ExecutionPolicy Bypass -File ./scripts/dev/run-identity-mysql.ps1

run-social-mysql:
	powershell -ExecutionPolicy Bypass -File ./scripts/dev/run-social-mysql.ps1

run-invite-mysql:
	powershell -ExecutionPolicy Bypass -File ./scripts/dev/run-invite-mysql.ps1

run-chat-mysql:
	powershell -ExecutionPolicy Bypass -File ./scripts/dev/run-chat-mysql.ps1

run-party-mysql:
	powershell -ExecutionPolicy Bypass -File ./scripts/dev/run-party-mysql.ps1

run-guild-mysql:
	powershell -ExecutionPolicy Bypass -File ./scripts/dev/run-guild-mysql.ps1

run-presence-redis:
	powershell -ExecutionPolicy Bypass -File ./scripts/dev/run-presence-redis.ps1

run-worker-redis:
	powershell -ExecutionPolicy Bypass -File ./scripts/dev/run-worker-redis.ps1

run-gateway-redis:
	powershell -ExecutionPolicy Bypass -File ./scripts/dev/run-gateway-redis.ps1

run-ops-durable:
	powershell -ExecutionPolicy Bypass -File ./scripts/dev/run-ops-durable.ps1

test-local-durable:
	powershell -ExecutionPolicy Bypass -File ./scripts/dev/test-local-durable.ps1

bootstrap-local-mysql:
	powershell -ExecutionPolicy Bypass -File ./scripts/dev/bootstrap-local-mysql.ps1

verify-local-mysql-migrations:
	powershell -ExecutionPolicy Bypass -File ./scripts/dev/verify-local-mysql-migrations.ps1

check-local-durable-status:
	powershell -ExecutionPolicy Bypass -File ./scripts/dev/check-local-durable-status.ps1

release-dry-run:
	powershell -ExecutionPolicy Bypass -File ./scripts/release/release-dry-run.ps1

load-hot-paths:
	powershell -ExecutionPolicy Bypass -File ./scripts/dev/load-hot-paths.ps1

fault-drill:
	powershell -ExecutionPolicy Bypass -File ./scripts/dev/fault-drill.ps1
