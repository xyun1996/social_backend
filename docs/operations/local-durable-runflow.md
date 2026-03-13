# Local Durable Runflow

## Purpose

Start the services that already have optional durable backends against the local MySQL and Redis defaults committed in this repository.

## Local Defaults

- MySQL: `localhost:3306`, user `root`, password `1234`, database `social_backend`
- Redis: `localhost:6379`, no username, no password, database `0`

## Durable Startup Commands

- `make run-identity-mysql`
- `make run-social-mysql`
- `make run-invite-mysql`
- `make run-chat-mysql`
- `make run-party-mysql`
- `make run-guild-mysql`
- `make run-presence-redis`
- `make run-worker-redis`
- `make run-gateway-redis`
- `make run-ops-durable`
- `make test-local-durable`
- `make bootstrap-local-mysql`
- `make verify-local-mysql-migrations`

## Notes

- The MySQL-backed targets explicitly enable `*_AUTO_MIGRATE=true`.
- `run-ops-durable` enables both `OPS_MYSQL_STATUS=true` and `OPS_REDIS_STATUS=true` so `ops` can inspect durable bootstrap and runtime state together.
- The owned MySQL bootstrap is now idempotent, so repeated local restarts do not fail just because tables already exist.
- `verify-local-mysql-migrations` reads `schema_migrations` and checks that every MySQL-backed service recorded its owned migration ids.
- `test-local-durable` runs the opt-in durable integration tests against local MySQL and Redis and leaves default `go test ./...` behavior unchanged.
- `bootstrap-local-mysql` runs each MySQL-backed service once in `BOOTSTRAP_ONLY=true` mode and then verifies `schema_migrations`, so schema initialization is an explicit step instead of a side effect of starting long-lived processes.
- These targets are for local iteration only; production startup should not assume inline schema bootstrap.
