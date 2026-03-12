# ADR-003 Realm Isolation Default

- Status: `accepted`
- Date: `2026-03-12`

## Context

The project aims to support both isolated realm deployments and future global social graph expansion.

## Decision

Treat realm isolation as the default operating model in v1 while preserving tenant/region/realm-aware identifiers and routing so future global-graph evolution remains possible.

## Alternatives Considered

- Global graph by default
- Single-server-only assumptions

## Consequences

- Data and protocol context must preserve routing identifiers early.
- Scope stays manageable for v1 without closing off future federation or global expansion.
