# 089 Ops Durable Summary Endpoint

## Goal

Add a single operator endpoint that summarizes the currently enabled durable status readers instead of requiring multiple calls.

## Scope

- add `GET /v1/ops/durable/summary`
- aggregate optional MySQL bootstrap and Redis runtime snapshots
- update local status check script to use the summary endpoint
- cover the endpoint in unit and durable integration tests

## Non-Goals

- changing existing durable status endpoints
- adding write operations

## Acceptance

- `ops` exposes one durable summary endpoint
- local durable status script uses the summary endpoint
- `go test ./...` remains green
