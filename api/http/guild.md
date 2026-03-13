# Guild HTTP Contract

Base purpose: guild creation, owner-scoped invite issuance, join via accepted invite, and basic owner governance.

## Health

- `GET /healthz`

## Create Guild

- `POST /v1/guilds`
- Request

```json
{
  "name": "Raiders",
  "owner_id": "p1"
}
```

- Response `200`

```json
{
  "id": "guild-1",
  "name": "Raiders",
  "owner_id": "p1",
  "members": [
    {
      "player_id": "p1",
      "role": "owner",
      "joined_at": "2026-03-13T10:00:00Z"
    }
  ],
  "created_at": "2026-03-13T10:00:00Z"
}
```

## Get Guild

- `GET /v1/guilds/{guildID}`
- Response `200`: guild shape from create response

## List Members

- `GET /v1/guilds/{guildID}/members`
- Response `200`

```json
{
  "guild_id": "guild-1",
  "count": 2,
  "members": [
    {
      "player_id": "p1",
      "role": "owner",
      "presence": "online",
      "session_id": "sess-1",
      "realm_id": "realm-1",
      "location": "lobby"
    }
  ]
}
```

## Create Guild Invite

- `POST /v1/guilds/{guildID}/invites`
- Request

```json
{
  "actor_player_id": "p1",
  "to_player_id": "p2"
}
```

- Response `200`: shared invite shape
- Rules
- Only `owner_id` can invite in the current prototype

## Join Guild

- `POST /v1/guilds/{guildID}/join`
- Request

```json
{
  "invite_id": "inv-1",
  "actor_player_id": "p2"
}
```

- Response `200`: updated guild shape
- Rules
- Invite must belong to `domain = guild`
- Invite must target this `guildID`
- Invite must already be `accepted`

## Kick Member

- `POST /v1/guilds/{guildID}/kick`
- Request

```json
{
  "actor_player_id": "p1",
  "target_player_id": "p2"
}
```

- Response `200`: updated guild shape
- Rules
- Only `owner_id` can kick members in the current prototype
- Owners cannot kick themselves through this endpoint; ownership must be transferred first

## Transfer Owner

- `POST /v1/guilds/{guildID}/transfer-owner`
- Request

```json
{
  "actor_player_id": "p1",
  "target_player_id": "p2"
}
```

- Response `200`: updated guild shape
- Rules
- Only the current `owner_id` can transfer ownership
- The transfer target must already be a guild member
