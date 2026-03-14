# Current Plan

- Version: `v1`
- Last updated: `2026-03-14`
- Source of truth level: highest

## Current Goal

`v1` is complete. Preserve the release line, keep the repository healthy, and route any feature-deepening work into backlog and future versions.

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

- Preserve the shipped `v1` release line
- Use [docs/plans/v1/freeze.md](v1/freeze.md) as the historical acceptance line
- Route new expansion work into backlog and `v2`

## Current Risks

- Future work could accidentally reopen `v1` scope instead of being scheduled as `v2`.
- Durable local assumptions are strong, but production deployment and advanced operations remain intentionally out of scope.

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

- Do not expand `v1` scope from this file; new feature work should go through backlog and future version plans.
- Keep release-facing documents aligned when the shipped baseline changes.
