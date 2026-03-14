# Current Plan

- Version: `v2.0`
- Last updated: `2026-03-14`
- Source of truth level: highest

## Current Goal

Ship the smallest complete `v2.0` slice by deepening `guild` into a real progression system and wiring `guild` governance and activity events into `chat`, `ops`, and `worker`.

## Success Criteria

- Guild progression, contribution, activity instance, and reward bookkeeping are durable and queryable.
- Guild activity submissions update guild xp and member contribution with period limits and idempotency.
- Guild activity and governance events are visible in the guild chat channel.
- Ops can read guild progression, contribution, activity instance, and reward state.
- `go test ./...`, `make check-dev`, and `make test-local-durable` stay green.

## In Scope

- `guild` progression and activity templates
- `guild` contribution leaderboard and reward record skeleton
- `guild` -> `chat` system event publishing
- Minimal `worker` support for guild activity period maintenance
- `ops` reads for guild progression surfaces
- `v2.0` plan and contract documentation

## Out of Scope

- Cross-guild competition or rankings
- Rich reward fulfillment or inventory settlement
- Full world/system/custom chat governance expansion
- Advanced worker scheduling DSL or retry orchestration
- Matchmaker lifecycle work

## Active Milestones

1. [V2.0-M1 Guild Progression](v2/v2.0/milestones/01-guild-progression.md)
2. [V2.0-M2 Guild Chat Integration](v2/v2.0/milestones/02-guild-chat-integration.md)

## Current Focus

- Use [docs/plans/v2/v2.0/roadmap.md](v2/v2.0/roadmap.md) as the active execution surface.
- Keep broader [docs/plans/v2/roadmap.md](v2/roadmap.md) as the higher-level post-v1 theme map.
- Route new implementation work through `v2.0` milestones first.

## Current Risks

- Proto contracts still lag the HTTP/runtime implementation and need a dedicated follow-up.
- Guild chat events currently use text-first system messages rather than rich cards.
- Worker support is intentionally minimal and should not be mistaken for a general scheduling platform.

## Key Dependencies

- [docs/plans/v2/v2.0/roadmap.md](v2/v2.0/roadmap.md)
- [docs/plans/v2/roadmap.md](v2/roadmap.md)
- [docs/plans/backlog.md](backlog.md)
- [docs/releases/project-archive-v1.md](../releases/project-archive-v1.md)

## Active ADRs

- [ADR-001 Transport Strategy](../memory/adr/ADR-001-transport-strategy.md)
- [ADR-002 Session Granularity](../memory/adr/ADR-002-session-granularity.md)
- [ADR-003 Realm Isolation Default](../memory/adr/ADR-003-realm-isolation-default.md)
- [ADR-004 Message Delivery Semantics](../memory/adr/ADR-004-message-delivery-semantics.md)
- [ADR-005 Contract-First Boundaries](../memory/adr/ADR-005-contract-first-boundaries.md)
- [ADR-006 Presence Authority](../memory/adr/ADR-006-presence-authority.md)

## Update Rules

- `v1` release docs remain historical facts.
- `v2.0` is the active implementation surface until its freeze line is declared.
- Wider `v2` roadmap themes should not absorb `v2.0` status details; keep them in the dedicated `v2.0` plan tree.
