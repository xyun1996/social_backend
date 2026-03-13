# 046 Integration Gateway Ack Compaction Flow

## Goal

Cover the new gateway ack compaction behavior with a real cross-service local integration flow.

## Scope

- drive `identity`, `presence`, `chat`, and `gateway` through local HTTP test servers
- verify delivery creates a buffered session event
- verify a session-scoped chat ack compacts that buffered event

## Acceptance

- integration coverage proves `gateway -> chat ack -> gateway inbox compaction`
- `go test ./services/integration/...` passes
