# Current Plan

- Version: `product-rebuild`
- Last updated: `2026-03-14`
- Source of truth level: highest

## Current Goal

Reposition the repository from “feature-complete prototype” to “product rebuild in progress”, using the current codebase as a reference asset set rather than the final runtime target.

## Success Criteria

- The new target runtime is explicitly defined as `api-gateway + social-core + ops-worker`.
- Existing prototype services are treated as frozen reference implementations.
- Product rebuild documentation, milestones, and runtime entrypoints exist.
- New implementation work starts on the rebuilt runtime line rather than widening the prototype surface.

## Status

`product-rebuild` is active.

## Active Milestones

1. [01 Runtime Consolidation](product/milestones/01-runtime-consolidation.md)
2. [02 Foundation Rebuild](product/milestones/02-foundation-rebuild.md)
3. [03 Phase A Core Social Package](product/milestones/03-phase-a-core-social.md)
4. [04 Staging Release Readiness](product/milestones/04-staging-release-readiness.md)

## Remaining Follow-ups

- Production hardening work remains useful, but now serves the rebuild instead of the frozen prototype line.
- The old per-service runtime remains available for regression comparison and domain reference.
- Product-grade implementation depth still needs to be rebuilt on the new runtime.

## Key Dependencies

- [docs/plans/product-rebuild.md](product-rebuild.md)
- [docs/plans/product/roadmap.md](product/roadmap.md)
- [docs/plans/v2/roadmap.md](v2/roadmap.md)
- [docs/plans/production/roadmap.md](production/roadmap.md)
- [docs/plans/backlog.md](backlog.md)
- [docs/releases/project-archive-v1.md](../releases/project-archive-v1.md)

## Update Rules

- `v1` and `v2` release docs remain historical facts.
- `production` hardening remains a valid reference line but is no longer the active implementation target.
- New runtime and product work should land on the rebuild path first.
