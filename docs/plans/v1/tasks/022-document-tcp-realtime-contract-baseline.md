# Task 022 - Document TCP Realtime Contract Baseline

## Context

ADR-001 already fixed the transport strategy: TCP with Protobuf is the primary realtime transport, with WebSocket as a compatibility path. The architecture docs also point to `api/tcp/` as the home for frame format, handshake, heartbeat, ack, and resume rules. But the directory is still just a placeholder, so the realtime surface has no written contract baseline yet.

## Goal

Write the first realtime contract baseline under `api/tcp/` so gateway and chat evolution can share an explicit transport model before implementation starts.

## Scope

- Add `api/tcp/README.md`
- Add `api/tcp/gateway.md`
- Add `api/tcp/chat.md`
- Update architecture docs and current plan to reference the new baseline

## Non-Goals

- Implementing a TCP server
- Defining a final binary framing format
- Wiring websocket compatibility handlers

## Acceptance Criteria

- `api/tcp/` is no longer empty
- Handshake, heartbeat, ack, and resume rules are documented
- Chat realtime delivery semantics are tied back to the transport contract

## Status

`done`
