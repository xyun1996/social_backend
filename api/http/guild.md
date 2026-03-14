# Guild HTTP Contract

Base purpose: guild creation, governance, progression, and guild-scoped activity flows.

## Health

- `GET /healthz`

## Guild Aggregate

- `POST /v1/guilds`
- `GET /v1/guilds/{guildID}`
- `GET /v1/guild-memberships/{playerID}`
- `GET /v1/guilds/{guildID}/members`
- `GET /v1/guilds/{guildID}/logs`

Guild aggregate fields now include:
- `level`
- `experience`
- `announcement`
- `announcement_updated_at`

## Governance Writes

- `POST /v1/guilds/{guildID}/announcement`
- `POST /v1/guilds/{guildID}/invites`
- `POST /v1/guilds/{guildID}/join`
- `POST /v1/guilds/{guildID}/kick`
- `POST /v1/guilds/{guildID}/transfer-owner`

Governance actions also emit guild channel system events when chat is configured.

## Progression Reads

- `GET /v1/guilds/{guildID}/progression`
- `GET /v1/guilds/{guildID}/contributions`
- `GET /v1/guilds/{guildID}/rewards`

### `GET /v1/guilds/{guildID}/progression`

```json
{
  "guild_id": "guild-1",
  "level": 2,
  "experience": 125,
  "next_level_xp": 200,
  "updated_at": "2026-03-14T10:00:00Z"
}
```

## Activity Templates

- `GET /v1/guilds/activity-templates`

Templates currently shipped in `v2.0`:
- `sign_in`
- `donate`
- `guild_task`

Each template now declares:
- `period_type`
- `max_submissions_per_period`
- `contribution_xp`
- optional reward bookkeeping defaults

## Activity Runtime

- `GET /v1/guilds/{guildID}/activities`
- `GET /v1/guilds/{guildID}/activities/{templateKey}/instances`
- `POST /v1/guilds/{guildID}/activities/{templateKey}`
- `POST /v1/guilds/{guildID}/activities/{templateKey}/submit`

### Submit Request

```json
{
  "actor_player_id": "p1",
  "idempotency_key": "guild-sign-in-2026-03-14-p1",
  "source_type": "api"
}
```

### Submit Response

```json
{
  "record": {
    "id": "act-1",
    "instance_id": "inst-1",
    "guild_id": "guild-1",
    "template_key": "sign_in",
    "player_id": "p1",
    "delta_xp": 10,
    "idempotency_key": "guild-sign-in-2026-03-14-p1",
    "source_type": "api",
    "created_at": "2026-03-14T10:10:00Z"
  },
  "guild": {
    "id": "guild-1",
    "level": 1,
    "experience": 10
  },
  "progression": {
    "guild_id": "guild-1",
    "level": 1,
    "experience": 10,
    "next_level_xp": 100,
    "updated_at": "2026-03-14T10:10:00Z"
  }
}
```

Rules:
- only current guild members can submit
- period limits are enforced per member and template
- repeated calls with the same `idempotency_key` return the previously recorded submission
- successful submissions write guild xp, member contribution, reward bookkeeping, governance log, and guild chat system events

## Internal Worker Hooks

- `POST /v1/internal/guilds/{guildID}/activities/ensure-current`
- `POST /v1/internal/guilds/{guildID}/activities/close-expired`

These endpoints are for local worker/runtime maintenance and are not part of the public game-client surface.
