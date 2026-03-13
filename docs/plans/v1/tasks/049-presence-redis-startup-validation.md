# 049 Presence Redis Startup Validation

## Goal

Tighten the optional Redis-backed `presence` startup path so the service fails fast on bad Redis configuration and the repo contract is explicitly tested.

## Scope

- ping Redis on startup when `PRESENCE_STORE=redis`
- add repository tests for canonical key shape and marshal roundtrip behavior
- update runflow notes

## Non-Goals

- Redis-based integration tests
- multi-key fanout or TTL policy changes
- changing the in-memory default path

## Acceptance

- `presence` fails fast if the configured Redis connection is unavailable
- repo tests cover key naming and marshal/unmarshal roundtrip behavior
- `go test ./services/presence/...` passes
