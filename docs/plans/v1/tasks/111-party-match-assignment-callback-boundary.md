# 111 Party Match Assignment Callback Boundary

- Title: Add the post-handoff match assignment callback boundary for party queue orchestration
- Status: `done`
- Version: `v1`
- Milestone: `06 Party Queue`

## Background

The party service now exposes a stable queue handoff snapshot for an external matchmaker, but it still has no explicit callback boundary to record the moment matchmaking ownership turns into an assigned match.

## Goal

Add a minimal party-owned callback boundary that records a match assignment against the active queue handoff ticket and locks the social queue state until a later match-resolution flow exists.

## Scope

- Add a queue assignment domain model and persistence boundary
- Add a callback endpoint that accepts `ticket_id` and `match_id`
- Persist the assignment in memory and MySQL-backed flows
- Expose a read endpoint for the current assignment snapshot
- Update party HTTP and proto contracts
- Add service, handler, and repository tests

## Non-Goals

- Match completion, cancellation, or post-match cleanup
- Game-server reservation orchestration beyond a lightweight connection hint
- Queue timeout workers or reconnect recovery policy changes

## Dependencies

- [108 Party Matchmaker Handoff Boundary](108-party-matchmaker-handoff-boundary.md)
- [06 Party Queue milestone](../milestones/06-party-queue.md)
- [Party HTTP contract](../../../api/http/party.md)

## Acceptance Criteria

- Queued parties can accept a match assignment callback using the active handoff ticket
- Assignment reads return the persisted `match_id`, queue name, and connection hint
- Once assigned, queue handoff and queue-leave semantics are blocked
- `go test ./...` passes

## Related Docs / ADRs

- [Current plan](../../current.md)
- [Architecture overview](../../../architecture/overview.md)

## Completion Notes

- Party queue state now transitions from `queued` to `assigned`
- Assignment callbacks validate the active handoff ticket before persisting
- MySQL-backed party storage now owns durable queue assignment snapshots
