# 062 Social MySQL Startup Bootstrap

## Goal

Wire the `social` service startup path to optionally use the MySQL-backed stores so durable friendship and block state can be enabled by configuration.

## Scope

- add `SOCIAL_STORE=mysql` startup selection
- add startup MySQL connectivity validation
- add optional `SOCIAL_AUTO_MIGRATE=true` owned schema bootstrap
- document local configuration defaults for the MySQL-backed mode

## Non-Goals

- changing social graph semantics
- making MySQL the default runtime mode
- adding migration tooling beyond owned schema bootstrap

## Acceptance

- `social` can still boot in memory mode without MySQL
- `social` can boot against MySQL when configured and fails fast on connectivity errors
- the owned schema can be bootstrapped at startup when explicitly enabled
