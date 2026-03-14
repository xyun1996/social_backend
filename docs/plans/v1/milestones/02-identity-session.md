# 02 Identity Session

- Status: `done`
- Version: `v1`

## Goal

Define and implement the account, token, player-selection, and real-time session model that all other domains rely on.

## Inputs

- Transport strategy ADR
- Session granularity ADR
- Environment and config conventions

## Outputs

- Identity service module
- Gateway handshake contracts
- Token validation flow
- Player-scoped session lifecycle rules
- Local in-memory login and refresh prototype

## Acceptance Criteria

- HTTP login and token refresh flow are specified.
- TCP/WebSocket connection establishment is documented and testable.
- Session ownership and reconnect behavior are explicit.
- Early identity endpoints are runnable without external dependencies.

## Risks

- Identity shortcuts can leak ambiguity into every downstream service.

## Completion Notes

- `identity`, `gateway`, and `presence` now provide the `v1` login, refresh, introspection, and player-scoped session baseline.
