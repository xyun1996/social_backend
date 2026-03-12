# Protocols

## Surfaces

- HTTP is the control surface for login, refresh, administrative actions, and non-real-time queries.
- TCP with Protobuf is the primary real-time transport.
- WebSocket is the compatibility real-time transport for tooling, cross-platform access, and debugging.
- gRPC is the primary service-to-service interface.

## Compatibility Rules

- New fields must be backward-compatible by default.
- Deprecated fields should remain readable for at least one active version cycle.
- Protocol changes that alter wire behavior must update this doc and, if durable, the relevant ADR.

## TCP Notes

- Frame format, handshake, heartbeat, ack, and resume rules belong under `api/tcp/`.
- Real-time envelope semantics should be shared across TCP and WebSocket where possible.

## HTTP Notes

- Endpoint contracts belong under `api/http/`.
- Authentication tokens are acquired and refreshed through the control surface.

## gRPC Notes

- Shared service contracts and envelope messages belong under `api/proto/`.
- Internal APIs should avoid leaking storage-specific details.

## Error Model

- Shared error code definitions belong under `api/errors/`.
- User-facing errors and internal errors should map through a consistent contract.
