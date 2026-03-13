# 084 Ops Repo Reader Tests

## Goal

Add repo-level unit coverage for the new `ops` durable readers so MySQL and Redis status reads are validated at their storage boundary.

## Scope

- add `sqlmock` coverage for the ops MySQL bootstrap reader
- add `miniredis` coverage for the ops Redis runtime reader

## Non-Goals

- new operator endpoints
- integration-only coverage

## Acceptance

- both ops repo readers have direct unit tests
- `go test ./...` remains green
