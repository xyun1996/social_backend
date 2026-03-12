# 005 Social Graph HTTP Prototype

- Title: Add an in-memory social graph prototype for friend and block flows
- Status: `done`
- Version: `v1`
- Milestone: `03 Social Graph`

## Background

Identity and gateway now define the first authenticated session flow, but the social graph milestone still has no executable relationship state machine.

## Goal

Add a local in-memory `social` service prototype that exercises friend request, acceptance, friend listing, block listing, and block-based request rejection.

## Scope

- Add friend request and friendship domain models
- Add block relationship model
- Add in-memory service logic for friend and block flows
- Add HTTP endpoints and tests for the core social graph lifecycle
- Add a runnable `social` service entrypoint

## Non-Goals

- Persistence
- Relationship fanout events
- Friend deletion
- Notes, tags, or recommendation logic

## Dependencies

- [003 Identity HTTP Prototype](003-identity-http-prototype.md)
- [004 Wire Gateway Session Introspection](004-wire-gateway-session-introspection.md)
- [03 Social Graph milestone](../milestones/03-social-graph.md)

## Acceptance Criteria

- Friend requests can be created and accepted
- Accepted requests create bidirectional friendships
- Block relationships prevent point-to-point friend requests
- `go test ./...` passes

## Related Docs / ADRs

- [Current plan](../../current.md)
- [Constraints](../../../memory/constraints.md)
