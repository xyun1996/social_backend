# Current Plan

- Version: `v2-planning`
- Last updated: `2026-03-14`
- Source of truth level: highest

## Current Goal

Turn the shipped `v1` release into a clear `v2` execution plan by selecting the highest-value deepening work, defining milestone boundaries, and keeping `v1` stable while new scope is introduced.

## Success Criteria

- `v1` remains green and releasable while planning moves forward.
- `v2` roadmap and milestone boundaries are explicit.
- Deferred items from `v1` are grouped into coherent, execution-ready themes.
- Future work can start without reopening `v1` scope accidentally.

## In Scope

- `v2` roadmap and milestone design
- Scope shaping for the highest-value post-`v1` capabilities
- Backlog consolidation and reprioritization
- Planning docs that separate `v2` from the shipped `v1`

## Out of Scope

- Reinterpreting or widening the released `v1` baseline
- Large implementation changes before `v2` milestone boundaries are clear
- Production infra automation and multi-region rollout planning beyond high-level placeholders

## Active Milestones

1. Social graph depth and richer relationship reads
2. Chat governance and advanced channel control
3. Guild progression depth and richer activity systems
4. Queue, worker, and runtime hardening
5. Operator tooling and release-readiness expansion

## Current Focus

- Use [docs/plans/v2/roadmap.md](v2/roadmap.md) as the active planning surface
- Keep `v1` release documents intact as the completed baseline
- Route new implementation work through `v2` milestones and tasks

## Current Risks

- It is easy to restart coding directly from backlog items without first tightening `v2` boundaries.
- Several deferred themes overlap across services, so poor grouping would create plan churn later.
- Without a clear `v2` shape, new work could accidentally erode the clean `v1` release line.

## Key Dependencies

- [docs/plans/v2/roadmap.md](v2/roadmap.md)
- [docs/plans/backlog.md](backlog.md)
- [docs/plans/v1/freeze.md](v1/freeze.md)
- [docs/releases/release-notes/v1.0.md](../releases/release-notes/v1.0.md)
- [docs/architecture/overview.md](../architecture/overview.md)

## Active ADRs

- [ADR-001 Transport Strategy](../memory/adr/ADR-001-transport-strategy.md)
- [ADR-002 Session Granularity](../memory/adr/ADR-002-session-granularity.md)
- [ADR-003 Realm Isolation Default](../memory/adr/ADR-003-realm-isolation-default.md)
- [ADR-004 Message Delivery Semantics](../memory/adr/ADR-004-message-delivery-semantics.md)
- [ADR-005 Contract-First Boundaries](../memory/adr/ADR-005-contract-first-boundaries.md)
- [ADR-006 Presence Authority](../memory/adr/ADR-006-presence-authority.md)

## Update Rules

- `v1` release docs are historical facts; do not edit them to absorb `v2` scope.
- New implementation work should be attached to a `v2` milestone first.
- Keep backlog, roadmap, and milestone docs aligned when priorities change.
