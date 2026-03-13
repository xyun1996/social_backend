# 038 Chat Replay Resume Alignment

## Goal

Turn the documented replay handoff between `gateway` and `chat` into an executable prototype so reconnecting sessions can request replay using gateway-owned session identity.

## Scope

- add a session-scoped gateway replay endpoint
- forward replay requests to chat using the active session `player_id`
- cover the flow with gateway handler tests
- document the HTTP contract update

## Non-Goals

- websocket or raw TCP frame handling
- replay batching or cursor compaction
- durable session buffering

## Acceptance

- `GET /v1/realtime/sessions/{sessionID}/replay` works with `conversation_id`, `after_seq`, and `limit`
- gateway rejects replay on missing or inactive sessions
- gateway never accepts caller-supplied `player_id` for replay
- `go test ./services/gateway/...` passes
