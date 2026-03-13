# 016 Proto Contract Baseline

- Title: Add the first internal proto contract baseline for shared service boundaries
- Status: `done`
- Version: `v1`
- Milestone: `01 Foundation`

## Background

HTTP contracts are now documented and several services already call each other over explicit boundaries, but `api/proto/` is still empty.

## Goal

Add the first proto design baseline for internal service-to-service contracts so future gRPC work starts from explicit service contracts instead of ad hoc client shapes.

## Scope

- Add `api/proto/README.md`
- Add first proto contracts for `identity`, `presence`, and `invite`
- Align proto files with the currently documented HTTP boundary semantics

## Non-Goals

- Code generation
- gRPC server implementation
- Full proto coverage for every domain

## Dependencies

- [ADR-005 Contract-First Boundaries](../../../memory/adr/ADR-005-contract-first-boundaries.md)
- [011 Document HTTP Contract Baseline](011-document-http-contract-baseline.md)

## Acceptance Criteria

- `api/proto/` is no longer empty
- Shared internal boundaries have a first proto baseline
- `go test ./...` passes

## Related Docs / ADRs

- [Current plan](../../current.md)
- [Architecture protocols](../../../architecture/protocols.md)
