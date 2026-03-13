# 088 Add DB Helper Tests

## Goal

Add direct unit coverage for shared durable connection helpers where local test doubles are available.

## Scope

- add `OpenRedis` coverage in `pkg/db`
- keep helper behavior unchanged

## Non-Goals

- end-to-end database integration
- MySQL network integration tests in default `go test ./...`

## Acceptance

- shared Redis connection helper has direct unit coverage
- `go test ./...` remains green
