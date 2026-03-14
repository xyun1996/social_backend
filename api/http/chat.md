# Chat HTTP Contract

Base purpose: conversation creation, sequencing, replay, read acknowledgements, and governance-aware policy reads.

## Health

- `GET /healthz`

## Core Conversation Flows

- `POST /v1/conversations`
- `GET /v1/conversations?player_id=p1`
- `GET /v1/conversation-summaries?player_id=p1`
- `POST /v1/conversations/{conversationID}/messages`
- `GET /v1/conversations/{conversationID}/messages?player_id=p1&after_seq=0&limit=50`
- `POST /v1/conversations/{conversationID}/ack`
- `GET /v1/conversations/{conversationID}/summary?player_id=p1`
- `GET /v1/conversations/{conversationID}/delivery?sender_player_id=p1`

## Policy Surface

- `GET /v1/conversations/{conversationID}/channel`
- `GET /v1/conversations/{conversationID}/governance`

Governance/channel responses now include:
- `send_policy`
- `visibility_policy`
- `moderation_mode`
- `moderator_ids`
- `muted_player_ids`

## V2 Chat Governance Additions

### Update Governance

- `POST /v1/conversations/{conversationID}/governance`

```json
{
  "actor_player_id": "system",
  "send_policy": "moderated",
  "visibility_policy": "public_read"
}
```

Supported policies:
- `send_policy`: `members`, `moderated`, `system_only`
- `visibility_policy`: `members`, `public_read`

### Set Moderator

- `POST /v1/conversations/{conversationID}/moderators`

```json
{
  "actor_player_id": "system",
  "target_player_id": "p1",
  "enabled": true
}
```

### Set Mute

- `POST /v1/conversations/{conversationID}/mutes`

```json
{
  "actor_player_id": "system",
  "target_player_id": "p2",
  "enabled": true
}
```

## Governance Rules

- `system` channels remain `system_only`
- `world` defaults to `moderated`
- `world` and `custom` may use `public_read`
- muted senders cannot publish messages
- moderated channels only allow `system` or listed moderators to send
- `guild` and `party` channels still enforce backing resource membership for current members

## Internal Offline Delivery Processing

- `POST /v1/internal/offline-deliveries`
