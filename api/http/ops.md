# Ops HTTP Contract

## Health

- `GET /healthz`

## Player Overview

- `GET /v1/ops/players/{playerID}`

## Presence / Party / Guild / Worker Reads

- `GET /v1/ops/presence/{playerID}`
- `GET /v1/ops/parties/{partyID}`
- `GET /v1/ops/guilds/{guildID}`
- `GET /v1/ops/workers`

## Durable Reads

- `GET /v1/ops/bootstrap/mysql`
- `GET /v1/ops/runtime/redis`
- `GET /v1/ops/durable/summary`

## Guild Snapshot Additions In V2.0

`GET /v1/ops/guilds/{guildID}` now also includes:
- `level`
- `experience`
- `next_level_xp`
- `contributions`
- `activity_instances`
- `reward_records`

Example additions:

```json
{
  "guild_id": "guild-1",
  "level": 2,
  "experience": 125,
  "next_level_xp": 200,
  "contributions": [
    {
      "player_id": "p1",
      "total_xp": 125,
      "last_source_type": "donate",
      "updated_at": "2026-03-14T10:10:00Z"
    }
  ],
  "activity_instances": [
    {
      "id": "inst-1",
      "template_key": "sign_in",
      "period_key": "2026-03-14",
      "status": "active",
      "starts_at": "2026-03-14T00:00:00Z",
      "ends_at": "2026-03-15T00:00:00Z"
    }
  ],
  "reward_records": [
    {
      "id": "reward-1",
      "player_id": "p1",
      "activity_id": "act-1",
      "template_key": "sign_in",
      "reward_type": "badge",
      "reward_ref": "guild_sign_in"
    }
  ]
}
```
