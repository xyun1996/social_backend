# 003 Identity HTTP Prototype

- Title: Add a local in-memory identity login and refresh prototype
- Status: `done`
- Version: `v1`
- Milestone: `02 Identity Session`

## Background

The repository now has a shared Go foundation, but the identity domain still lacks a concrete executable flow for account-to-player login and token refresh.

## Goal

Provide a local in-memory identity prototype that defines the first HTTP login and refresh flow without taking on MySQL, Redis, or full auth integration yet.

## Scope

- Add identity domain session models
- Add in-memory token issuance and refresh logic
- Add HTTP login and refresh endpoints
- Add unit tests for service and handler behavior

## Non-Goals

- Real credential verification
- Persistent token storage
- Gateway token validation
- WebSocket or TCP handshake logic

## Dependencies

- [001 Bootstrap Go Foundation](001-bootstrap-go-foundation.md)
- [002 Standardize HTTP Foundation](002-standardize-http-foundation.md)
- [02 Identity Session milestone](../milestones/02-identity-session.md)
- [ADR-002 Session Granularity](../../../memory/adr/ADR-002-session-granularity.md)

## Acceptance Criteria

- Identity service exposes local HTTP login and refresh endpoints
- Login returns account-bound and player-bound token metadata
- Refresh rotates refresh tokens and invalidates old ones
- `go test ./...` passes

## Related Docs / ADRs

- [Current plan](../../current.md)
- [V1 roadmap](../roadmap.md)
