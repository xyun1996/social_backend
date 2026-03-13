# Task 031 - Invite Worker Expiry Consumer

## Context

`invite` now enqueues `invite.expire` jobs and `worker` can execute registered handlers, but there is still no handler that actually consumes invite expiry jobs. The queue path exists without a business-side effect.

## Goal

Make `worker` consume `invite.expire` jobs and invoke an internal invite expiry boundary.

## Scope

- Add an internal expire endpoint to `invite`
- Add a worker-side invite client and job handler
- Register the handler in worker startup
- Update docs, config examples, and tests

## Non-Goals

- Durable scheduling guarantees
- Distributed locking
- Replacing lazy invite expiry checks

## Acceptance Criteria

- Worker can execute `invite.expire` jobs end to end
- Invite expiry remains idempotent
- Tests cover the new invite expiry path

## Status

`done`
