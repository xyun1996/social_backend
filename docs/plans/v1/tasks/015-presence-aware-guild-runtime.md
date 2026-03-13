# 015 Presence-Aware Guild Runtime

- Title: Add presence-aware member state views to the guild prototype
- Status: `done`
- Version: `v1`
- Milestone: `05 Guild`

## Background

Guild now supports shared invites and membership, but still has no runtime-aware member view tied to the system presence authority.

## Goal

Add a presence-aware member-state API to the in-memory `guild` prototype so guild consumers can inspect role and online state together.

## Scope

- Add a presence client boundary for guild
- Add a member-state API combining guild role and presence
- Update local run configuration and HTTP contract docs

## Non-Goals

- Guild governance workflows beyond owner/member
- Presence-triggered notifications
- Audit or moderation actions

## Dependencies

- [012 Presence HTTP Prototype](012-presence-http-prototype.md)
- [ADR-006 Presence Authority](../../../memory/adr/ADR-006-presence-authority.md)

## Acceptance Criteria

- Guild member views expose presence-aware runtime state
- `go test ./...` passes

## Related Docs / ADRs

- [Current plan](../../current.md)
- [Guild HTTP contract](../../../../api/http/guild.md)
