# Task 028 - Gateway Realtime Session Prototype

## Context

The repository already documents TCP handshake, heartbeat, resume, and disconnect semantics under `api/tcp/`, but `gateway` still only exposes control-plane HTTP endpoints for `session/me` and presence reporting. There is no executable prototype for the realtime session state machine itself.

## Goal

Add an in-memory realtime session prototype to `gateway` so handshake, heartbeat, resume, and close behavior can be exercised before a TCP server exists.

## Scope

- Add a gateway-owned realtime session manager
- Expose HTTP prototype endpoints for handshake, heartbeat, resume, close, and session lookup
- Forward connect, heartbeat, and disconnect transitions through the existing presence boundary
- Update docs, config notes, and tests

## Non-Goals

- Implementing the final TCP server
- WebSocket compatibility transport
- Gateway-driven chat push delivery

## Acceptance Criteria

- Gateway can create, resume, heartbeat, inspect, and close realtime sessions
- Resume validates authenticated ownership of the session
- Presence reporting stays derived from gateway-owned session lifecycle

## Status

`done`
