# Chat HTTP Contract

Base purpose: conversation creation, message sequencing, read acknowledgement, and replay.

## Health

- `GET /healthz`

## Create Conversation

- `POST /v1/conversations`
- Request

```json
{
  "kind": "private",
  "resource_id": "",
  "member_player_ids": ["p1", "p2"]
}
```

- Response `200`

```json
{
  "id": "conv-1",
  "kind": "private",
  "resource_id": "",
  "member_player_ids": ["p1", "p2"],
  "last_seq": 0,
  "created_at": "2026-03-13T10:00:00Z"
}
```

- Rules
- Supported `kind`: `private`, `group`, `guild`, `party`, `world`, `system`, `custom`
- `private` requires exactly 2 distinct members
- `system` requires at least 1 member

## List Conversations

- `GET /v1/conversations?player_id=p1`
- Response `200`

```json
{
  "player_id": "p1",
  "count": 1,
  "conversations": []
}
```

## Send Message

- `POST /v1/conversations/{conversationID}/messages`
- Request

```json
{
  "sender_player_id": "p1",
  "body": "hello"
}
```

- Response `200`

```json
{
  "id": "msg-1",
  "conversation_id": "conv-1",
  "seq": 1,
  "sender_player_id": "p1",
  "body": "hello",
  "created_at": "2026-03-13T10:00:00Z"
}
```

- Rules
- `seq` is monotonic within a conversation
- `system` conversations only allow sender `system`
- Other kinds require sender membership
- When the worker boundary is configured, offline recipients also enqueue `chat.offline_delivery` job intent

## Replay Messages

- `GET /v1/conversations/{conversationID}/messages?player_id=p2&after_seq=1&limit=50`
- Response `200`

```json
{
  "conversation_id": "conv-1",
  "player_id": "p2",
  "after_seq": 1,
  "count": 1,
  "messages": []
}
```

- Rules
- Returns messages with `seq > after_seq`
- `limit` defaults to service default and is capped by service max

## Ack Conversation

- `POST /v1/conversations/{conversationID}/ack`
- Request

```json
{
  "player_id": "p2",
  "ack_seq": 2
}
```

- Response `200`

```json
{
  "conversation_id": "conv-1",
  "player_id": "p2",
  "ack_seq": 2,
  "updated_at": "2026-03-13T10:00:00Z"
}
```

- Rules
- `ack_seq` cannot exceed `last_seq`
- Ack cursor is monotonic and never moves backward

## Delivery Plan

- `GET /v1/conversations/{conversationID}/delivery?sender_player_id=p1`
- Response `200`

```json
{
  "conversation_id": "conv-1",
  "sender_player_id": "p1",
  "count": 1,
  "targets": [
    {
      "player_id": "p2",
      "presence": "online",
      "delivery_mode": "online_push",
      "session_id": "sess-2",
      "realm_id": "realm-1",
      "location": "lobby"
    }
  ]
}
```

- Rules
- Only valid conversation senders can request delivery planning
- Online members are marked `online_push`
- Missing or offline presence falls back to `offline_replay`

## Internal Offline Delivery Processing

- `POST /v1/internal/offline-deliveries`
- Request

```json
{
  "conversation_id": "conv-1",
  "message_id": "msg-1",
  "recipient_player": "p2",
  "delivery_mode": "offline_replay"
}
```

- Response `200`: offline delivery receipt
