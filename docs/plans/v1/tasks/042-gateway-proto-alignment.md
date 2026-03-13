# 042 Gateway Proto Alignment

## Goal

Align `gateway.proto` with the current executable realtime prototype so proto no longer stops at session resolution and presence forwarding.

## Scope

- add realtime session lifecycle messages
- add inbox, chat delivery, ack, and replay messages
- expand the proto service surface to match the current gateway runtime behavior

## Non-Goals

- generated bindings
- TCP frame definitions inside proto
- changing the existing HTTP routes

## Acceptance

- `gateway.proto` covers handshake, resume, heartbeat, close, session lookup, inbox reads, chat dispatch, ack, and replay
- proto semantics reflect the current gateway prototype boundaries
