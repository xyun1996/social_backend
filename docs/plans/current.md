# Current Plan

- Version: `v1`
- Last updated: `2026-03-13`
- Source of truth level: highest

## Current Goal

Advance the prototype stack into durable local runtime shape by closing MySQL and Redis startup, migration, and cross-service integration gaps.

## Success Criteria

- Existing prototype services remain runnable and documented.
- Durable-backed services can bootstrap owned state safely on local MySQL and Redis.
- Shared migration and bootstrap behavior is explicit instead of duplicated per service.
- Cross-service local durable flows remain executable after storage wiring changes.

## In Scope

- Service-owned MySQL and Redis startup wiring
- Shared bootstrap and migration behavior for durable-backed services
- Local durable integration coverage for cross-service runtime flows
- Documentation alignment between active plan, tasks, and architecture notes

## Out of Scope

- Production-ready migration orchestration beyond local service-owned bootstrap
- CI/CD pipelines beyond placeholder entrypoints
- Full deployment manifests or runtime configs
- Non-local infrastructure automation

## Active Milestones

1. Foundation scaffold
2. Identity and session prototype
3. Social graph prototype
4. Invite lifecycle prototype
5. Chat and offline messaging prototype
6. Guild system design
7. Party and queue design

## Current Risks

- MySQL migration handling is still service-owned and local-first; there is no external promotion or rollback workflow yet.
- Redis-backed runtime state still depends on service-local key ownership instead of a broader operational policy.
- HTTP and proto contracts are ahead of generated bindings, so interface drift still needs discipline.
- Without disciplined updates, future plan drift could appear between `current`, tasks, and architecture docs.

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
- [docs/plans/v1/tasks/052-social-mysql-repo-foundation.md](v1/tasks/052-social-mysql-repo-foundation.md)
- [docs/plans/v1/tasks/053-worker-redis-repo-foundation.md](v1/tasks/053-worker-redis-repo-foundation.md)
- [docs/plans/v1/tasks/054-chat-store-boundary-refactor.md](v1/tasks/054-chat-store-boundary-refactor.md)
- [docs/plans/v1/tasks/055-invite-store-boundary-refactor.md](v1/tasks/055-invite-store-boundary-refactor.md)
- [docs/plans/v1/tasks/056-chat-mysql-store-implementation.md](v1/tasks/056-chat-mysql-store-implementation.md)
- [docs/plans/v1/tasks/057-invite-mysql-store-implementation.md](v1/tasks/057-invite-mysql-store-implementation.md)
- [docs/plans/v1/tasks/058-chat-mysql-startup-bootstrap.md](v1/tasks/058-chat-mysql-startup-bootstrap.md)
- [docs/plans/v1/tasks/059-invite-mysql-startup-bootstrap.md](v1/tasks/059-invite-mysql-startup-bootstrap.md)
- [docs/plans/v1/tasks/060-social-store-boundary-refactor.md](v1/tasks/060-social-store-boundary-refactor.md)
- [docs/plans/v1/tasks/061-social-mysql-store-implementation.md](v1/tasks/061-social-mysql-store-implementation.md)
- [docs/plans/v1/tasks/062-social-mysql-startup-bootstrap.md](v1/tasks/062-social-mysql-startup-bootstrap.md)
- [docs/plans/v1/tasks/063-worker-store-boundary-refactor.md](v1/tasks/063-worker-store-boundary-refactor.md)
- [docs/plans/v1/tasks/064-worker-redis-store-implementation.md](v1/tasks/064-worker-redis-store-implementation.md)
- [docs/plans/v1/tasks/065-worker-redis-startup-integration.md](v1/tasks/065-worker-redis-startup-integration.md)
- [docs/plans/v1/tasks/066-idempotent-mysql-bootstrap.md](v1/tasks/066-idempotent-mysql-bootstrap.md)
- [docs/plans/v1/tasks/067-local-durable-run-targets.md](v1/tasks/067-local-durable-run-targets.md)
- [docs/plans/v1/tasks/068-local-durable-integration-tests.md](v1/tasks/068-local-durable-integration-tests.md)
- [docs/plans/v1/tasks/069-local-durable-gateway-auth-flow.md](v1/tasks/069-local-durable-gateway-auth-flow.md)
- [docs/plans/v1/tasks/070-local-durable-worker-runtime-flows.md](v1/tasks/070-local-durable-worker-runtime-flows.md)
- [docs/plans/v1/tasks/071-party-store-boundary-refactor.md](v1/tasks/071-party-store-boundary-refactor.md)
- [docs/plans/v1/tasks/072-party-mysql-store-and-startup.md](v1/tasks/072-party-mysql-store-and-startup.md)
- [docs/plans/v1/tasks/073-guild-store-and-mysql-startup.md](v1/tasks/073-guild-store-and-mysql-startup.md)
- [docs/plans/v1/tasks/074-gateway-redis-session-store.md](v1/tasks/074-gateway-redis-session-store.md)
- [docs/plans/v1/tasks/075-bootstrap-policy-and-tooling.md](v1/tasks/075-bootstrap-policy-and-tooling.md)
- [docs/plans/v1/tasks/076-expand-durable-runtime-coverage.md](v1/tasks/076-expand-durable-runtime-coverage.md)
- [docs/plans/v1/tasks/077-shared-mysql-migration-runner.md](v1/tasks/077-shared-mysql-migration-runner.md)
- [docs/plans/v1/tasks/078-normalize-service-owned-mysql-migrations.md](v1/tasks/078-normalize-service-owned-mysql-migrations.md)
- [docs/plans/v1/tasks/079-verify-local-mysql-migration-state.md](v1/tasks/079-verify-local-mysql-migration-state.md)
- [docs/plans/v1/tasks/080-verify-local-redis-runtime-state.md](v1/tasks/080-verify-local-redis-runtime-state.md)
- [docs/plans/v1/tasks/081-ops-mysql-bootstrap-visibility.md](v1/tasks/081-ops-mysql-bootstrap-visibility.md)
- [docs/plans/v1/tasks/082-ops-redis-runtime-visibility.md](v1/tasks/082-ops-redis-runtime-visibility.md)
- [docs/plans/v1/tasks/083-shared-durable-connection-helpers.md](v1/tasks/083-shared-durable-connection-helpers.md)
- [docs/plans/v1/tasks/084-ops-repo-reader-tests.md](v1/tasks/084-ops-repo-reader-tests.md)
- [docs/plans/v1/tasks/085-complete-local-durable-run-targets.md](v1/tasks/085-complete-local-durable-run-targets.md)
- [docs/plans/v1/tasks/086-align-config-docs-with-durable-status-flags.md](v1/tasks/086-align-config-docs-with-durable-status-flags.md)
- [docs/plans/v1/tasks/087-add-local-durable-status-check-script.md](v1/tasks/087-add-local-durable-status-check-script.md)
- [docs/plans/v1/tasks/088-add-db-helper-tests.md](v1/tasks/088-add-db-helper-tests.md)
- [docs/plans/v1/tasks/089-ops-durable-summary-endpoint.md](v1/tasks/089-ops-durable-summary-endpoint.md)
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
