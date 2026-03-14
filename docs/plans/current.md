# Current Plan

- Version: `v1-freeze`
- Last updated: `2026-03-14`
- Source of truth level: highest

## Current Goal

Finish the smallest shippable `v1` of the Social Backend by freezing scope, closing only delivery-blocking gaps, and preparing a clean release handoff.

## Success Criteria

- Core services are locally runnable in durable mode.
- The main social flows are present and testable:
  - identity login/session
  - social friend/block
  - invite lifecycle
  - chat conversation/send/ack/replay/summary
  - guild create/invite/join/governance/growth baseline
  - party create/invite/ready/queue/assignment/resolution
  - ops player/guild/party/durable reads
- `go test ./...`, `make check-dev`, and `make test-local-durable` remain green.
- Scope boundaries between `v1` and `v2` are explicit.

## In Scope

- Final `v1` contract and implementation alignment
- Only the missing cross-service rules that block a credible `v1` handoff
- Local durable verification and release-oriented documentation
- Milestone and task status cleanup so project state matches code state

## Out of Scope

- Deeper feature expansion that is not required for `v1` acceptance
- Production deployment automation, CI/CD, and multi-region work
- Rich-media chat, advanced moderation, or heavy ops UI
- Advanced worker retry orchestration and full matchmaker lifecycle modeling

## Active Milestones

1. Freeze `v1` scope and handoff criteria
2. Close the last core runtime alignment gaps
3. Run final local durable regression
4. Publish `v1` release notes and known gaps

## Current Focus

- Keep implementation changes limited to `v1`-blocking gaps
- Treat feature-deepening work as `v2` unless it is required for acceptance
- Use [docs/plans/v1/freeze.md](v1/freeze.md) as the detailed finish line

## Current Risks

- The codebase is ahead of milestone bookkeeping in a few areas, so release-state drift is still possible without deliberate doc updates.
- Some services already support deeper prototypes than `v1` needs, which makes accidental scope creep likely.
- Local durable checks are strong, but release framing still needs a single, explicit finish line.

## Key Dependencies

- [docs/plans/v1/freeze.md](v1/freeze.md)
- [docs/plans/v1/roadmap.md](v1/roadmap.md)
- [docs/plans/backlog.md](backlog.md)
- [docs/releases/changelog.md](../releases/changelog.md)
- [docs/operations/local-durable-runflow.md](../operations/local-durable-runflow.md)
- [docs/operations/dev-checks.md](../operations/dev-checks.md)

## Active ADRs

- [ADR-001 Transport Strategy](../memory/adr/ADR-001-transport-strategy.md)
- [ADR-002 Session Granularity](../memory/adr/ADR-002-session-granularity.md)
- [ADR-003 Realm Isolation Default](../memory/adr/ADR-003-realm-isolation-default.md)
- [ADR-004 Message Delivery Semantics](../memory/adr/ADR-004-message-delivery-semantics.md)
- [ADR-005 Contract-First Boundaries](../memory/adr/ADR-005-contract-first-boundaries.md)
- [ADR-006 Presence Authority](../memory/adr/ADR-006-presence-authority.md)

## Update Rules

- Update this file first when the `v1` finish line changes.
- Move anything non-blocking for `v1` into [docs/plans/backlog.md](backlog.md) instead of expanding active scope.
- Reflect release-facing changes in freeze notes, milestones, tasks, and changelog together.
