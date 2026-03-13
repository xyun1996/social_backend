# 018 Worker HTTP Prototype

- Title: Add an in-memory worker prototype for async jobs and compensation retries
- Status: `done`
- Version: `v1`
- Milestone: `01 Foundation`

## Background

Every core runtime service now has a prototype, but `worker` is still an empty shell even though operations docs already assume worker retry backlogs and recovery workflows will exist.

## Goal

Add a local in-memory `worker` prototype that exercises enqueue, claim, complete, fail, and retry-oriented job flows.

## Scope

- Add worker job domain model
- Add in-memory worker queue logic
- Add HTTP endpoints and tests for job lifecycle operations
- Add a runnable `worker` service entrypoint
- Update local run configuration and HTTP contract docs

## Non-Goals

- Durable job persistence
- Cron scheduling
- Domain-specific job execution handlers
- Distributed locking

## Dependencies

- [Current plan](../../current.md)
- [Runbooks](../../../operations/runbooks.md)

## Acceptance Criteria

- Jobs can be enqueued, claimed, completed, and failed
- Failed jobs can be retried through claim again
- `go test ./...` passes

## Related Docs / ADRs

- [Architecture overview](../../../architecture/overview.md)
- [HTTP contracts](../../../../api/http/README.md)
