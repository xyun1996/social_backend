# 040 Presence Redis Repo Integration

## Goal

Move `presence` from a repo/redis skeleton to an executable optional Redis-backed store while preserving the current in-memory default.

## Scope

- define a store interface inside `presence`
- wire the Redis repository into connect, heartbeat, disconnect, and get flows
- add runtime selection through `PRESENCE_STORE`
- keep the existing HTTP contract unchanged

## Non-Goals

- expiry sweeps or multi-session fanout
- presence aggregation across shards
- making Redis mandatory for local runs

## Acceptance

- `presence` still runs with in-memory state by default
- `presence` can construct a Redis-backed service when `PRESENCE_STORE=redis`
- service and handler tests still pass after the store injection refactor
