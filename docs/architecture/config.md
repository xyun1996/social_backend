# Config Strategy

## Directory Layout

- `configs/local/`
- `configs/dev/`
- `configs/staging/`
- `configs/prod/`
- `configs/examples/`

## Rules

- Commit templates and examples, not secrets.
- Prefer environment variables for secrets and environment-specific overrides.
- Document every config key that affects runtime behavior.

## Ownership

- Service-owned keys should be documented close to the service and summarized here.
- Shared infrastructure keys belong here first and may be referenced elsewhere.

## Current Local Defaults

The committed developer example for shared local infrastructure is `configs/examples/local-infra.env.example`.

Shared keys currently documented there:

- `MYSQL_HOST=localhost`
- `MYSQL_PORT=3306`
- `MYSQL_USER=root`
- `MYSQL_PASSWORD=1234`
- `MYSQL_DATABASE=social_backend`
- `REDIS_ADDR=localhost:6379`
- `REDIS_USERNAME=`
- `REDIS_PASSWORD=`
- `REDIS_DB=0`
- `IDENTITY_STORE=memory`
- `IDENTITY_AUTO_MIGRATE=false`
- `INVITE_STORE=memory`
- `INVITE_AUTO_MIGRATE=false`
- `CHAT_STORE=memory`
- `CHAT_AUTO_MIGRATE=false`
- `SOCIAL_STORE=memory`
- `SOCIAL_AUTO_MIGRATE=false`
- `PARTY_STORE=memory`
- `PARTY_AUTO_MIGRATE=false`
- `GUILD_STORE=memory`
- `GUILD_AUTO_MIGRATE=false`
- `PRESENCE_STORE=memory`
- `GATEWAY_STORE=memory`
- `WORKER_STORE=memory`
- `OPS_MYSQL_STATUS=false`
- `OPS_REDIS_STATUS=false`

## Future Work

- Decide on exact config format and loading library when Go modules are introduced.
- Shared MySQL foundation currently reads `MYSQL_HOST`, `MYSQL_PORT`, `MYSQL_USER`, `MYSQL_PASSWORD`, and `MYSQL_DATABASE`.
- Shared Redis foundation currently reads `REDIS_ADDR`, `REDIS_USERNAME`, `REDIS_PASSWORD`, and `REDIS_DB`.
- Service-local runtime selection currently uses `*_STORE` toggles, while owned schema bootstrap uses `*_AUTO_MIGRATE=true` for MySQL-backed services.
- MySQL-backed bootstrap records service-owned progress in `schema_migrations`, so repeated local bootstrap skips already applied migration ids.
- `ops` uses `OPS_MYSQL_STATUS` and `OPS_REDIS_STATUS` to opt into durable status readers without making MySQL or Redis mandatory for the default read path.
