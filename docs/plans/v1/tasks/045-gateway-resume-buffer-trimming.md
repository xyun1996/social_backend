# 045 Gateway Resume Buffer Trimming

## Status

`done`

## Goal

Make `last_server_event_id` affect gateway-local runtime buffering during session resume so reconnect behavior is consistent with the documented realtime contract.

## Scope

- trim buffered gateway events through the acknowledged event id during resume
- preserve the existing chat-owned replay handoff for durable gap recovery
- add realtime service and handler tests
- document the rule in HTTP and TCP gateway contracts

## Non-Goals

- durable event storage in gateway
- replay synthesis during resume
- trimming non-matching events when the event id is unknown

## Acceptance

- resume removes buffered events through `last_server_event_id` when present in the session inbox
- unmatched event ids leave the inbox unchanged
- `go test ./services/gateway/...` passes

## Completion Notes

- gateway resume now trims buffered session events through the matched `last_server_event_id` before returning the active session
- unmatched event ids leave the session inbox untouched, preserving the replay handoff to chat
- realtime service and handler tests cover both the matched and unmatched resume paths
