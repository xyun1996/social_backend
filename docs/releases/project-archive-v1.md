# Project Archive: V1

- Archive date: `2026-03-14`
- Project: `social_backend`
- Stage: `v1.0 completed`

## Summary

`social_backend` has completed its `v1.0` release line as a reusable Go social backend baseline for medium-light multiplayer games. The repository is in a stable handoff state, with release docs, milestone closure, and verified local durable flows in place.

## Source Of Truth

- [Current plan](../plans/current.md)
- [V1 freeze](../plans/v1/freeze.md)
- [V1 release notes](release-notes/v1.0.md)
- [V2 roadmap](../plans/v2/roadmap.md)

## V1 Delivered

- Repository governance, ADRs, milestones, tasks, and release documents
- Shared Go service bootstrap and multi-service repository structure
- `identity`: login, refresh, introspection, player-scoped session identity
- `gateway`: session handshake, resume, ack, and replay baseline
- `presence`: online state and Redis-backed runtime state
- `social`: friends, blocks, and baseline relationship reads
- `invite`: create, accept, reject, cancel, expire
- `chat`: conversations, send, ack, replay, unread summaries, resource-backed channels, guild/party membership-aware permission checks
- `guild`: create, invite, join, owner transfer, kick, announcement, governance logs, baseline growth, initial activity template skeleton
- `party`: create, invite, ready, queue lifecycle, handoff, assignment, resolution cleanup
- `ops`: player overview, guild snapshot, party snapshot, worker reads, durable summary
- `worker`: baseline job lifecycle and durable runtime behavior
- Verified local durable MySQL and Redis flows

## Verification

- `go test ./...`
- `make check-dev`
- `make test-local-durable`

## Environment Notes

- Local Redis executable path:
  - `E:\eworkspace\software\Redis-8.6.1-Windows-x64-cygwin`
- If Redis is not running, start it from that directory before durable verification.
- Windows Redis may emit `maint_notifications` fallback warnings; these do not block the current verified flows.

## Deferred To V2

- richer social graph queries, metadata, and recommendation-oriented reads
- deeper chat governance and moderation-ready channel policy
- deeper guild progression, activity lifecycle, and reward systems
- stronger worker retry, backoff, and runtime hardening
- fuller matchmaker lifecycle modeling
- richer ops and support-facing views
- production deployment automation, CI/CD, and broader infrastructure depth

## V2 Themes

1. Social Depth
2. Chat Governance
3. Guild Progression
4. Runtime Hardening
5. Ops Expansion

## Final State

`v1.0` should be treated as completed and historically stable. New implementation work should start from the `v2` roadmap instead of reopening the shipped `v1` baseline.
