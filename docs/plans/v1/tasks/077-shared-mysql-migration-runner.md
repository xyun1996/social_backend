# 077 Shared MySQL Migration Runner

## Goal

Replace per-repository ad hoc bootstrap loops with a shared MySQL migration runner that records service-owned schema progress.

## Scope

- add a shared `pkg/db` migration runner for MySQL-backed services
- persist applied migration ids in a shared `schema_migrations` table
- add unit tests for applied, pending, and failing migration paths

## Non-Goals

- external migration tooling
- cross-service schema ownership changes
- Redis migration semantics

## Acceptance

- MySQL-backed services can apply service-owned migrations through one shared path
- repeated bootstrap runs skip already applied migration ids
- `go test ./...` remains green
