# Task 024 - Enqueue Invite Expiry Jobs

## Context

`worker` now exists as a runnable async job queue prototype and `ops` can inspect queued jobs, but no domain service is actually producing jobs yet. `invite` already owns TTL-based expiry semantics, so it is the cleanest first boundary to consume worker scheduling intent.

## Goal

Make `invite` enqueue an `invite.expire` job when a new invite is created, while preserving current lazy expiry behavior for the in-memory prototype.

## Scope

- Add an optional worker client boundary to `invite`
- Enqueue an `invite.expire` job for newly created invites
- Keep duplicate pending invites idempotent without duplicate jobs
- Update docs, example config, and tests

## Non-Goals

- Executing expiry jobs automatically
- Replacing lazy expiry checks in invite reads
- Adding delayed job scheduling semantics to worker

## Acceptance Criteria

- New invite creation enqueues a worker job when the worker boundary is configured
- Returning an existing pending invite does not enqueue another job
- Invite service tests cover the enqueue behavior

## Status

`done`
