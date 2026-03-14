# V2 Roadmap

- Version: `v2`
- Status: `planned`
- Last updated: `2026-03-14`

## Goal

Deepen the shipped social backend from a strong local-durable `v1` baseline into a more operationally complete platform, focusing on richer relationship reads, stronger chat governance, deeper guild systems, and more resilient queue/runtime behavior.

## Milestone Order

1. [01 Social Depth](milestones/01-social-depth.md)
2. [02 Chat Governance](milestones/02-chat-governance.md)
3. [03 Guild Progression](milestones/03-guild-progression.md)
4. [04 Runtime Hardening](milestones/04-runtime-hardening.md)
5. [05 Ops Expansion](milestones/05-ops-expansion.md)

## Deliverables

- richer social graph queries and relationship metadata
- stronger channel governance for chat
- deeper guild progression and activity lifecycle rules
- stronger worker, queue, and runtime resilience
- richer operator visibility and release-readiness tooling

## Dependencies

- `v1.0` remains the stable baseline for all `v2` work
- chat and guild deepening both depend on the shipped membership-aware channel foundation
- runtime hardening depends on the durable local verification path already established in `v1`

## Deferred Items

- rich media chat attachments
- multi-region active-active routing
- formal event bus adoption
- full GM console UI
