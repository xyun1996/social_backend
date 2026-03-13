# Task 032 - Chat Offline Replay Delivery Consumer

## Context

`chat` already enqueues `chat.offline_delivery` jobs and `worker` can execute registered handlers, but there is still no chat-side consumer that records or processes those jobs. Offline delivery intent exists without any execution-side trace.

## Goal

Make `worker` consume `chat.offline_delivery` jobs and invoke a chat-owned internal boundary that records offline delivery processing.

## Scope

- Add a chat internal endpoint for offline delivery processing
- Add a worker-side chat client and job handler
- Register the handler in worker startup
- Update docs, config examples, and tests

## Non-Goals

- Final push retry semantics
- Durable delivery receipts
- Client-visible notification transport

## Acceptance Criteria

- Worker can execute `chat.offline_delivery` jobs end to end
- Chat records offline delivery processing in an observable in-memory view
- Tests cover the new handler and chat-side processing path

## Status

`done`
