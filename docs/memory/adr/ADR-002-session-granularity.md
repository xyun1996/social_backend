# ADR-002 Session Granularity

- Status: `accepted`
- Date: `2026-03-12`

## Context

The system models both platform accounts and game-facing player identities. Real-time features need a clear ownership boundary.

## Decision

Real-time sessions are scoped to `player_id`. Accounts authenticate and choose a player; downstream social, guild, chat, and party operations execute in player context.

## Alternatives Considered

- Account-scoped sessions
- Dual account-and-player session model in v1

## Consequences

- Session, presence, unread state, guild membership, and party state remain aligned to the game-facing identity.
- Identity service must make player selection explicit before real-time connection establishment.
