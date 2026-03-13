# Party HTTP Contract

Base purpose: party creation, shared-invite-based joins, and ready state updates.

## Health

- `GET /healthz`

## Create Party

- `POST /v1/parties`
- Request

```json
{
  "leader_id": "p1"
}
```

- Response `200`

```json
{
  "id": "party-1",
  "leader_id": "p1",
  "member_ids": ["p1"],
  "created_at": "2026-03-13T10:00:00Z"
}
```

## Get Party

- `GET /v1/parties/{partyID}`
- Response `200`: party shape from create response

## Create Party Invite

- `POST /v1/parties/{partyID}/invites`
- Request

```json
{
  "actor_player_id": "p1",
  "to_player_id": "p2"
}
```

- Response `200`: shared invite shape
- Rules
- Only `leader_id` can invite

## Join Party

- `POST /v1/parties/{partyID}/join`
- Request

```json
{
  "invite_id": "inv-1",
  "actor_player_id": "p2"
}
```

- Response `200`: updated party shape
- Rules
- Invite must belong to `domain = party`
- Invite must target this `partyID`
- Invite must already be `accepted`

## Set Ready

- `POST /v1/parties/{partyID}/ready`
- Request

```json
{
  "actor_player_id": "p2",
  "is_ready": true
}
```

- Response `200`

```json
{
  "party_id": "party-1",
  "player_id": "p2",
  "is_ready": true,
  "updated_at": "2026-03-13T10:00:00Z"
}
```

- Rules
- Only online members can update ready state in the current prototype

## List Ready States

- `GET /v1/parties/{partyID}/ready`
- Response `200`

```json
{
  "party_id": "party-1",
  "count": 2,
  "ready_states": []
}
```

## List Members

- `GET /v1/parties/{partyID}/members`
- Response `200`

```json
{
  "party_id": "party-1",
  "count": 2,
  "members": [
    {
      "player_id": "p1",
      "is_leader": true,
      "is_ready": false,
      "presence": "online",
      "session_id": "sess-1",
      "realm_id": "realm-1",
      "location": "lobby"
    }
  ]
}
```
