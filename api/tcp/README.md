# TCP Realtime Contracts

This directory is the source of truth for the current realtime transport baseline used by the future `gateway` transport layer.

## Scope

- connection handshake
- session resume
- heartbeat and idle timeout
- event envelope and acknowledgements
- chat-oriented replay handoff

## Rules

- Transport contracts describe wire semantics, not internal storage layout.
- TCP and WebSocket compatibility mode should share envelope semantics whenever practical.
- Sequence ownership stays with the domain service that owns ordering, not the gateway.
- Wire-visible changes here should update `docs/architecture/protocols.md` and any affected ADR.

## Current Surfaces

- [gateway](gateway.md): realtime connection lifecycle, heartbeat, and resume
- [chat](chat.md): event envelope, delivery, ack, and replay handoff
