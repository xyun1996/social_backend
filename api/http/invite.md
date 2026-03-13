# Invite HTTP Contract

Base purpose: shared cross-domain invite lifecycle for party, guild, and future invite consumers.

## Health

- `GET /healthz`

## Create Invite

- `POST /v1/invites`
- Request

```json
{
  "domain": "party",
  "resource_id": "party-1",
  "from_player_id": "p1",
  "to_player_id": "p2",
  "ttl_seconds": 900
}
```

- Response `200`

```json
{
  "id": "inv-1",
  "domain": "party",
  "resource_id": "party-1",
  "from_player_id": "p1",
  "to_player_id": "p2",
  "status": "pending",
  "created_at": "2026-03-13T10:00:00Z",
  "expires_at": "2026-03-13T10:15:00Z"
}
```

- Rules
- `domain`, `from_player_id`, and `to_player_id` are required
- Self-invite is rejected
- Non-positive TTL falls back to the service default
- Existing pending invite for the same tuple is returned
- When the worker boundary is configured, creating a new invite also enqueues an `invite.expire` job intent

## Get Invite

- `GET /v1/invites/{inviteID}`
- Response `200`: same shape as create response
- Rules
- Pending invite may transition to `expired` during fetch

## Accept Invite

- `POST /v1/invites/{inviteID}/accept`
- Request

```json
{
  "actor_player_id": "p2"
}
```

- Response `200`: invite shape with `status = accepted`

## Decline Invite

- `POST /v1/invites/{inviteID}/decline`
- Request shape matches accept
- Response `200`: invite shape with `status = declined`

## List Invites

- `GET /v1/invites?player_id=p2&role=inbox&status=pending`
- Response `200`

```json
{
  "player_id": "p2",
  "role": "inbox",
  "status": "pending",
  "count": 1,
  "invites": []
}
```

- Rules
- `player_id` is required
- `role` defaults to `all`
- `role` must be one of `all`, `inbox`, `outbox`
