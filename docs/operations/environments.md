# Environments

## Local

- Purpose: fast developer iteration
- Expected dependencies: local MySQL, Redis, optional tracing backend
- Tolerance: mock integrations acceptable
- Default local MySQL: `localhost:3306`, user `root`, password `1234`, database `social_backend`
- Default local Redis: `localhost:6379`, no username, no password, database `0`

## Dev

- Purpose: integration validation across services
- Expected dependencies: shared MySQL/Redis, basic observability

## Staging

- Purpose: release validation and protocol compatibility testing
- Expected dependencies: production-like topology where practical

## Prod

- Purpose: live game traffic
- Requirements: audited changes, stronger alerting, stable runbooks, and release notes

## Rules

- Document environment differences before relying on environment-specific behavior.
- Avoid introducing runtime assumptions that only exist in local development.
- Keep local-only credentials in committed example files only when they are explicitly disposable developer defaults.

## Runtime Toggles

- `IDENTITY_STORE`: `memory` or `mysql`
- `IDENTITY_AUTO_MIGRATE`: bootstrap owned MySQL schema on startup when `true`
- `INVITE_STORE`: `memory` or `mysql`
- `INVITE_AUTO_MIGRATE`: bootstrap owned MySQL schema on startup when `true`
- `CHAT_STORE`: `memory` or `mysql`
- `CHAT_AUTO_MIGRATE`: bootstrap owned MySQL schema on startup when `true`
- `SOCIAL_STORE`: `memory` or `mysql`
- `SOCIAL_AUTO_MIGRATE`: bootstrap owned MySQL schema on startup when `true`
- `PARTY_STORE`: `memory` or `mysql`
- `PARTY_AUTO_MIGRATE`: bootstrap owned MySQL schema on startup when `true`
- `GUILD_STORE`: `memory` or `mysql`
- `GUILD_AUTO_MIGRATE`: bootstrap owned MySQL schema on startup when `true`
- `PRESENCE_STORE`: `memory` or `redis`
- `WORKER_STORE`: `memory` or `redis`
- `WORKER_AUTO_RUN`: enable the worker background drain loop when `true`
- `PRESENCE_AUTO_PING`: validate Redis connectivity on startup when `true`
