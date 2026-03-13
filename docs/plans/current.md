# Current Plan

- Version: `v1`
- Last updated: `2026-03-13`
- Source of truth level: highest

## Current Goal

Extend the prototype layer with runtime-aware cross-service behavior, prioritizing presence-backed chat and party flows after contract and authority boundaries were established.

## Success Criteria

- Existing prototype services remain runnable and documented.
- Shared boundary decisions are explicit for API contracts and presence ownership.
- `chat` and `party` now consume `presence` through explicit service boundaries.
- The next implementation slice can expand runtime-aware behavior without re-deciding gateway or presence ownership.

## In Scope

- Presence-backed runtime behavior for chat and party
- HTTP and error contract baseline documentation
- Documentation alignment between active plan, milestones, tasks, and ADRs
- Continued prototype hardening through explicit service clients

## Out of Scope

- Durable storage integrations
- Production-ready service decomposition
- CI/CD pipelines beyond placeholder entrypoints
- Full deployment manifests or runtime configs

## Active Milestones

1. Foundation scaffold
2. Identity and session prototype
3. Social graph prototype
4. Invite lifecycle prototype
5. Chat and offline messaging prototype
6. Guild system design
7. Party and queue design

## Current Risks

- HTTP contracts are still baseline-level, so deeper per-service specs and future proto contracts remain to be written.
- Guild and chat push behavior still need richer runtime rules beyond current prototype reads.
- Without disciplined updates, future plan drift could appear between `current`, milestones, tasks, and ADRs.

## Key Dependencies

- [docs/plans/v1/roadmap.md](v1/roadmap.md)
- [docs/plans/v1/tasks/007-invite-http-prototype.md](v1/tasks/007-invite-http-prototype.md)
- [docs/plans/v1/tasks/008-chat-http-prototype.md](v1/tasks/008-chat-http-prototype.md)
- [docs/plans/v1/tasks/009-party-http-prototype.md](v1/tasks/009-party-http-prototype.md)
- [docs/plans/v1/tasks/010-guild-http-prototype.md](v1/tasks/010-guild-http-prototype.md)
- [docs/plans/v1/tasks/011-document-http-contract-baseline.md](v1/tasks/011-document-http-contract-baseline.md)
- [docs/plans/v1/tasks/012-presence-http-prototype.md](v1/tasks/012-presence-http-prototype.md)
- [docs/plans/v1/tasks/013-presence-aware-chat-delivery.md](v1/tasks/013-presence-aware-chat-delivery.md)
- [docs/plans/v1/tasks/014-presence-aware-party-runtime.md](v1/tasks/014-presence-aware-party-runtime.md)
- [docs/plans/v1/tasks/015-presence-aware-guild-runtime.md](v1/tasks/015-presence-aware-guild-runtime.md)
- [docs/plans/v1/tasks/016-proto-contract-baseline.md](v1/tasks/016-proto-contract-baseline.md)
- [docs/plans/v1/tasks/017-ops-http-read-prototype.md](v1/tasks/017-ops-http-read-prototype.md)
- [docs/plans/v1/tasks/018-worker-http-prototype.md](v1/tasks/018-worker-http-prototype.md)
- [docs/plans/v1/tasks/019-expand-proto-contracts.md](v1/tasks/019-expand-proto-contracts.md)
- [docs/plans/v1/tasks/020-ops-worker-queue-visibility.md](v1/tasks/020-ops-worker-queue-visibility.md)
- [api/http/README.md](../api/http/README.md)
- [api/errors/README.md](../api/errors/README.md)
- [docs/memory/constraints.md](../memory/constraints.md)
- [docs/memory/glossary.md](../memory/glossary.md)
- [docs/architecture/overview.md](../architecture/overview.md)
- [docs/operations/environments.md](../operations/environments.md)

## Active ADRs

- [ADR-001 Transport Strategy](../memory/adr/ADR-001-transport-strategy.md)
- [ADR-002 Session Granularity](../memory/adr/ADR-002-session-granularity.md)
- [ADR-003 Realm Isolation Default](../memory/adr/ADR-003-realm-isolation-default.md)
- [ADR-004 Message Delivery Semantics](../memory/adr/ADR-004-message-delivery-semantics.md)
- [ADR-005 Contract-First Boundaries](../memory/adr/ADR-005-contract-first-boundaries.md)
- [ADR-006 Presence Authority](../memory/adr/ADR-006-presence-authority.md)

## Update Rules

- Update this file first when scope, version direction, or milestone priority changes.
- Reflect downstream changes in roadmap, tasks, ADRs, and architecture docs after this file is updated.
