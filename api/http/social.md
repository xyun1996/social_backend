# Social HTTP Contract

Base purpose: friend requests, accepted friendships, blocks, richer relationship reads, and lightweight friend metadata.

## Health

- `GET /healthz`

## Core V1 Flows

- `POST /v1/friends/requests`
- `GET /v1/friends/requests?player_id=p2&role=inbox&status=pending`
- `POST /v1/friends/requests/{requestID}/accept`
- `GET /v1/friends?player_id=p1`
- `POST /v1/blocks`
- `GET /v1/blocks?player_id=p2`

## V2 Social Depth Additions

### Set Friend Remark

- `POST /v1/friends/remarks`

```json
{
  "player_id": "p1",
  "friend_id": "p2",
  "remark": "raid lead"
}
```

### List Friend Remarks

- `GET /v1/friends/remarks?player_id=p1`

### Get Relationship Snapshot

- `GET /v1/relationships/{targetID}?player_id=p1`

Response fields include:
- `state`
- `is_friend`
- `has_pending_inbox`
- `has_pending_outbox`
- `is_blocked`
- `is_blocked_by`
- `remark`
- `reverse_remark`

### List Relationship Snapshots

- `GET /v1/relationships?player_id=p1`
- Optional filter: `state=friend|pending_inbox|pending_outbox|blocked|blocked_by`

### Pending Social Summary

- `GET /v1/pending-social?player_id=p1`

```json
{
  "player_id": "p1",
  "inbox": ["p2"],
  "outbox": ["p3"],
  "inbox_count": 1,
  "outbox_count": 1,
  "total_pending": 2
}
```
