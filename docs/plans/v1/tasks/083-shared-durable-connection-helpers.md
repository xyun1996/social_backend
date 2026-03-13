# 083 Shared Durable Connection Helpers

## Goal

Reduce repeated MySQL and Redis startup wiring in service entrypoints by introducing shared validated connection helpers.

## Scope

- add shared `pkg/db` helpers for opening and pinging MySQL and Redis
- refactor durable service `cmd/main.go` entrypoints to use the helpers
- keep service-local bootstrap and repo ownership unchanged

## Non-Goals

- dependency injection framework changes
- centralizing repo construction

## Acceptance

- durable-backed service entrypoints no longer duplicate raw MySQL and Redis open/ping code
- `go test ./...` remains green
