# ADR-001 Transport Strategy

- Status: `accepted`
- Date: `2026-03-12`

## Context

The project needs a real-time transport suitable for game clients while still supporting tooling, debugging, and broader compatibility.

## Decision

Use TCP with Protobuf as the primary real-time transport. Support WebSocket as a compatibility transport. Use HTTP for login and other control-surface operations. Use gRPC for service-to-service communication.

## Alternatives Considered

- WebSocket as the only real-time transport
- Fully custom binary protocol without shared Protobuf model
- HTTP polling for all client communication

## Consequences

- Gateway must own transport adaptation cleanly.
- Protocol compatibility rules must be documented and tested.
- Tooling should support both TCP and WebSocket flows.
