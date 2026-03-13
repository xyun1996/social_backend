# 067 Local Durable Run Targets

## Goal

Make the current durable-path services easy to start against the committed local MySQL and Redis defaults.

## Scope

- add `make` targets for MySQL-backed identity, social, invite, and chat
- add `make` targets for Redis-backed presence and worker
- document the local durable runflow

## Non-Goals

- container orchestration
- production deployment commands
- multi-service process supervision

## Acceptance

- operators can start each durable-capable service with a single `make` target
- the local durable runflow is documented under `docs/operations`
