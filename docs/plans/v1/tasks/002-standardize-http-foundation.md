# 002 Standardize HTTP Foundation

- Title: Standardize shared HTTP response and error handling
- Status: `done`
- Version: `v1`
- Milestone: `01 Foundation`

## Background

The initial `gateway` and `identity` binaries can run, but they still need a shared response shape and test-covered transport helpers before feature endpoints are added.

## Goal

Introduce a reusable JSON response layer and transport-safe error model, then align the starter services to use it.

## Scope

- Add a shared application error type
- Add JSON response helpers under shared transport code
- Convert health endpoints to use the shared response format
- Add tests for config loading and HTTP response helpers

## Non-Goals

- Domain-specific validation errors
- Authentication middleware
- RPC or TCP protocol work

## Dependencies

- [001 Bootstrap Go Foundation](001-bootstrap-go-foundation.md)
- [01 Foundation milestone](../milestones/01-foundation.md)
- [Architecture protocols](../../../architecture/protocols.md)

## Acceptance Criteria

- Shared helpers produce consistent JSON output
- `gateway` and `identity` health responses use the shared helper
- `go test ./...` passes with unit coverage for the new foundation packages

## Related Docs / ADRs

- [Current plan](../../current.md)
- [V1 roadmap](../roadmap.md)
