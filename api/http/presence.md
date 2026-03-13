# Presence HTTP Contract

Base purpose: gateway-facing online-state reporting and downstream lookup of short-lived runtime context.

## Health

- `GET /healthz`

## Connect

- `POST /v1/presence/connect`
- Request

```json
{
  "player_id": "p1",
  "session_id": "sess-1",
  "realm_id": "realm-1",
  "location": "lobby"
}
```

- Response `200`

```json
{
  "player_id": "p1",
  "status": "online",
  "session_id": "sess-1",
  "realm_id": "realm-1",
  "location": "lobby",
  "last_heartbeat_at": "2026-03-13T10:00:00Z",
  "last_seen_at": "2026-03-13T10:00:00Z",
  "connected_at": "2026-03-13T10:00:00Z"
}
```

## Heartbeat

- `POST /v1/presence/heartbeat`
- Request shape matches connect
- Response `200`: presence shape
- Rules
- `session_id` must match the active session

## Disconnect

- `POST /v1/presence/disconnect`
- Request

```json
{
  "player_id": "p1",
  "session_id": "sess-1"
}
```

- Response `200`: presence shape with `status = offline`
- Rules
- `session_id` must match the active session

## Get Presence

- `GET /v1/presence/{playerID}`
- Response `200`: current presence shape

## Ownership Rules

- `gateway` is expected to call connect, heartbeat, and disconnect
- downstream services should treat this surface as the source of truth for online state
