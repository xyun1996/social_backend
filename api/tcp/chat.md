# Chat TCP Contract

Base purpose: define the shared realtime envelope semantics for chat delivery over gateway-managed connections.

## Envelope Shape

Gateway should transport chat events in a domain-neutral envelope so future party or guild pushes can reuse the same shape.

```json
{
  "event_id": "evt-101",
  "stream": "chat",
  "kind": "chat.message",
  "conversation_id": "conv-1",
  "seq": 12,
  "sent_at": "2026-03-13T10:00:00Z",
  "payload": {
    "message_id": "msg-12",
    "sender_player_id": "player-2",
    "body": "hello"
  }
}
```

## Ordering Rule

- `chat` owns per-conversation ordering through `seq`.
- Gateway must not renumber or reorder chat events within a connection.
- Clients use `seq` and `conversation_id` to detect gaps.

## Client Ack

- Ack is per conversation, not per transport packet.
- Client acks the highest durable `seq` it has processed.
- Gateway or downstream delivery workers may batch these acks before they reach `chat`.

Example ack:

```json
{
  "type": "ack",
  "stream": "chat",
  "conversation_id": "conv-1",
  "ack_seq": 12
}
```

## Replay Handoff

- If the client reconnects and detects a sequence gap, replay responsibility belongs to `chat`.
- Gateway may carry the client's last seen event metadata, but replay data must come from `chat` using `conversation_id` and `after_seq`.
- Offline clients should rely on `chat` replay rather than gateway-local buffering as the source of truth.

## Delivery Modes

- `direct`: recipient is online and has an active runtime target from presence
- `store_and_replay`: recipient is offline or has no active target, so delivery depends on later replay

The delivery mode is planned by `chat` and consumed by gateway or later workers; gateway should not infer delivery mode on its own.
