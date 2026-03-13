# 068 Local Durable Integration Tests

## Goal

Add opt-in integration coverage that exercises the current durable-path services against real local MySQL and Redis instances.

## Scope

- add opt-in integration tests for MySQL-backed chat, invite, and social flows
- add opt-in integration tests for Redis-backed presence and worker flows
- keep the tests skipped by default unless explicitly enabled

## Non-Goals

- making durable integration tests mandatory in every `go test ./...` run
- adding containerized test infrastructure
- covering every service in one end-to-end process graph

## Acceptance

- local durable integration tests run when `ENABLE_LOCAL_DURABLE_TESTS=true`
- tests use temporary MySQL databases and isolated Redis state
- default repository test flow remains fast and skip-safe
