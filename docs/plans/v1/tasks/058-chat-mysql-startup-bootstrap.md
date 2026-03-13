# 058 Chat MySQL Startup Bootstrap

## Goal

Wire the `chat` service startup path to optionally use the MySQL-backed stores so durable chat behavior can be enabled by configuration instead of further service rewrites.

## Scope

- add `CHAT_STORE=mysql` startup selection
- add startup MySQL connectivity validation
- add optional `CHAT_AUTO_MIGRATE=true` owned schema bootstrap
- document local configuration defaults for the MySQL-backed mode

## Non-Goals

- changing chat delivery semantics
- adding migration tooling beyond owned schema bootstrap
- making MySQL the default runtime mode

## Acceptance

- `chat` can still boot in memory mode without MySQL
- `chat` can boot against MySQL when configured and fails fast on connectivity errors
- the owned schema can be bootstrapped at startup when explicitly enabled
