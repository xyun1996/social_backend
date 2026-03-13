# Gateway TCP Contract

Base purpose: define the realtime session lifecycle managed by `gateway`.

## Connection Phases

1. Client opens a TCP stream to gateway.
2. Client sends a handshake envelope carrying an access token and client metadata.
3. Gateway resolves the subject through the identity boundary.
4. Gateway reports `connect` to presence.
5. Gateway starts accepting realtime envelopes for that session.

## Handshake Request

- Direction: client to gateway
- Semantic fields

```json
{
  "type": "handshake",
  "request_id": "req-1",
  "access_token": "token-1",
  "session_id": "sess-1",
  "realm_id": "realm-1",
  "location": "lobby",
  "client_version": "dev"
}
```

## Handshake Response

- Direction: gateway to client

```json
{
  "type": "handshake_ack",
  "request_id": "req-1",
  "account_id": "acc-1",
  "player_id": "player-1",
  "session_id": "sess-1",
  "presence_state": "online"
}
```

## Heartbeat

- Client sends heartbeat periodically on the active session.
- Gateway forwards heartbeat to presence using the authenticated subject from the session.
- Missing heartbeats transition the session toward disconnect handling.

Example heartbeat:

```json
{
  "type": "heartbeat",
  "session_id": "sess-1"
}
```

## Resume

- Resume is allowed only for the same authenticated subject.
- Resume should reuse the existing session identity when possible.
- Gateway must refresh presence ownership before accepting resumed traffic.
- If `last_server_event_id` matches buffered session events, gateway should trim buffered delivery through that event before continuing.

Example resume request:

```json
{
  "type": "resume",
  "access_token": "token-1",
  "session_id": "sess-1",
  "last_server_event_id": "evt-42"
}
```

## Disconnect

- Gateway reports disconnect to presence when the client explicitly closes or the session times out.
- Disconnect should preserve enough metadata for ops and runtime reads to understand the last active location.
