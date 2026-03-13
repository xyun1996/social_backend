# Architecture Overview

## Purpose

This project is a reusable social backend for medium-light multiplayer games. The architecture is split into:

- client and operator entry surfaces
- domain services
- shared infrastructure packages
- governance documents that preserve intent and decisions over time

## Planned Runtime Topology

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

## Repository View

- `api/` for shared contracts
- `services/` for domain implementations
- `pkg/` for shared infrastructure only
- `docs/` for plans, architecture, operations, releases, and memory

## Design Entry Points

- [Module Design](module-design.md)
- [Technical Principles](technical-principles.md)
- [Runtime Flows](runtime-flows.md)
- [Bounded Contexts](bounded-contexts.md)
- [Dependencies](dependencies.md)
- [Protocols](protocols.md)

## Governance View

- `docs/plans/current.md` is the active source of truth for current version scope.
- ADRs preserve durable decisions.
- architecture docs preserve structural facts.
- session notes preserve discussion history.
