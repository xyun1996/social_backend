# Architecture Overview

## Purpose

This project is a reusable social backend for medium-light multiplayer games. The architecture is split into:

- client and operator entry surfaces
- domain services
- shared infrastructure packages
- governance documents that preserve intent and decisions over time

## Planned Runtime Topology

### Frozen Prototype Topology

- `gateway` handles TCP/WebSocket connections, authentication checks, session ownership, and push delivery.
- `identity` handles login, refresh, account binding, and player selection.
- `presence` tracks online state and short-lived player context.
- `social` owns friend and block relationships.
- `guild` owns guild organization, roles, progression, and activity state.
- `invite` owns invitation lifecycle across domains.
- `chat` owns conversations, channels, message sequencing, and offline replay.
- `party` owns party management and social queue orchestration.
- `ops` provides operator-facing APIs.
- `worker` executes asynchronous jobs and compensations.

### Active Product Rebuild Topology

- `api-gateway` is the only client-facing ingress target for the rebuilt runtime.
- `social-core` holds the first product-grade implementation of the core social package.
- `ops-worker` combines support reads, repair flows, and async execution for the rebuilt runtime.

## Repository View

- `api/` for shared contracts
- `api/http`, `api/proto`, and `api/tcp` hold the current wire-level baselines
- `services/` for domain implementations
- `pkg/` for shared infrastructure only
- `docs/` for plans, architecture, operations, releases, and memory

## Design Entry Points

- [Module Design](module-design.md)
- [Product Runtime Topology](product-runtime-topology.md)
- [Technical Principles](technical-principles.md)
- [Runtime Flows](runtime-flows.md)
- [Persistence Boundaries](persistence.md)
- [Bounded Contexts](bounded-contexts.md)
- [Dependencies](dependencies.md)
- [Protocols](protocols.md)
- [Proto generation workflow](../operations/proto-generation.md)

## Governance View

- `docs/plans/current.md` is the active source of truth for current version scope.
- ADRs preserve durable decisions.
- architecture docs preserve structural facts.
- session notes preserve discussion history.
