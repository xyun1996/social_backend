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
- `make check-local-durable-status`

## Verified Local Sequence

The following sequence was validated successfully on Windows on `2026-03-13`:

1. `make bootstrap-local-mysql`
2. `make verify-local-mysql-migrations`
3. `make run-ops-durable`
4. `make check-local-durable-status`
5. `make test-local-durable`

Expected outcomes:

- `verify-local-mysql-migrations` prints the owned migration ids for `identity`, `social`, `invite`, `chat`, `party`, and `guild`
- `check-local-durable-status` passes once `ops` is running with durable readers enabled
- `test-local-durable` passes without requiring any manual schema work after bootstrap

## Notes

- The MySQL-backed targets explicitly enable `*_AUTO_MIGRATE=true`.
- `run-ops-durable` enables both `OPS_MYSQL_STATUS=true` and `OPS_REDIS_STATUS=true` so `ops` can inspect durable bootstrap and runtime state together.
- `check-local-durable-status` calls the `ops` durable summary endpoint, prints the current MySQL bootstrap and Redis runtime snapshots, and exits non-zero if the full local durable topology is not visible.
- The status gate verifies durable reader visibility and expected MySQL service ownership. It does not require Redis counters to be non-zero unless you intentionally start runtime-producing services such as `presence`, `worker`, or `gateway`.
- The owned MySQL bootstrap is now idempotent, so repeated local restarts do not fail just because tables already exist.
- `verify-local-mysql-migrations` reads `schema_migrations` and checks that every MySQL-backed service recorded its owned migration ids.
- `test-local-durable` runs the opt-in durable integration tests against local MySQL and Redis and leaves default `go test ./...` behavior unchanged.
- `bootstrap-local-mysql` runs each MySQL-backed service once in `BOOTSTRAP_ONLY=true` mode and then verifies `schema_migrations`, so schema initialization is an explicit step instead of a side effect of starting long-lived processes.
- The status script can still be relaxed for partial topologies by overriding `REQUIRE_MYSQL_SUMMARY`, `REQUIRE_REDIS_SUMMARY`, or `EXPECTED_MYSQL_SERVICES`.
- Use [Local durable bootstrap and status triage](local-durable-troubleshooting.md) when bootstrap or status checks fail.
- These targets are for local iteration only; production startup should not assume inline schema bootstrap.
- On Windows, the durable `make` targets now call PowerShell wrappers under `scripts/dev/`, so environment variables are applied consistently.
