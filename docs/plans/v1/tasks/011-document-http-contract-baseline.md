# 011 Document HTTP Contract Baseline

- Title: Document the current HTTP contract baseline under `api/http`
- Status: `done`
- Version: `v1`
- Milestone: `01 Foundation`

## Background

Runnable prototypes now exist for identity, social, invite, chat, party, and guild, but the shared `api/http/` and `api/errors/` directories are still effectively empty.

## Goal

Create a written HTTP contract baseline for the current control-plane services so future handler changes are anchored to explicit wire contracts.

## Scope

- Add shared HTTP contract documentation under `api/http/`
- Add shared error contract documentation under `api/errors/`
- Document endpoint shapes for the currently runnable prototype services
- Define update rules for future wire-visible changes

## Non-Goals

- Protobuf generation
- TCP frame definitions
- Full OpenAPI generation pipeline

## Dependencies

- [ADR-005 Contract-First Boundaries](../../../memory/adr/ADR-005-contract-first-boundaries.md)
- [002 Standardize HTTP Foundation](002-standardize-http-foundation.md)

## Acceptance Criteria

- `api/http/` documents the current runnable HTTP surfaces
- `api/errors/` documents the shared application error model
- Future prototype changes have an obvious contract update home

## Related Docs / ADRs

- [Current plan](../../current.md)
- [Architecture protocols](../../../architecture/protocols.md)
