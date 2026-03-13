# Identity HTTP Contract

Base purpose: login, refresh, and token introspection for player-scoped sessions.

## Health

- `GET /healthz`
- Response `200`

```json
{
  "service": "identity",
  "status": "ok"
}
```

## Login

- `POST /v1/auth/login`
- Request

```json
{
  "account_id": "acc-1",
  "player_id": "player-1"
}
```

- Response `200`

```json
{
  "access_token": "access-token",
  "refresh_token": "refresh-token",
  "account_id": "acc-1",
  "player_id": "player-1"
}
```

- Rules
- `account_id` and `player_id` are required
- Creates a player-scoped token pair

## Refresh

- `POST /v1/auth/refresh`
- Request

```json
{
  "refresh_token": "refresh-token"
}
```

- Response `200`: same shape as login
- Rules
- `refresh_token` is required

## Introspect

- `POST /v1/auth/introspect`
- Request

```json
{
  "access_token": "access-token"
}
```

- Response `200`

```json
{
  "account_id": "acc-1",
  "player_id": "player-1"
}
```

- Rules
- `access_token` is required
