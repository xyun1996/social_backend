# 008 Chat HTTP Prototype

- Title: Add an in-memory chat prototype for conversation, seq, ack, and replay flows
- Status: `done`
- Version: `v1`
- Milestone: `04 Chat Offline`

## Background

The roadmap and ADRs define chat delivery semantics, but the repository still lacks an executable prototype for conversation lifecycle, message sequencing, read cursors, and offline replay.

## Goal

Add a local in-memory `chat` service prototype that exercises conversation creation, built-in channel validation, message send sequencing, read acknowledgement, and replay by seq window.

## Scope

- Add chat domain models for conversation, message, and read cursor
- Add in-memory chat service logic for channel validation, send, ack, and replay
- Add HTTP endpoints and tests for conversation creation, message send, ack, and replay
- Add a runnable `chat` service entrypoint
- Align the active plan with the move from invite into chat prototyping

## Non-Goals

- Durable hot or cold message persistence
- Gateway push fanout or websocket delivery
- Attachment, moderation, or rich media support
- Resume tokens or multi-device dedupe protocols

## Dependencies

- [003 Identity HTTP Prototype](003-identity-http-prototype.md)
- [005 Social Graph HTTP Prototype](005-social-graph-http-prototype.md)
- [007 Invite HTTP Prototype](007-invite-http-prototype.md)
- [04 Chat Offline milestone](../milestones/04-chat-offline.md)
- [ADR-004 Message Delivery Semantics](../../../memory/adr/ADR-004-message-delivery-semantics.md)

## Acceptance Criteria

- Conversations can be created for supported built-in kinds
- Messages receive stable monotonic per-conversation seq values
- Read cursors can ack seq values without moving backward
- Replay returns messages with `seq > after_seq`
- `go test ./...` passes

## Related Docs / ADRs

- [Current plan](../../current.md)
- [Constraints](../../../memory/constraints.md)
- [Architecture overview](../../../architecture/overview.md)
