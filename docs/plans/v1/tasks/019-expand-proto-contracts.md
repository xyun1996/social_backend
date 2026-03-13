# 019 Expand Proto Contracts

- Title: Expand the proto baseline to runtime-facing chat, party, guild, ops, and worker services
- Status: `done`
- Version: `v1`
- Milestone: `01 Foundation`

## Background

The initial proto baseline now covers `identity`, `presence`, and `invite`, but the remaining runtime-facing services still only have HTTP contract documentation.

## Goal

Extend `api/proto/` so the main runtime service boundaries have a first explicit internal contract baseline.

## Scope

- Add proto contracts for `chat`
- Add proto contracts for `party`
- Add proto contracts for `guild`
- Add proto contracts for `ops`
- Add proto contracts for `worker`

## Non-Goals

- gRPC server generation
- Transport migrations
- One-to-one parity for every HTTP envelope field

## Dependencies

- [016 Proto Contract Baseline](016-proto-contract-baseline.md)
- [013 Presence-Aware Chat Delivery](013-presence-aware-chat-delivery.md)
- [014 Presence-Aware Party Runtime](014-presence-aware-party-runtime.md)
- [015 Presence-Aware Guild Runtime](015-presence-aware-guild-runtime.md)
- [017 Ops HTTP Read Prototype](017-ops-http-read-prototype.md)
- [018 Worker HTTP Prototype](018-worker-http-prototype.md)

## Acceptance Criteria

- `api/proto/` covers the remaining runtime-facing services
- Proto baselines reflect the currently documented HTTP boundary semantics
- `go test ./...` passes

## Related Docs / ADRs

- [Current plan](../../current.md)
- [Architecture protocols](../../../architecture/protocols.md)
