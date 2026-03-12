# 001 Bootstrap Go Foundation

- Title: Bootstrap Go module and shared service runtime
- Status: `done`
- Version: `v1`
- Milestone: `01 Foundation`

## Background

The repository scaffold and governance documents are in place, but the code layer still needs a first executable task so future services do not start from empty directories.

## Goal

Initialize the root Go module, create shared process bootstrap packages, and add minimal runnable entrypoints for the first services.

## Scope

- Create the root `go.mod`
- Introduce shared config, logging, and service lifecycle packages under `pkg/`
- Add minimal `gateway` and `identity` binaries with health endpoints
- Align `go.mod` and `go.work` with the installed Go version baseline

## Non-Goals

- Implement business logic
- Define RPC contracts
- Add persistence, Redis, or real transport protocols

## Dependencies

- [01 Foundation milestone](../milestones/01-foundation.md)
- [Architecture overview](../../../architecture/overview.md)
- [ADR-001 Transport Strategy](../../../memory/adr/ADR-001-transport-strategy.md)
- [ADR-002 Session Granularity](../../../memory/adr/ADR-002-session-granularity.md)

## Acceptance Criteria

- `go test ./...` passes
- `gateway` and `identity` binaries can start and expose `/healthz`
- Shared bootstrap packages are reusable by later services

## Related Docs / ADRs

- [Current plan](../../current.md)
- [V1 roadmap](../roadmap.md)
