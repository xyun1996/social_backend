# Task 037 - Gateway Chat Ack Prototype

## Context

`gateway` can now own realtime sessions and route chat events into per-session inboxes, but there is still no reverse path for client acknowledgements to flow back into `chat` read cursors.

## Goal

Add a gateway-owned chat ack path that validates the active session and forwards ack state to `chat` using the session subject.

## Scope

- Extend the gateway chat client with ack support
- Add a gateway realtime ack endpoint
- Update docs and tests

## Non-Goals

- Transport frame parsing
- Per-event ack at the gateway layer
- Durable ack storage inside gateway

## Acceptance Criteria

- Gateway can accept a chat ack for an active session
- The ack forwarded to `chat` uses the session player identity, not a client-provided player id

## Status

`done`
