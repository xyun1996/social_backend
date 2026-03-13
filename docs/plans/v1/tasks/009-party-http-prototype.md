# 009 Party HTTP Prototype

- Title: Add an in-memory party prototype wired to the shared invite service
- Status: `in-progress`
- Version: `v1`
- Milestone: `06 Party Queue`

## Background

The repository now has executable invite and chat prototypes, but party still has no runnable flow to validate leader operations or reuse of shared invite semantics.

## Goal

Add a local in-memory `party` service prototype that exercises party creation, leader-only invite issuance through the `invite` service boundary, join-after-acceptance, and member ready state.

## Scope

- Add party domain models for party membership and ready state
- Add in-memory party service logic for create, join, and ready flows
- Add an HTTP client for the shared invite service boundary
- Add HTTP endpoints and tests for party lifecycle operations
- Add a runnable `party` service entrypoint

## Non-Goals

- Queue entry or matchmaker handoff
- Presence-aware reconnect handling
- Party leave, kick, promote, or dissolve flows
- Durable persistence

## Dependencies

- [007 Invite HTTP Prototype](007-invite-http-prototype.md)
- [06 Party Queue milestone](../milestones/06-party-queue.md)
- [Architecture overview](../../../architecture/overview.md)

## Acceptance Criteria

- Party leaders can create a party and issue party invites
- Join requires an accepted invite returned by the invite boundary
- Members can update ready state after joining
- `go test ./...` passes

## Related Docs / ADRs

- [Current plan](../../current.md)
- [Bounded contexts](../../../architecture/bounded-contexts.md)
