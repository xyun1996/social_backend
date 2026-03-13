# ADR-005 Contract-First Boundaries

- Status: `accepted`
- Date: `2026-03-13`

## Context

The repository now has runnable in-memory prototypes for `identity`, `social`, `invite`, `chat`, `party`, and `guild`, but the shared contract directories under `api/` are still empty placeholders.

If implementation continues without a contract baseline, the HTTP prototype shapes will drift from future gRPC and TCP integration, and downstream services will start depending on code-local request or response structs instead of explicit service contracts.

## Decision

Adopt a contract-first boundary workflow for the next implementation slice:

- `api/http/` becomes the source of truth for current control-plane endpoint contracts used by runnable prototypes.
- `api/proto/` is reserved for future service-to-service gRPC contracts, but protobuf generation is deferred until the HTTP contract baseline is written and the prototype shapes settle.
- Shared error codes and response envelope conventions must be documented under `api/errors/` before additional cross-service clients are added.
- New prototype endpoints should update the matching `api/http/` contract doc in the same change whenever wire-visible behavior changes.

## Alternatives Considered

- Continue prototype-first coding and defer all contract documentation until after service behavior stabilizes
- Skip HTTP contract documentation and jump directly to protobuf definitions

## Consequences

- The next iteration must spend time documenting wire-visible behavior before adding more services.
- HTTP remains the practical contract surface for current runnable prototypes.
- Future gRPC work should be shaped by stabilized domain contracts rather than by ad hoc handler structs.
- Review of API changes becomes easier because contract diffs will live under `api/`.
