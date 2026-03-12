# 004 Wire Gateway Session Introspection

- Title: Wire gateway session lookup through identity token introspection
- Status: `done`
- Version: `v1`
- Milestone: `02 Identity Session`

## Background

The identity prototype can issue player-scoped tokens, but the gateway still lacks a way to resolve those tokens into authenticated session context.

## Goal

Add token introspection to the identity prototype and use it from the gateway to expose a first authenticated session endpoint.

## Scope

- Add access-token introspection to the in-memory identity service
- Add an identity introspection HTTP endpoint
- Add a gateway-side identity client and authenticated `/v1/session/me` endpoint
- Add tests for introspection and gateway bearer-token behavior

## Non-Goals

- Real inter-service authentication hardening
- TCP handshake enforcement
- Persistent token storage
- Gateway request proxying

## Dependencies

- [003 Identity HTTP Prototype](003-identity-http-prototype.md)
- [02 Identity Session milestone](../milestones/02-identity-session.md)
- [ADR-001 Transport Strategy](../../../memory/adr/ADR-001-transport-strategy.md)
- [ADR-002 Session Granularity](../../../memory/adr/ADR-002-session-granularity.md)

## Acceptance Criteria

- Identity exposes a token introspection endpoint
- Gateway can resolve a bearer token into account and player context
- `go test ./...` passes

## Related Docs / ADRs

- [Current plan](../../current.md)
- [Architecture protocols](../../../architecture/protocols.md)
