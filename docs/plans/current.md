# Current Plan

- Version: `production`
- Last updated: `2026-03-14`
- Source of truth level: highest

## Current Goal

Turn the completed `v1 + v2` feature line into a single-region production baseline with explicit security, observability, deployment, and operator recovery rules.

## Success Criteria

- Internal endpoints require service tokens when configured.
- Ops endpoints require bearer auth when configured.
- Shared request logging, request IDs, recovery, audit events, metrics, and mutating-request rate limiting are enabled by default.
- CI validates tests, proto checks, contract inventory, and release dry-run entrypoints.
- Production runbooks exist for Redis, MySQL, worker backlog, gateway disconnect storms, and chat delivery failures.
- `go test ./...`, `make check-dev`, and `make test-local-durable` stay green.

## Status

`production` hardening is active.

## Active Milestones

1. Security and trust boundaries
2. Observability and alerting baseline
3. Release and rollback baseline
4. Incident runbooks and drills
5. Load and failure validation

## Remaining Follow-ups

- Proto contracts still trail some HTTP/runtime surfaces.
- Deep moderation products, advanced worker scheduling, and multi-region rollout remain backlog items.
- Production dashboards and external alert routing still need infrastructure-side hookup.

## Key Dependencies

- [docs/plans/v2/roadmap.md](v2/roadmap.md)
- [docs/plans/production/roadmap.md](production/roadmap.md)
- [docs/plans/backlog.md](backlog.md)
- [docs/releases/project-archive-v1.md](../releases/project-archive-v1.md)

## Update Rules

- `v1` and `v2` release docs remain historical facts.
- Production hardening is a separate line and should not reopen closed feature scopes without an explicit plan update.
