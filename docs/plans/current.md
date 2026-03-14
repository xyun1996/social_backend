# Current Plan

- Version: `v2`
- Last updated: `2026-03-14`
- Source of truth level: highest

## Current Goal

Record `v2` as delivered. The active implementation line is complete across:
- `guild` progression and guild-chat integration (`v2.0`)
- `social` depth
- `chat` governance
- `runtime` hardening
- `ops` expansion

## Success Criteria

- Richer social relationship reads and friend remarks are available.
- World/custom/system chat governance is policy-aware and moderation-ready.
- Worker retry/backoff and party queue expiry behavior are visible and testable.
- Ops surfaces expose deeper social, queue, worker, and guild state than `v1`.
- `go test ./...`, `make check-dev`, and `make test-local-durable` stay green.

## Status

`v2` is complete.

## Delivered Milestones

1. [01 Social Depth](v2/milestones/01-social-depth.md)
2. [02 Chat Governance](v2/milestones/02-chat-governance.md)
3. [03 Guild Progression](v2/milestones/03-guild-progression.md)
4. [04 Runtime Hardening](v2/milestones/04-runtime-hardening.md)
5. [05 Ops Expansion](v2/milestones/05-ops-expansion.md)

## Remaining Follow-ups

- Proto contracts still trail some HTTP/runtime surfaces.
- Rich chat cards and deeper moderation workflows remain backlog items.
- Production-grade scheduling and rollout automation remain beyond `v2`.

## Key Dependencies

- [docs/plans/v2/roadmap.md](v2/roadmap.md)
- [docs/plans/backlog.md](backlog.md)
- [docs/releases/project-archive-v1.md](../releases/project-archive-v1.md)

## Update Rules

- `v1` and `v2` release docs remain historical facts.
- New work should start a fresh post-`v2` plan instead of widening the closed `v2` scope.
