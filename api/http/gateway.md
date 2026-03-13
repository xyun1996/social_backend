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
