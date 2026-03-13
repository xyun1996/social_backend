# Task 023 - Document Persistence Boundaries

## Context

The repository now has runnable in-memory prototypes for all core services, plus explicit HTTP, proto, and TCP baselines. Local MySQL and Redis defaults are documented, but there is still no concrete repository-level design for how each service should move from in-memory state to durable or hot-state storage.

## Goal

Write the first persistence boundary baseline so future MySQL and Redis integration work follows explicit ownership and storage rules instead of ad hoc repo design.

## Scope

- Add a persistence design document under `docs/architecture/`
- Map each runnable service to its intended MySQL or Redis responsibility
- Update related architecture and plan docs so persistence guidance is discoverable

## Non-Goals

- Implementing actual database repositories
- Creating migrations or Redis key schemas in code
- Finalizing shard or multi-region topology

## Acceptance Criteria

- Storage ownership is explicit per service
- MySQL and Redis roles are described at the service boundary level
- The plan and architecture overview link to the new persistence baseline

## Status

`done`
