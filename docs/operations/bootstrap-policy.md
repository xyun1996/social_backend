# Bootstrap Policy

## Current Rule

- Owned schema bootstrap is allowed only for local and early prototype environments.
- Bootstrap must be explicit when possible; use `BOOTSTRAP_ONLY=true` or the local bootstrap script instead of relying on long-lived service startup as the only initialization path.
- Repeated bootstrap runs must be safe. MySQL-backed services now record applied service-owned steps in `schema_migrations`, and the owned statements remain idempotent.

## Local MySQL Bootstrap

- Preferred command: `make bootstrap-local-mysql`
- Verification command: `make verify-local-mysql-migrations`
- Under the hood this runs `scripts/dev/bootstrap-local-mysql.ps1`
- The script first ensures `MYSQL_DATABASE` exists before invoking service-owned schema bootstrap
- Each MySQL-backed service records `(service_name, migration_id)` progress in `schema_migrations`
- Covered MySQL-backed services:
  - `identity`
  - `social`
  - `invite`
  - `chat`
  - `party`
  - `guild`

## Production Direction

- Production should move from local service-owned bootstrap to reviewed migration promotion.
- Service startup should validate dependencies, not silently redefine schema shape.
- Future migration tooling should preserve current service ownership boundaries rather than centralizing all schema logic into one opaque layer.
