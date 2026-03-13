# 017 Ops HTTP Read Prototype

- Title: Add an operator-facing read prototype for presence, party, and guild snapshots
- Status: `done`
- Version: `v1`
- Milestone: `01 Foundation`

## Background

The repository now has multiple runtime-aware domain services, but no operator-facing query surface to verify cross-service state from outside individual domain APIs.

## Goal

Add a local in-memory `ops` read prototype that aggregates operator-facing presence, party, and guild snapshots through explicit service boundaries.

## Scope

- Add the `ops` service entrypoint
- Add presence, party, and guild HTTP clients for operator reads
- Add HTTP endpoints for player presence, party snapshot, and guild snapshot
- Update local run configuration and docs

## Non-Goals

- Operator write actions
- Moderation workflows
- Audit log persistence

## Dependencies

- [012 Presence HTTP Prototype](012-presence-http-prototype.md)
- [014 Presence-Aware Party Runtime](014-presence-aware-party-runtime.md)
- [015 Presence-Aware Guild Runtime](015-presence-aware-guild-runtime.md)

## Acceptance Criteria

- Ops can query player presence through the presence boundary
- Ops can query party and guild runtime snapshots through domain boundaries
- `go test ./...` passes

## Related Docs / ADRs

- [Current plan](../../current.md)
- [Observability](../../../operations/observability.md)
