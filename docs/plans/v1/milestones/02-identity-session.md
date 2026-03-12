# 02 Identity Session

- Status: `planned`
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

## Acceptance Criteria

- HTTP login and token refresh flow are specified.
- TCP/WebSocket connection establishment is documented and testable.
- Session ownership and reconnect behavior are explicit.

## Risks

- Identity shortcuts can leak ambiguity into every downstream service.
