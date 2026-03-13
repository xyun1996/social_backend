# Persistence Boundaries

## Purpose

This document defines the first storage ownership baseline for moving the current in-memory prototypes toward MySQL and Redis without collapsing service boundaries.

## Storage Roles

### MySQL

Use MySQL for durable relational state:

- accounts and issued identity artifacts that must survive process restarts
- friend graph and block graph
- invite records and long-lived status transitions
- chat conversations, message metadata, and read cursors
- party and guild membership records
- operator-visible audit or recovery state where durability matters

### Redis

Use Redis for hot or short-lived state:

- presence snapshots and heartbeat timestamps
- session-scoped routing hints for gateway delivery
- worker queue claims, retry timing, and hot backlog inspection
- replay acceleration caches and ephemeral fanout helpers

## Service Ownership Map

### `identity`

- MySQL-owned: account, player, refresh-token lineage if refresh durability is required
- Redis-owned: optional short-lived token introspection cache
- First code foundation lives under `services/identity/internal/repo/mysql`

### `gateway`

- Redis-owned: active session routing hints only
- Rule: gateway does not become the source of truth for social, chat, party, or guild state

### `presence`

- Redis-owned: player online state, heartbeat, realm/location/session snapshot
- Rule: presence stays the authority for online state even if gateway keeps transport-local connection objects
- First code foundation lives under `services/presence/internal/repo/redis`

### `social`

- MySQL-owned: friend requests, accepted friendships, block relationships
- Redis-owned: optional relationship read-through cache

### `invite`

- MySQL-owned: invite lifecycle records, TTL deadline metadata, response timestamps
- Redis-owned: optional expiry scheduling hints if worker polling is introduced

### `chat`

- MySQL-owned: conversations, membership, messages, read cursors
- Redis-owned: optional recent-message cache, delivery planning cache, replay acceleration windows
- Rule: durable message order remains chat-owned and must not depend on gateway buffering

### `party`

- MySQL-owned: party membership and durable party state if parties must survive restarts
- Redis-owned: hot runtime state such as ready toggles, queue intent, or transient orchestration metadata

### `guild`

- MySQL-owned: guild roster, role assignments, durable guild metadata
- Redis-owned: hot member activity snapshots when needed for runtime reads

### `worker`

- MySQL-owned: durable job intent when jobs must survive restarts
- Redis-owned: claim state, retry scheduling hints, queue visibility, and hot dispatch bookkeeping

### `ops`

- Rule: ops should aggregate from service-owned stores or service APIs, not create an independent source of truth

## Repository Shape

When persistence is introduced, service-owned repositories should remain local to each service:

```text
services/<service>/internal/repo/mysql/
services/<service>/internal/repo/redis/
services/<service>/internal/repo/memory/
```

Rules:

- `repo/memory` remains useful for tests and local prototype parity
- Cross-service packages under `pkg/` must not own domain-specific queries
- Shared database helpers may live in `pkg/`, but schema ownership stays inside the service

## Local Development Baseline

Current disposable local defaults:

- MySQL: `localhost:3306`, `root / 1234`, database `social_backend`
- Redis: `localhost:6379`, no username, no password, database `0`

These defaults are for local development only and should map cleanly onto environment-specific overrides later.
