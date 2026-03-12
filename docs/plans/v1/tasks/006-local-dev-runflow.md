# 006 Local Dev Runflow

- Title: Add local run targets and config examples for starter services
- Status: `done`
- Version: `v1`
- Milestone: `01 Foundation`

## Background

The repository now has runnable `gateway`, `identity`, and `social` binaries, but the local developer workflow still depends on knowing package paths and environment variable names by memory.

## Goal

Add standard local run targets and example environment files for the initial services.

## Scope

- Add `make` targets for running starter services
- Add example env files for `gateway`, `identity`, and `social`
- Update the README quick-start section to reference the new workflow

## Non-Goals

- Process orchestration
- Docker Compose
- Secret management

## Dependencies

- [001 Bootstrap Go Foundation](001-bootstrap-go-foundation.md)
- [005 Social Graph HTTP Prototype](005-social-graph-http-prototype.md)
- [01 Foundation milestone](../milestones/01-foundation.md)

## Acceptance Criteria

- The repository documents how to run the initial services locally
- Example environment variables exist for each starter service
- `go test ./...` still passes after the workflow additions

## Related Docs / ADRs

- [Current plan](../../current.md)
- [Architecture config](../../../architecture/config.md)
