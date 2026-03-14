# 119 Chat Resource Membership Permission Alignment

- Title: Align guild and party chat permissions with current resource membership
- Status: `done`
- Version: `v1`
- Milestone: `04 Chat Offline`

## Background

Chat already modeled `guild` and `party` as resource-backed channels, but read and send permissions still depended only on the locally stored member list. That left a drift risk whenever party or guild membership changed after channel creation.

## Goal

Use current guild and party membership to validate access to resource-backed chat channels without redesigning the public chat contract surface.

## Scope

- add optional guild and party membership readers to chat
- enforce current membership for guild and party channel reads and sends
- filter delivery planning to currently authorized members
- wire local HTTP clients from chat to guild and party membership reads
- add service tests for stale membership rejection

## Non-Goals

- new public chat endpoints
- channel moderation rules
- retroactive conversation membership rewriting

## Dependencies

- [112 Chat Resource Channel Model](112-chat-resource-channel-model.md)
- [04 Chat Offline milestone](../milestones/04-chat-offline.md)
- [Chat HTTP contract](../../../api/http/chat.md)

## Acceptance Criteria

- stale guild or party members cannot send into resource-bound channels
- stale guild or party members lose conversation visibility and replay access
- delivery planning only targets currently authorized members for resource-bound channels
- `go test ./...` passes

## Related Docs / ADRs

- [Current plan](../../current.md)
- [V1 freeze](../freeze.md)

## Completion Notes

- chat now uses lightweight guild and party membership readers for resource-bound permission checks
- public chat routes remain stable while runtime access is aligned to current guild and party state
