# 046 Integration Gateway Ack Compaction Flow

## Status

`done`

## Goal

Cover the new gateway ack compaction behavior with a real cross-service local integration flow.

## Scope

- drive `identity`, `presence`, `chat`, and `gateway` through local HTTP test servers
- verify delivery creates a buffered session event
- verify a session-scoped chat ack compacts that buffered event

## Acceptance

- integration coverage proves `gateway -> chat ack -> gateway inbox compaction`
- `go test ./services/integration/...` passes

## Completion Notes

- local integration coverage now drives identity, presence, chat, and gateway through a real ack compaction flow
- the integration test verifies that a delivered chat event is buffered and then removed after the session-scoped ack succeeds
