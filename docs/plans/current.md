# Current Plan

- Version: `v1`
- Last updated: `2026-03-12`
- Source of truth level: highest

## Current Goal

Stand up the repository scaffold and governance model for the Social Backend project so implementation can proceed without structural ambiguity.

## Success Criteria

- The repository has stable top-level directories for code, protocols, tooling, tests, and docs.
- `docs/plans/current.md` clearly points to the active version roadmap and milestones.
- Key constraints, glossary terms, ADRs, and architecture entrypoints exist.
- New contributors can find current scope and next steps within 10 minutes.

## In Scope

- Repository layout and governance documents
- `v1` roadmap and milestone placeholders
- Initial ADR set for foundational decisions
- Templates for plans, tasks, ADRs, runbooks, and release notes

## Out of Scope

- Production-ready Go modules
- Service implementation code
- CI/CD pipelines beyond placeholder entrypoints
- Full deployment manifests or runtime configs

## Active Milestones

1. Foundation scaffold
2. Identity and session design
3. Social graph design
4. Chat and offline messaging design
5. Guild system design
6. Party and queue design

## Current Risks

- Go toolchain is not installed or not available in the current environment.
- Service contracts are not authored yet, so protocol compatibility rules are still documentation-only.
- Without disciplined updates, future plan drift could appear between `current`, milestones, tasks, and ADRs.

## Key Dependencies

- [docs/plans/v1/roadmap.md](v1/roadmap.md)
- [docs/memory/constraints.md](../memory/constraints.md)
- [docs/memory/glossary.md](../memory/glossary.md)
- [docs/architecture/overview.md](../architecture/overview.md)
- [docs/operations/environments.md](../operations/environments.md)

## Active ADRs

- [ADR-001 Transport Strategy](../memory/adr/ADR-001-transport-strategy.md)
- [ADR-002 Session Granularity](../memory/adr/ADR-002-session-granularity.md)
- [ADR-003 Realm Isolation Default](../memory/adr/ADR-003-realm-isolation-default.md)
- [ADR-004 Message Delivery Semantics](../memory/adr/ADR-004-message-delivery-semantics.md)

## Update Rules

- Update this file first when scope, version direction, or milestone priority changes.
- Reflect downstream changes in roadmap, tasks, ADRs, and architecture docs after this file is updated.
