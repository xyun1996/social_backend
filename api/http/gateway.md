# Gateway HTTP Contract

Base purpose: expose authenticated session context to control-plane clients.

## Health

- `GET /healthz`
- Response `200`

```json
{
  "service": "gateway",
  "status": "ok"
}
```

## Session Me

- `GET /v1/session/me`
- Required header

```text
Authorization: Bearer <access_token>
```

- Response `200`

```json
{
  "account_id": "acc-1",
  "player_id": "player-1"
}
```

- Rules
- Bearer token is required
- Gateway resolves the token through the identity introspection boundary

## Presence Connect

- `POST /v1/session/presence/connect`
- Required header

```text
Authorization: Bearer <access_token>
```

- Request

```json
{
  "session_id": "sess-1",
  "realm_id": "realm-1",
  "location": "lobby"
}
```

- Response `200`: presence snapshot from the presence service
- Rules
- `player_id` is derived from the authenticated subject, not the request body
- Gateway forwards the update to the presence boundary

## Presence Heartbeat

- `POST /v1/session/presence/heartbeat`
- Header and request shape match connect
- Response `200`: presence snapshot

## Presence Disconnect

- `POST /v1/session/presence/disconnect`
- Header and request shape match connect
- Response `200`: presence snapshot

## Realtime Handshake Prototype

- `POST /v1/realtime/handshake`
- Request

```json
{
  "access_token": "token-1",
  "session_id": "sess-1",
  "realm_id": "realm-1",
  "location": "lobby",
  "client_version": "dev"
}
```

- Response `200`: gateway-owned realtime session snapshot

## Realtime Resume Prototype

- `POST /v1/realtime/resume`
- Request

```json
{
  "access_token": "token-1",
  "session_id": "sess-1",
  "last_server_event_id": "evt-42"
}
```

- Response `200`: refreshed realtime session snapshot
- Rules
- Resume is only allowed for the original authenticated subject

## Realtime Heartbeat Prototype

- `POST /v1/realtime/sessions/{sessionID}/heartbeat`
- Response `200`: realtime session snapshot

## Realtime Close Prototype

- `POST /v1/realtime/sessions/{sessionID}/close`
- Response `200`: realtime session snapshot with `state = closed`

## Realtime Session Lookup

- `GET /v1/realtime/sessions/{sessionID}`
- Response `200`: realtime session snapshot

## Realtime Chat Dispatch Prototype

- `POST /v1/realtime/chat/deliveries`
- Request

```json
{
  "conversation_id": "conv-1",
  "sender_player_id": "p1",
  "message_id": "msg-1",
  "seq": 1,
  "body": "hello",
  "sent_at": "2026-03-13T10:00:00Z"
}
```

- Response `200`: pushed and deferred delivery summary
- Rules
- Gateway resolves targets through the chat delivery planning boundary
- `online_push` targets are written into active gateway session inboxes
- `offline_replay` targets remain deferred for replay or worker follow-up

## Realtime Session Events

- `GET /v1/realtime/sessions/{sessionID}/events`
- Response `200`

```json
{
  "session_id": "sess-2",
  "count": 1,
  "events": []
}
```

## Realtime Chat Ack Prototype

- `POST /v1/realtime/sessions/{sessionID}/acks`
- Request

```json
{
  "conversation_id": "conv-1",
  "ack_seq": 3
}
```

- Response `200`
- Rules
- Gateway resolves the `player_id` from the active session before forwarding the ack to chat

## Realtime Chat Replay Prototype

- `GET /v1/realtime/sessions/{sessionID}/replay?conversation_id=conv-1&after_seq=3&limit=50`
- Response `200`

```json
{
  "session_id": "sess-2",
  "conversation_id": "conv-1",
  "player_id": "p2",
  "after_seq": 3,
  "count": 1,
  "messages": []
}
```

- Rules
- Gateway resolves the `player_id` from the active session before requesting replay from chat
- Replay remains chat-owned; gateway only scopes and forwards the request
