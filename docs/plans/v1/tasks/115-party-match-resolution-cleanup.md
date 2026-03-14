# 115 Party Match Resolution Cleanup

- Title: Add post-assignment queue resolution cleanup for party orchestration
- Status: `done`
- Version: `v1`
- Milestone: `06 Party Queue`

## Background

The party service could already accept a match assignment callback, but once assigned there was no service-owned way to clear assignment ownership after the match handoff was consumed or cancelled.

## Goal

Add a minimal resolution flow that lets party clear the current queue assignment using the active ticket and reopen the party for normal mutations after the external matchmaker handoff is finished.

## Scope

- add a queue resolution domain model
- expose a resolution endpoint that accepts `ticket_id`, `match_id`, and `status`
- validate the active assignment before clearing queue ownership
- add a player-scoped membership lookup for downstream reads
- update HTTP and proto contracts
- add service and handler tests

## Non-Goals

- full post-match result ingestion
- queue timeout workers
- reconnect-aware reservation recovery

## Dependencies

- [111 Party Match Assignment Callback Boundary](111-party-match-assignment-callback-boundary.md)
- [06 Party Queue milestone](../milestones/06-party-queue.md)
- [Party HTTP contract](../../../api/http/party.md)

## Acceptance Criteria

- assigned parties can resolve the active assignment using the current ticket
- successful resolution clears queue ownership and unlocks leave/member mutation flows
- player-scoped party lookups return the current party snapshot for ops aggregation
- `go test ./...` passes

## Related Docs / ADRs

- [Current plan](../../current.md)
- [Architecture overview](../../../architecture/overview.md)

## Completion Notes

- party queue resolution now clears active assignment state after a handoff is consumed or cancelled
- queue ownership unlocks after resolution, so leave and other mutations can proceed again
- party now exposes a stable player-membership read for downstream ops aggregation
