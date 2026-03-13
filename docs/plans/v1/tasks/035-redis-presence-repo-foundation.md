# Task 035 - Redis Presence Repo Foundation

## Context

The architecture and persistence docs already mark `presence` as the authority for Redis-backed short-lived state, but the repository still lacks both a shared Redis configuration helper and a service-local `presence` repo/redis skeleton.

## Goal

Add the first Redis repository foundation with shared config helpers and a `presence`-local repo/redis skeleton.

## Scope

- Add shared Redis config helpers under `pkg/db`
- Add `services/presence/internal/repo/redis` foundation code
- Update docs and example config notes

## Non-Goals

- Real Redis client integration
- Replacing the in-memory presence service
- TTL or eviction policy implementation

## Acceptance Criteria

- The repo has reusable Redis config helpers
- `presence` has a service-local repo/redis skeleton with explicit key ownership
- Tests cover the shared Redis config helper

## Status

`done`
