# Task 033 - Social Request Reads And Ops Expansion

## Context

`social` currently supports friend request creation and acceptance, but operator and service-facing read surfaces still cannot inspect pending friend requests. `ops` player overview therefore misses a meaningful part of the player's current social state.

## Goal

Add friend request read APIs to `social`, expand `ops` player overview to include pending request state, and align `social.proto`.

## Scope

- Add friend request listing to `social`
- Expand `ops` social snapshot and player overview
- Update `social.proto`
- Update docs and tests

## Non-Goals

- Friend request rejection or cancellation
- Pagination
- Historical audit queries

## Acceptance Criteria

- `social` exposes a request list read endpoint
- `ops` player overview includes pending inbox and outbox requests
- `social.proto` reflects the new read surface

## Status

`done`
