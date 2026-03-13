# Local Durable Troubleshooting

- Title: Local durable bootstrap and status triage
- Owner: backend/platform

## Trigger

- `make bootstrap-local-mysql` fails
- `make verify-local-mysql-migrations` reports missing services or migration ids
- `make run-ops-durable` starts but `make check-local-durable-status` exits non-zero
- durable integration tests fail against local MySQL or Redis

## Checks

- Confirm local defaults are still the expected ones:
  - MySQL `localhost:3306`, `root / 1234`, database `social_backend`
  - Redis `localhost:6379`, no username, no password, database `0`
- Confirm MySQL is reachable:
  - `make verify-local-mysql-migrations`
- Confirm Redis is reachable:
  - start one Redis-backed service such as `make run-presence-redis`
- Confirm `ops` durable readers are enabled:
  - `make run-ops-durable`
- Confirm the durable summary is visible:
  - `make check-local-durable-status`

## Steps

1. Bootstrap the MySQL-owned schemas explicitly:
   - `make bootstrap-local-mysql`
2. Verify the expected MySQL service set is recorded:
   - `make verify-local-mysql-migrations`
   - expected services are `identity,social,invite,chat,party,guild`
3. Start the durable-backed runtime services:
   - MySQL-backed: `identity`, `social`, `invite`, `chat`, `party`, `guild`
   - Redis-backed: `presence`, `worker`, `gateway`
   - operator reader: `ops`
4. Run the status gate:
   - `make check-local-durable-status`
5. If only a partial topology is running, relax the gate explicitly instead of editing the script:
   - `set REQUIRE_MYSQL_SUMMARY=false`
   - `set REQUIRE_REDIS_SUMMARY=false`
   - `set EXPECTED_MYSQL_SERVICES=identity,invite`
6. Re-run the relevant durable integration tests after recovery:
   - `make test-local-durable`

## Rollback / Recovery

- If MySQL bootstrap drift is the problem:
  - stop all services
  - rerun `make bootstrap-local-mysql`
  - rerun `make verify-local-mysql-migrations`
- If Redis runtime visibility is the problem:
  - restart `presence`, `worker`, `gateway`, then `ops`
  - rerun `make check-local-durable-status`
- If only one service is misconfigured:
  - check its `*_STORE` setting first
  - then check `*_AUTO_MIGRATE=true` for MySQL-backed services
  - then check Redis/MySQL host defaults
- If the status gate is correct but too strict for the current session:
  - override `EXPECTED_MYSQL_SERVICES` to the subset you intentionally started

## Windows Notes

- Durable `make` targets now route through PowerShell wrappers in `scripts/dev/`, so you should prefer `make ...` over copying the raw environment-variable chains manually.
- If a durable `make` target still behaves differently from the equivalent PowerShell script, pull the latest `main` first and confirm the wrapper script exists for that target.

## Signals to Watch

- `mysql summary is required but missing`
  - `ops` was started without `OPS_MYSQL_STATUS=true`, or MySQL connectivity failed during startup
- `redis summary is required but missing`
  - `ops` was started without `OPS_REDIS_STATUS=true`, or Redis connectivity failed during startup
- `mysql summary is missing expected services: ...`
  - one or more MySQL-backed services did not record migrations in `schema_migrations`
- `redis: ... maint_notifications ... unknown subcommand`
  - local Redis does not support that optional client capability; the client falls back automatically and this warning is non-fatal
- `Durable summary` shows Redis with zero counters
  - this is valid when `ops` can read Redis but no runtime-producing Redis-backed services are currently active
