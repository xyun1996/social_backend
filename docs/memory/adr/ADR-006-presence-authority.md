# ADR-006 Presence Authority

- Status: `accepted`
- Date: `2026-03-13`

## Context

`presence` is already a declared bounded context, and party, guild, and chat all list presence expectations as inputs or dependencies. However, the repository has no executable presence service and no documented authority split between `gateway` and `presence`.

Without an explicit decision, online state can end up split across gateway-local memory, chat session state, and domain-specific membership checks.

## Decision

Use `presence` as the single service authority for online state and short-lived player runtime context.

- `gateway` owns connection lifecycle and authentication, but reports connect, heartbeat, and disconnect events to `presence`.
- `presence` owns online or offline state, last-seen timestamps, and lightweight runtime metadata such as realm or location hints.
- Downstream services (`chat`, `party`, `guild`, `ops`) should read presence through an explicit service boundary rather than keep their own online-state source of truth.
- The first executable prototype may use an in-memory HTTP implementation, but its model should align with future Redis-backed short-lived state.

## Alternatives Considered

- Keep online state inside `gateway` and let downstream services call gateway directly
- Allow each domain service to cache or own its own player online state

## Consequences

- A dedicated presence prototype is now a near-term prerequisite for deeper chat, guild, and party work.
- Gateway and presence will need a documented contract for session state reporting.
- Future Redis adoption should replace the storage mechanism, not the ownership model.
