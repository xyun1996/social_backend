# Party HTTP Contract

Base purpose: party creation, shared-invite-based joins, ready state updates, core leader/member management, and social queue orchestration.

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

## Leave Party

- `POST /v1/parties/{partyID}/leave`
- Request

```json
{
  "actor_player_id": "p2"
}
```

- Response `200`: updated party shape
- Rules
- Only existing members can leave
- The current leader must transfer leadership before leaving when other members remain

## Kick Member

- `POST /v1/parties/{partyID}/kick`
- Request

```json
{
  "actor_player_id": "p1",
  "target_player_id": "p2"
}
```

- Response `200`: updated party shape
- Rules
- Only `leader_id` can kick members
- Leaders cannot kick themselves through this endpoint; use leave semantics instead
- Kicked members have their ready state removed from the party snapshot

## Transfer Leader

- `POST /v1/parties/{partyID}/transfer-leader`
- Request

```json
{
  "actor_player_id": "p1",
  "target_player_id": "p2"
}
```

- Response `200`: updated party shape
- Rules
- Only the current `leader_id` can transfer leadership
- The transfer target must already be a party member

## Join Queue

- `POST /v1/parties/{partyID}/queue/join`
- Request

```json
{
  "actor_player_id": "p1",
  "queue_name": "casual-2v2"
}
```

- Response `200`

```json
{
  "party_id": "party-1",
  "queue_name": "casual-2v2",
  "status": "queued",
  "joined_by": "p1",
  "joined_at": "2026-03-13T10:00:00Z"
}
```

- Rules
- Only the current `leader_id` can join queue
- All current members must already be online and `is_ready = true`
- Rejoining the same queue is idempotent
- Joining a different queue while already queued is rejected

## Get Queue State

- `GET /v1/parties/{partyID}/queue`
- Response `200`: queue state shape from join response

## Get Queue Handoff

- `GET /v1/parties/{partyID}/queue/handoff`
- Response `200`

```json
{
  "ticket_id": "ticket:party-1:casual-2v2:1760000000",
  "party_id": "party-1",
  "queue_name": "casual-2v2",
  "leader_id": "p1",
  "member_ids": ["p1", "p2"],
  "joined_at": "2026-03-13T10:00:00Z",
  "member_count": 2,
  "members": []
}
```

- Rules
- Only queued parties can produce a handoff snapshot
- The handoff payload is the stable boundary intended for a future external matchmaker
- Matchmaker ownership starts after consuming this snapshot; party still owns queue state and membership rules

## Leave Queue

- `POST /v1/parties/{partyID}/queue/leave`
- Request

```json
{
  "actor_player_id": "p1"
}
```

- Response `200`

```json
{
  "party_id": "party-1",
  "queue_name": "casual-2v2",
  "status": "left",
  "left_at": "2026-03-13T10:05:00Z"
}
```

- Rules
- Only the current `leader_id` can leave queue
- Member-changing operations and ready toggles are rejected while the party has an active queue state

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
