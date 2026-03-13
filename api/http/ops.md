# Ops HTTP Contract

Base purpose: operator-facing read queries across runtime-aware service boundaries.

## Health

- `GET /healthz`

## Player Presence

- `GET /v1/ops/players/{playerID}/presence`
- Response `200`: presence snapshot shape

## Party Snapshot

- `GET /v1/ops/parties/{partyID}`
- Response `200`

```json
{
  "party_id": "party-1",
  "count": 2,
  "members": []
}
```

## Guild Snapshot

- `GET /v1/ops/guilds/{guildID}`
- Response `200`

```json
{
  "guild_id": "guild-1",
  "count": 2,
  "members": []
}
```

## Worker Snapshot

- `GET /v1/ops/jobs?status=queued&type=invite.expire`
- Response `200`

```json
{
  "status": "queued",
  "type": "invite.expire",
  "count": 1,
  "jobs": []
}
```
