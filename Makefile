.PHONY: help bootstrap test proto proto-check check-contracts lint format docs run-gateway run-identity run-social run-invite run-chat run-party run-guild run-presence run-ops run-worker run-identity-mysql run-social-mysql run-invite-mysql run-chat-mysql run-party-mysql run-guild-mysql run-presence-redis run-worker-redis run-gateway-redis run-ops-durable test-local-durable bootstrap-local-mysql verify-local-mysql-migrations check-local-durable-status

help:
	@echo "Available targets:"
	@echo "  bootstrap - verify basic repo structure"
	@echo "  test      - reserved for future Go test entrypoint"
	@echo "  proto     - lint and generate Go bindings from api/proto via buf"
	@echo "  proto-check - run proto smoke tests and buf lint when available"
	@echo "  check-contracts - print and validate current HTTP/proto/TCP contract inventory"
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

bootstrap:
	@echo "Repository scaffold is in place."
	@echo "Read README.md and docs/plans/current.md to get started."

test:
	go test ./...

proto:
	powershell -ExecutionPolicy Bypass -File ./scripts/dev/proto-generate.ps1

proto-check:
	powershell -ExecutionPolicy Bypass -File ./scripts/dev/proto-check.ps1

check-contracts:
	go run ./scripts/dev/cmd/check_contract_inventory

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

run-identity-mysql:
	set APP_ENV=local && set IDENTITY_STORE=mysql && set IDENTITY_AUTO_MIGRATE=true && set MYSQL_HOST=localhost && set MYSQL_PORT=3306 && set MYSQL_USER=root && set MYSQL_PASSWORD=1234 && set MYSQL_DATABASE=social_backend && go run ./services/identity/cmd/identity

run-social-mysql:
	set APP_ENV=local && set SOCIAL_STORE=mysql && set SOCIAL_AUTO_MIGRATE=true && set MYSQL_HOST=localhost && set MYSQL_PORT=3306 && set MYSQL_USER=root && set MYSQL_PASSWORD=1234 && set MYSQL_DATABASE=social_backend && go run ./services/social/cmd/social

run-invite-mysql:
	set APP_ENV=local && set INVITE_STORE=mysql && set INVITE_AUTO_MIGRATE=true && set MYSQL_HOST=localhost && set MYSQL_PORT=3306 && set MYSQL_USER=root && set MYSQL_PASSWORD=1234 && set MYSQL_DATABASE=social_backend && go run ./services/invite/cmd/invite

run-chat-mysql:
	set APP_ENV=local && set CHAT_STORE=mysql && set CHAT_AUTO_MIGRATE=true && set MYSQL_HOST=localhost && set MYSQL_PORT=3306 && set MYSQL_USER=root && set MYSQL_PASSWORD=1234 && set MYSQL_DATABASE=social_backend && go run ./services/chat/cmd/chat

run-party-mysql:
	set APP_ENV=local && set PARTY_STORE=mysql && set PARTY_AUTO_MIGRATE=true && set MYSQL_HOST=localhost && set MYSQL_PORT=3306 && set MYSQL_USER=root && set MYSQL_PASSWORD=1234 && set MYSQL_DATABASE=social_backend && go run ./services/party/cmd/party

run-guild-mysql:
	set APP_ENV=local && set GUILD_STORE=mysql && set GUILD_AUTO_MIGRATE=true && set MYSQL_HOST=localhost && set MYSQL_PORT=3306 && set MYSQL_USER=root && set MYSQL_PASSWORD=1234 && set MYSQL_DATABASE=social_backend && go run ./services/guild/cmd/guild

run-presence-redis:
	set APP_ENV=local && set PRESENCE_STORE=redis && set REDIS_ADDR=localhost:6379 && set REDIS_USERNAME= && set REDIS_PASSWORD= && set REDIS_DB=0 && go run ./services/presence/cmd/presence

run-worker-redis:
	set APP_ENV=local && set WORKER_STORE=redis && set REDIS_ADDR=localhost:6379 && set REDIS_USERNAME= && set REDIS_PASSWORD= && set REDIS_DB=0 && go run ./services/worker/cmd/worker

run-gateway-redis:
	set APP_ENV=local && set GATEWAY_STORE=redis && set REDIS_ADDR=localhost:6379 && set REDIS_USERNAME= && set REDIS_PASSWORD= && set REDIS_DB=0 && go run ./services/gateway/cmd/gateway

run-ops-durable:
	set APP_ENV=local && set OPS_MYSQL_STATUS=true && set OPS_REDIS_STATUS=true && set MYSQL_HOST=localhost && set MYSQL_PORT=3306 && set MYSQL_USER=root && set MYSQL_PASSWORD=1234 && set MYSQL_DATABASE=social_backend && set REDIS_ADDR=localhost:6379 && set REDIS_USERNAME= && set REDIS_PASSWORD= && set REDIS_DB=0 && go run ./services/ops/cmd/ops

test-local-durable:
	set ENABLE_LOCAL_DURABLE_TESTS=true && set MYSQL_HOST=localhost && set MYSQL_PORT=3306 && set MYSQL_USER=root && set MYSQL_PASSWORD=1234 && set MYSQL_DATABASE=social_backend && set REDIS_ADDR=localhost:6379 && set REDIS_USERNAME= && set REDIS_PASSWORD= && go test ./services/integration -run TestLocalDurable -v

bootstrap-local-mysql:
	powershell -ExecutionPolicy Bypass -File ./scripts/dev/bootstrap-local-mysql.ps1

verify-local-mysql-migrations:
	set MYSQL_HOST=localhost && set MYSQL_PORT=3306 && set MYSQL_USER=root && set MYSQL_PASSWORD=1234 && set MYSQL_DATABASE=social_backend && go run ./scripts/dev/cmd/verify_mysql_migrations

check-local-durable-status:
	set OPS_BASE_URL=http://localhost:8088 && set REQUIRE_MYSQL_SUMMARY=true && set REQUIRE_REDIS_SUMMARY=true && set EXPECTED_MYSQL_SERVICES=identity,social,invite,chat,party,guild && go run ./scripts/dev/cmd/check_local_durable_status
