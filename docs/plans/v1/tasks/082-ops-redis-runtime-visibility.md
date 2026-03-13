# 082 Ops Redis Runtime Visibility

## Goal

Expose Redis-backed runtime state through `ops` so local durable runtime health is visible alongside MySQL bootstrap state.

## Scope

- add an optional `ops` Redis runtime reader
- expose `GET /v1/ops/runtime/redis`
- include presence, gateway, and worker Redis state counts
- cover the new read surface in unit and durable integration tests

## Non-Goals

- Redis writes from `ops`
- full per-record dumps for runtime data

## Acceptance

- `ops` can summarize Redis-backed runtime state when Redis status reading is enabled
- the new read surface is documented in `api/http/ops.md`
- `go test ./...` remains green
