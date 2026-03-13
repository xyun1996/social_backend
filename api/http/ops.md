# Ops HTTP Contract

Base purpose: operator-facing read queries across runtime-aware service boundaries.

## Health

- `GET /healthz`

## Player Presence

- `GET /v1/ops/players/{playerID}/presence`
- Response `200`: presence snapshot shape

## Player Overview

- `GET /v1/ops/players/{playerID}/overview`
- Response `200`

```json
{
  "player_id": "p1",
  "presence": {
    "player_id": "p1",
    "status": "online"
  },
  "friends": ["p2"],
  "blocks": ["p3"],
  "pending_inbox": ["p4"],
  "pending_outbox": ["p5"],
  "friend_count": 1,
  "block_count": 1,
  "pending_inbox_count": 1,
  "pending_outbox_count": 1
}
```

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

## MySQL Bootstrap Snapshot

- `GET /v1/ops/bootstrap/mysql`
- Response `200`

```json
{
  "count": 2,
  "services": [
    {
      "service": "chat",
      "count": 1,
      "migration_ids": ["001_chat_core"]
    },
    {
      "service": "invite",
      "count": 1,
      "migration_ids": ["001_invite_core"]
    }
  ]
}
```
