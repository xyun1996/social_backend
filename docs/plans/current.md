# Current Plan

- Version: `v1`
- Last updated: `2026-03-13`
- Source of truth level: highest

## Current Goal

Move from executable in-memory prototypes into explicit boundary decisions, prioritizing contract documentation and presence authority before deeper realtime integration.

## Success Criteria

- Existing prototype services remain runnable and documented.
- Shared boundary decisions are explicit for API contracts and presence ownership.
- `api/http/` and `api/errors/` become active homes for wire-contract documentation.
- The next implementation slice can add presence without re-deciding gateway ownership or contract direction.

## In Scope

- Boundary decisions for contract ownership and presence authority
- HTTP and error contract baseline documentation
- Presence prototype planning aligned with gateway and downstream services
- Documentation alignment between active plan, milestones, tasks, and ADRs

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

- HTTP contracts are only baseline-level, so detailed per-service specs still need to be written.
- Presence has no executable prototype yet, so online-state consumers still rely on future work.
- Without disciplined updates, future plan drift could appear between `current`, milestones, tasks, and ADRs.

## Key Dependencies

- [docs/plans/v1/roadmap.md](v1/roadmap.md)
- [docs/plans/v1/tasks/007-invite-http-prototype.md](v1/tasks/007-invite-http-prototype.md)
- [docs/plans/v1/tasks/008-chat-http-prototype.md](v1/tasks/008-chat-http-prototype.md)
- [docs/plans/v1/tasks/009-party-http-prototype.md](v1/tasks/009-party-http-prototype.md)
- [docs/plans/v1/tasks/010-guild-http-prototype.md](v1/tasks/010-guild-http-prototype.md)
- [docs/plans/v1/tasks/011-document-http-contract-baseline.md](v1/tasks/011-document-http-contract-baseline.md)
- [docs/plans/v1/tasks/012-presence-http-prototype.md](v1/tasks/012-presence-http-prototype.md)
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
