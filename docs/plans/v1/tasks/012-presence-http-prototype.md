# 012 Presence HTTP Prototype

- Title: Add an in-memory presence prototype and gateway-facing reporting contract
- Status: `done`
- Version: `v1`
- Milestone: `02 Identity Session`

## Background

Chat, guild, and party flows now exist as runnable prototypes, but there is still no executable source of truth for online state and no explicit contract between gateway and presence.

## Goal

Add a local in-memory `presence` service prototype that exercises player connect, heartbeat, disconnect, and query flows while establishing presence as the authoritative online-state boundary.

## Scope

- Add presence domain models for online state and last-seen metadata
- Add in-memory service logic for connect, heartbeat, disconnect, and query
- Add HTTP endpoints and tests for reporting and lookup
- Define the first gateway-to-presence reporting contract
- Add a runnable `presence` service entrypoint

## Non-Goals

- Redis integration
- WebSocket or TCP push fanout
- Rich location graphs or friend presence subscriptions

## Dependencies

- [ADR-006 Presence Authority](../../../memory/adr/ADR-006-presence-authority.md)
- [004 Wire Gateway Session Introspection](004-wire-gateway-session-introspection.md)

## Acceptance Criteria

- Presence can report player online and offline state transitions
- Gateway-facing update semantics are explicit
- Query APIs expose enough state for downstream read flows
- `go test ./...` passes

## Related Docs / ADRs

- [Current plan](../../current.md)
- [Architecture overview](../../../architecture/overview.md)
