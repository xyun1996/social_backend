# 014 Presence-Aware Party Runtime

- Title: Add presence-aware ready and member state handling to the party prototype
- Status: `done`
- Version: `v1`
- Milestone: `06 Party Queue`

## Background

Party now reuses shared invites, but its ready and member-state behavior still ignores the system's presence authority.

## Goal

Add a presence-aware runtime layer to the in-memory `party` prototype so ready transitions and member inspection can reflect online state.

## Scope

- Add a presence client boundary for party
- Require online presence for ready-state updates
- Add a member-state API combining role, ready state, and presence
- Update local run configuration and HTTP contract docs

## Non-Goals

- Queue orchestration
- Matchmaker integration
- Friend presence subscriptions

## Dependencies

- [012 Presence HTTP Prototype](012-presence-http-prototype.md)
- [ADR-006 Presence Authority](../../../memory/adr/ADR-006-presence-authority.md)

## Acceptance Criteria

- Ready transitions require online presence
- Party member views expose presence-aware runtime state
- `go test ./...` passes

## Related Docs / ADRs

- [Current plan](../../current.md)
- [Party HTTP contract](../../../../api/http/party.md)
