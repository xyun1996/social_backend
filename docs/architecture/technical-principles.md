# Technical Principles

## Purpose

This document explains the core technical principles behind the current design, including why service boundaries exist, how state is expected to flow, and what implementation constraints shape the code.

## Principle 1: Player Context Is The Runtime Unit

The active runtime identity is `player_id`, not `account_id`.

Why:

- social state is player-facing
- presence is player-facing
- chat unread and replay are player-facing
- guild and party membership are player-facing

Implementation consequence:

- identity must resolve account login into player-scoped subjects
- gateway forwards player-scoped context downstream

## Principle 2: Boundary Ownership Beats Convenience

Each service owns exactly one kind of truth.

- `identity` owns auth subject truth
- `presence` owns online truth
- `invite` owns invite status truth
- `chat` owns message ordering truth
- `party` owns party membership truth
- `guild` owns guild membership truth

Why:

- prevents silent duplication
- makes future Redis/MySQL moves local to one service
- gives API and protobuf contracts a stable center

## Principle 3: Contract First For Wire Behavior

The repository now treats `api/http/` as the source of truth for current wire-visible HTTP behavior.

Why:

- prototypes were multiplying faster than explicit contracts
- future gRPC and TCP design should derive from stable boundary semantics, not handler-local structs

Implementation consequence:

- any endpoint shape change should update `api/http/`
- shared error semantics live in `api/errors/`

## Principle 4: Realtime Delivery Is At-Least-Once

Chat delivery is intentionally designed around:

- stable conversation ids
- monotonic per-conversation seq
- monotonic ack cursor
- replay with `seq > after_seq`

Why:

- exactly-once semantics would force storage and delivery complexity too early
- game social systems usually tolerate dedupe better than dropped state

Implementation consequence:

- gateway and clients must dedupe by stable identifiers
- chat owns ordering and replay, not gateway

## Principle 5: Gateway Is A Coordinator, Not A State Store

Gateway should authenticate, attribute, and forward. It should not become the long-term owner of social state.

Why:

- otherwise gateway turns into a monolith that owns auth, presence, chat, and social state by accident
- service boundaries become impossible to preserve once realtime logic grows

Implementation consequence:

- gateway calls identity for introspection
- gateway reports presence transitions to presence
- gateway should later push chat and domain events without owning their state

## Principle 6: Hot State And Durable State Must Split

The intended storage model is:

- MySQL for durable relational state
- Redis for short-lived hot state

Why:

- presence, queue state, and chat hot buffers have different read/write patterns from durable domain data

Implementation consequence:

- current in-memory maps are placeholders for the future storage role, not for future co-location

## Principle 7: Prototypes Should Mirror Future Shapes

Even though current services are in-memory, their code layout should already mirror the durable architecture.

Why:

- moving from prototype to production is easier when domain, handler, client, and repo seams already exist

Implementation consequence:

- transport-specific code stays in `handler`
- outbound service calls stay in `client`
- future persistence belongs in `repo`

## Principle 8: Security And Observability Are Part Of Module Design

Security and observability are not afterthoughts.

- authenticated player identity should be derived at the gateway edge
- audit-sensitive state transitions must remain attributable
- request and player correlation should be loggable
- cross-service flows should become traceable

Why:

- social systems accumulate hard-to-debug state bugs quickly
- multiplayer support work depends on being able to explain why a player ended up in a given state

## Practical Design Heuristics

- prefer explicit service clients over direct package coupling
- prefer stable ids and monotonic counters over inferred ordering
- prefer shared invite semantics over per-domain reinvention
- prefer one source of truth with many readers over many partially-correct caches
- prefer adding a contract doc before adding a second consumer
