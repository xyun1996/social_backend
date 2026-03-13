# 020 Ops Worker Queue Visibility

- Title: Extend ops with worker queue visibility
- Status: `done`
- Version: `v1`
- Milestone: `01 Foundation`

## Background

Ops can now inspect runtime player and group state, and worker now has its own queue prototype, but worker backlog visibility is still isolated inside the worker service.

## Goal

Expose worker queue visibility through the `ops` service so operator reads can inspect async job backlog and filtered queue slices.

## Scope

- Add a worker client boundary for ops
- Add an ops endpoint for worker queue snapshots
- Update local run configuration and HTTP contract docs

## Non-Goals

- Worker control actions through ops
- Retry or replay orchestration
- Alerting implementation

## Dependencies

- [017 Ops HTTP Read Prototype](017-ops-http-read-prototype.md)
- [018 Worker HTTP Prototype](018-worker-http-prototype.md)

## Acceptance Criteria

- Ops can query worker job snapshots through the worker boundary
- Filtered worker visibility is exposed by status and type
- `go test ./...` passes

## Related Docs / ADRs

- [Current plan](../../current.md)
- [Observability](../../../operations/observability.md)
