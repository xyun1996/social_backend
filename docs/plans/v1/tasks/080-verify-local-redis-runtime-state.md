# 080 Verify Local Redis Runtime State

## Status

`done`

## Goal

Extend local durable verification so Redis-backed runtime state survives service restarts for all current Redis-backed services.

## Scope

- add local durable integration coverage for `presence(redis)` restart persistence
- add local durable integration coverage for `worker(redis)` queue persistence
- update current plan references for the new verification slice

## Non-Goals

- Redis schema or migration tooling
- production observability stacks

## Acceptance

- local durable tests prove `presence` snapshots survive restart on Redis
- local durable tests prove queued `worker` jobs survive restart on Redis
- default `go test ./...` remains green

## Completion Notes

- `make test-local-durable` passes and covers Redis-backed restart persistence for both `presence` and `worker`
- `make check-local-durable-status` confirms the Redis runtime reader is visible through `ops`
