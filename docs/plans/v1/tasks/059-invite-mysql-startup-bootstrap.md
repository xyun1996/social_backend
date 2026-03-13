# 059 Invite MySQL Startup Bootstrap

## Goal

Wire the `invite` service startup path to optionally use the MySQL-backed store so durable invite state can be enabled by configuration.

## Scope

- add `INVITE_STORE=mysql` startup selection
- add startup MySQL connectivity validation
- add optional `INVITE_AUTO_MIGRATE=true` owned schema bootstrap
- document local configuration defaults for the MySQL-backed mode

## Non-Goals

- changing invite lifecycle behavior
- redesigning invite expiry scheduling
- adding migration tooling beyond owned schema bootstrap

## Acceptance

- `invite` can still boot in memory mode without MySQL
- `invite` can boot against MySQL when configured and fails fast on connectivity errors
- the owned schema can be bootstrapped at startup when explicitly enabled
