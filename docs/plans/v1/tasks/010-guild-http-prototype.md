# 010 Guild HTTP Prototype

- Title: Add an in-memory guild prototype wired to the shared invite service
- Status: `in-progress`
- Version: `v1`
- Milestone: `05 Guild`

## Background

The repository now has executable invite, chat, and party prototypes, but guild still has no runnable flow to validate organization ownership or join semantics against the shared invite boundary.

## Goal

Add a local in-memory `guild` service prototype that exercises guild creation, owner-scoped invite issuance through the `invite` service boundary, and join-after-acceptance.

## Scope

- Add guild domain models for guild and member role state
- Add in-memory guild service logic for create, invite, and join flows
- Add an HTTP client for the shared invite service boundary
- Add HTTP endpoints and tests for guild lifecycle operations
- Add a runnable `guild` service entrypoint

## Non-Goals

- Guild role management beyond owner/member
- Progression, activity, rewards, or audit logs
- Presence-aware online governance features
- Durable persistence

## Dependencies

- [007 Invite HTTP Prototype](007-invite-http-prototype.md)
- [05 Guild milestone](../milestones/05-guild.md)
- [Architecture overview](../../../architecture/overview.md)

## Acceptance Criteria

- Guild owners can create a guild and issue guild invites
- Join requires an accepted invite returned by the invite boundary
- Created guilds expose owner and member role state
- `go test ./...` passes

## Related Docs / ADRs

- [Current plan](../../current.md)
- [Bounded contexts](../../../architecture/bounded-contexts.md)
