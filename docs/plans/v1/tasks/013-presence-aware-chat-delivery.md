# 013 Presence-Aware Chat Delivery

- Title: Add presence-aware delivery planning to the chat prototype
- Status: `done`
- Version: `v1`
- Milestone: `04 Chat Offline`

## Background

Chat now has sequencing, ack, and replay primitives, and presence is now the authority for online state, but chat still does not consume presence when deciding whether a recipient should be treated as online-push or offline-replay.

## Goal

Add a presence-aware delivery planning path to the in-memory `chat` prototype so message fanout can distinguish online recipients from offline replay recipients.

## Scope

- Add a presence client boundary for chat
- Add delivery planning logic based on conversation members and current presence state
- Add an HTTP endpoint to inspect delivery planning output
- Update local run configuration and HTTP contract docs

## Non-Goals

- WebSocket or TCP push implementation
- Durable offline message storage
- Presence subscriptions or broadcast fanout

## Dependencies

- [012 Presence HTTP Prototype](012-presence-http-prototype.md)
- [ADR-006 Presence Authority](../../../memory/adr/ADR-006-presence-authority.md)

## Acceptance Criteria

- Chat can resolve online recipients through the presence boundary
- Offline recipients fall back to replay-oriented delivery planning
- `go test ./...` passes

## Related Docs / ADRs

- [Current plan](../../current.md)
- [Chat HTTP contract](../../../../api/http/chat.md)
