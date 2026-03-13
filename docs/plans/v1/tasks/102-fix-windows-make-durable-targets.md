# 102 Fix Windows Make Durable Targets

- Title: Fix Windows make targets for durable local commands
- Status: `done`
- Version: `v1`
- Milestone: `01 Foundation`

## Background

GNU Make is now available locally, but Windows-style `set ... && go run ...` recipes do not reliably preserve environment variables under the current make shell behavior.

## Goal

Route durable make targets through PowerShell scripts so local MySQL and Redis commands behave consistently on Windows.

## Scope

- Add PowerShell wrappers for MySQL-backed and Redis-backed local run commands
- Add PowerShell wrappers for durable verification commands
- Update `Makefile` durable targets to call the wrappers

## Non-Goals

- Cross-platform shell abstraction beyond the current Windows-first workflow
- Changes to service runtime behavior

## Dependencies

- [067 Local Durable Run Targets](067-local-durable-run-targets.md)
- [085 Complete Local Durable Run Targets](085-complete-local-durable-run-targets.md)
- [100 Unified Dev Check Flow](100-unified-dev-check-flow.md)

## Acceptance Criteria

- `make verify-local-mysql-migrations` succeeds on Windows
- `make check-dev` remains green
- Durable run targets no longer depend on inline `set ... && ...` chains

## Related Docs / ADRs

- [Local durable runflow](../../../operations/local-durable-runflow.md)
