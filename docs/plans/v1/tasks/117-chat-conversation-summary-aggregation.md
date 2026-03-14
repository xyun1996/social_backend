# 117 Chat Conversation Summary Aggregation

- Title: Add unread-aware conversation summaries for chat reads
- Status: `done`
- Version: `v1`
- Milestone: `04 Chat Offline`

## Background

Chat already supported replay and ack cursors, but clients and operator reads still lacked a compact summary view that answered “what is unread and what was the last message” without replaying whole conversations.

## Goal

Add conversation summary reads that derive unread counts, ack cursors, and last-message snapshots from the existing chat state.

## Scope

- add a conversation summary domain model
- expose list and single-summary HTTP reads
- include `ack_seq`, `last_seq`, `unread_count`, and the last message snapshot
- align chat proto contracts with the new read surfaces
- add service and handler tests

## Non-Goals

- push-side unread badge fanout
- per-device read cursors
- summary pagination optimizations

## Dependencies

- [112 Chat Resource Channel Model](112-chat-resource-channel-model.md)
- [04 Chat Offline milestone](../milestones/04-chat-offline.md)
- [Chat HTTP contract](../../../api/http/chat.md)

## Acceptance Criteria

- `GET /v1/conversation-summaries?player_id=...` returns unread-aware summaries
- `GET /v1/conversations/{conversationID}/summary?player_id=...` returns the same derived state for a single conversation
- chat proto contracts cover the new summary reads
- `go test ./...` passes

## Related Docs / ADRs

- [Current plan](../../current.md)
- [Message delivery ADR](../../../memory/adr/ADR-004-message-delivery-semantics.md)

## Completion Notes

- chat summaries now expose unread counts directly from `last_seq` and `ack_seq`
- each summary includes the last delivered message snapshot and updated timestamp
- summary reads are available both as per-conversation and player-wide list endpoints
