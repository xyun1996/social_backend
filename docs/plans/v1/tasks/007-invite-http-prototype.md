# 007 Invite HTTP Prototype

- Title: Add an in-memory invite prototype for cross-domain invitation flows
- Status: `done`
- Version: `v1`
- Milestone: `05 Guild`

## Background

Guild and party milestones both depend on shared invite semantics, but the repository still lacks an executable invitation lifecycle outside of milestone notes.

## Goal

Add a local in-memory `invite` service prototype that exercises invite creation, acceptance, decline, expiration, and inbox or outbox listing.

## Scope

- Add invite domain models
- Add in-memory invite lifecycle logic with TTL handling
- Add HTTP endpoints and tests for create, respond, and list flows
- Add a runnable `invite` service entrypoint
- Align the active plan with the move from scaffold work into invite/chat prototyping

## Non-Goals

- Persistence
- Push fanout to gateway or presence
- Domain-specific guild or party side effects
- Invite cancellation and operator moderation flows

## Dependencies

- [003 Identity HTTP Prototype](003-identity-http-prototype.md)
- [004 Wire Gateway Session Introspection](004-wire-gateway-session-introspection.md)
- [005 Social Graph HTTP Prototype](005-social-graph-http-prototype.md)
- [05 Guild milestone](../milestones/05-guild.md)
- [06 Party Queue milestone](../milestones/06-party-queue.md)

## Acceptance Criteria

- Invites can be created and listed by inbox or outbox role
- Invited players can accept or decline pending invites
- Expired invites transition to `expired` and cannot be accepted
- `go test ./...` passes

## Related Docs / ADRs

- [Current plan](../../current.md)
- [04 Chat Offline milestone](../milestones/04-chat-offline.md)
- [Constraints](../../../memory/constraints.md)
