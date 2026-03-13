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
- Proto and realtime transport baseline documentation
- Persistence boundary documentation for MySQL and Redis rollout
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
- Internal proto coverage now spans all runnable control-plane services, but generated clients and transport binding are still deferred.
- Storage integration is still design-only, so repo and migration work remain ahead of the current prototype layer.
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
- [docs/plans/v1/tasks/021-complete-proto-contract-coverage.md](v1/tasks/021-complete-proto-contract-coverage.md)
- [docs/plans/v1/tasks/022-document-tcp-realtime-contract-baseline.md](v1/tasks/022-document-tcp-realtime-contract-baseline.md)
- [docs/plans/v1/tasks/023-document-persistence-boundaries.md](v1/tasks/023-document-persistence-boundaries.md)
- [docs/plans/v1/tasks/024-enqueue-invite-expiry-jobs.md](v1/tasks/024-enqueue-invite-expiry-jobs.md)
- [docs/plans/v1/tasks/025-enqueue-chat-offline-delivery-jobs.md](v1/tasks/025-enqueue-chat-offline-delivery-jobs.md)
- [docs/plans/v1/tasks/026-ops-player-overview-aggregation.md](v1/tasks/026-ops-player-overview-aggregation.md)
- [docs/plans/v1/tasks/027-align-ops-proto-contract.md](v1/tasks/027-align-ops-proto-contract.md)
- [docs/plans/v1/tasks/028-gateway-realtime-session-prototype.md](v1/tasks/028-gateway-realtime-session-prototype.md)
- [docs/plans/v1/tasks/029-chat-realtime-delivery-prototype.md](v1/tasks/029-chat-realtime-delivery-prototype.md)
- [docs/plans/v1/tasks/030-worker-job-executor-prototype.md](v1/tasks/030-worker-job-executor-prototype.md)
- [docs/plans/v1/tasks/031-invite-worker-expiry-consumer.md](v1/tasks/031-invite-worker-expiry-consumer.md)
- [docs/plans/v1/tasks/032-chat-offline-replay-delivery-consumer.md](v1/tasks/032-chat-offline-replay-delivery-consumer.md)
- [docs/plans/v1/tasks/033-social-request-reads-and-ops-expansion.md](v1/tasks/033-social-request-reads-and-ops-expansion.md)
- [docs/plans/v1/tasks/034-mysql-repo-foundation.md](v1/tasks/034-mysql-repo-foundation.md)
- [docs/plans/v1/tasks/035-redis-presence-repo-foundation.md](v1/tasks/035-redis-presence-repo-foundation.md)
- [docs/plans/v1/tasks/036-worker-background-runner.md](v1/tasks/036-worker-background-runner.md)
- [docs/plans/v1/tasks/037-gateway-chat-ack-prototype.md](v1/tasks/037-gateway-chat-ack-prototype.md)
- [docs/plans/v1/tasks/038-chat-replay-resume-alignment.md](v1/tasks/038-chat-replay-resume-alignment.md)
- [docs/plans/v1/tasks/039-identity-mysql-repo-integration.md](v1/tasks/039-identity-mysql-repo-integration.md)
- [docs/plans/v1/tasks/040-presence-redis-repo-integration.md](v1/tasks/040-presence-redis-repo-integration.md)
- [docs/plans/v1/tasks/041-worker-proto-alignment.md](v1/tasks/041-worker-proto-alignment.md)
- [docs/plans/v1/tasks/042-gateway-proto-alignment.md](v1/tasks/042-gateway-proto-alignment.md)
- [docs/plans/v1/tasks/043-integration-local-flow-tests.md](v1/tasks/043-integration-local-flow-tests.md)
- [docs/plans/v1/tasks/044-gateway-ack-inbox-compaction.md](v1/tasks/044-gateway-ack-inbox-compaction.md)
- [docs/plans/v1/tasks/045-gateway-resume-buffer-trimming.md](v1/tasks/045-gateway-resume-buffer-trimming.md)
- [docs/plans/v1/tasks/046-integration-gateway-ack-compaction-flow.md](v1/tasks/046-integration-gateway-ack-compaction-flow.md)
- [docs/plans/v1/tasks/047-integration-gateway-resume-trim-flow.md](v1/tasks/047-integration-gateway-resume-trim-flow.md)
- [docs/plans/v1/tasks/048-identity-mysql-startup-bootstrap.md](v1/tasks/048-identity-mysql-startup-bootstrap.md)
- [docs/plans/v1/tasks/049-presence-redis-startup-validation.md](v1/tasks/049-presence-redis-startup-validation.md)
- [docs/plans/v1/tasks/050-invite-mysql-repo-foundation.md](v1/tasks/050-invite-mysql-repo-foundation.md)
- [docs/plans/v1/tasks/051-chat-mysql-repo-foundation.md](v1/tasks/051-chat-mysql-repo-foundation.md)
- [api/http/README.md](../api/http/README.md)
- [api/tcp/README.md](../api/tcp/README.md)
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
