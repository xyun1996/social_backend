# Task 025 - Enqueue Chat Offline Delivery Jobs

## Context

`chat` already computes presence-aware delivery plans, and `worker` now receives real async job intent from `invite`. But chat still stops at delivery planning, so offline recipients do not yet produce any asynchronous delivery or replay follow-up intent.

## Goal

Make `chat` enqueue offline delivery jobs for recipients who resolve to `offline_replay`, without changing chat's role as the owner of message ordering.

## Scope

- Add an optional worker boundary to `chat`
- Enqueue `chat.offline_delivery` jobs for offline recipients after send
- Keep online recipients on the direct push path only
- Update docs, config examples, and tests

## Non-Goals

- Executing actual push delivery
- Building gateway-side transport dispatch
- Turning worker enqueue into a transactional requirement for message send

## Acceptance Criteria

- Sending a message with offline recipients produces worker jobs
- Sending to online recipients does not create offline delivery jobs
- Message sequencing still behaves as before

## Status

`done`
