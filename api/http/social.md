# Social HTTP Contract

Base purpose: friend request, acceptance, friend listing, block creation, and block listing.

## Health

- `GET /healthz`

## Send Friend Request

- `POST /v1/friends/requests`
- Request

```json
{
  "from_player_id": "p1",
  "to_player_id": "p2"
}
```

- Response `200`

```json
{
  "id": "req-1",
  "from_player_id": "p1",
  "to_player_id": "p2",
  "status": "pending",
  "created_at": "2026-03-13T10:00:00Z"
}
```

- Rules
- Self-friend is rejected
- Existing pending request for the same pair is returned
- Blocked relationships reject the request

## Accept Friend Request

- `POST /v1/friends/requests/{requestID}/accept`
- Request

```json
{
  "actor_player_id": "p2"
}
```

- Response `200`: same object as request, with `status = accepted`
- Rules
- Only `to_player_id` can accept

## List Friends

- `GET /v1/friends?player_id=p1`
- Response `200`

```json
{
  "player_id": "p1",
  "friends": ["p2"]
}
```

## Block Player

- `POST /v1/blocks`
- Request

```json
{
  "player_id": "p2",
  "blocked_player_id": "p1"
}
```

- Response `200`

```json
{
  "player_id": "p2",
  "blocked_id": "p1",
  "created_at": "2026-03-13T10:00:00Z"
}
```

## List Blocks

- `GET /v1/blocks?player_id=p2`
- Response `200`

```json
{
  "player_id": "p2",
  "blocks": ["p1"]
}
```
