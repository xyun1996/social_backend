# 044 Gateway Ack Inbox Compaction

## Goal

Make chat acknowledgements affect gateway runtime state, not only chat cursors, by compacting already acknowledged buffered chat events from a session inbox.

## Scope

- add gateway-local inbox compaction on session-scoped chat ack
- surface compaction results from the ack endpoint
- add handler and service tests
- align HTTP, TCP, and proto contract notes

## Non-Goals

- transport-packet acknowledgement
- durable event storage in gateway
- non-chat stream compaction

## Acceptance

- session chat ack prunes buffered events for that conversation through `ack_seq`
- ack response includes compaction summary
- `go test ./services/gateway/...` passes
