# Bootstrap Policy

## Current Rule

- Owned schema bootstrap is allowed only for local and early prototype environments.
- Bootstrap must be explicit when possible; use `BOOTSTRAP_ONLY=true` or the local bootstrap script instead of relying on long-lived service startup as the only initialization path.
- Repeated bootstrap runs must be safe. Current MySQL-owned `CREATE TABLE` statements are idempotent.

## Local MySQL Bootstrap

- Preferred command: `make bootstrap-local-mysql`
- Under the hood this runs `scripts/dev/bootstrap-local-mysql.ps1`
- The script first ensures `MYSQL_DATABASE` exists before invoking service-owned schema bootstrap
- Covered MySQL-backed services:
  - `identity`
  - `social`
  - `invite`
  - `chat`
  - `party`
  - `guild`

## Production Direction

- Production should move from inline bootstrap to reviewed migrations.
- Service startup should validate dependencies, not silently redefine schema shape.
- Future migration tooling should preserve current service ownership boundaries rather than centralizing all schema logic into one opaque layer.
