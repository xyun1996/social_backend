# Runtime Flows

## Purpose

This document captures the main request and state flows that matter in the current prototype stage. It focuses on sequence, ownership, and interaction points rather than transport implementation details.

## 1. Login And Session Resolution

```mermaid
sequenceDiagram
    participant C as Client
    participant I as identity
    participant G as gateway

    C->>I: POST /v1/auth/login
    I-->>C: access_token + refresh_token + player context
    C->>G: GET /v1/session/me (Bearer token)
    G->>I: introspect(access_token)
    I-->>G: account_id + player_id
    G-->>C: authenticated subject
```

### Notes

- identity owns token issuance and refresh
- gateway does not parse token meaning by itself
- authenticated subject becomes the basis for downstream requests

## 2. Presence Reporting

```mermaid
sequenceDiagram
    participant C as Client
    participant G as gateway
    participant I as identity
    participant P as presence

    C->>G: POST /v1/session/presence/connect
    G->>I: introspect(Bearer token)
    I-->>G: player_id
    G->>P: connect(player_id, session_id, realm_id, location)
    P-->>G: presence snapshot
    G-->>C: presence snapshot

    C->>G: POST /v1/session/presence/heartbeat
    G->>P: heartbeat(player_id, session_id, realm_id, location)
    P-->>G: updated snapshot

    C->>G: POST /v1/session/presence/disconnect
    G->>P: disconnect(player_id, session_id)
    P-->>G: offline snapshot
```

### Notes

- gateway owns authenticated attribution
- presence owns online or offline truth
- downstream runtime checks should read from presence, not from gateway memory

## 3. Shared Invite Flow

```mermaid
sequenceDiagram
    participant A as Actor
    participant D as Domain Service
    participant V as invite
    participant B as Invitee

    A->>D: create domain invite intent
    D->>V: POST /v1/invites
    V-->>D: pending invite
    D-->>A: invite created

    B->>V: accept or decline invite
    V-->>B: updated invite state
    B->>D: join with invite_id
    D->>V: GET /v1/invites/{id}
    V-->>D: accepted invite
    D-->>B: joined domain resource
```

### Notes

- invite is reused by guild and party
- domain services own membership, invite owns acceptance state

## 4. Party Join Flow

```mermaid
flowchart LR
    Leader["Party leader"] --> CreateParty["Create party"]
    CreateParty --> InviteParty["Issue party invite"]
    InviteParty --> InviteSvc["invite"]
    Invitee["Invited player"] --> AcceptInvite["Accept invite"]
    AcceptInvite --> InviteSvc
    Invitee --> JoinParty["Join with invite_id"]
    JoinParty --> Party["party"]
    Party --> InviteSvc
    Party --> Ready["Ready state updates"]
```

### Notes

- party leader is the only inviter in the current prototype
- join requires an already-accepted invite
- ready state is owned by party, not invite

## 5. Guild Join Flow

```mermaid
flowchart LR
    Owner["Guild owner"] --> CreateGuild["Create guild"]
    CreateGuild --> Guild["guild"]
    Owner --> GuildInvite["Issue guild invite"]
    GuildInvite --> Invite["invite"]
    Member["Invited player"] --> Accept["Accept invite"]
    Accept --> Invite
    Member --> Join["Join guild with invite_id"]
    Join --> Guild
    Guild --> Invite
```

### Notes

- current prototype uses owner-only invitation
- role model is intentionally minimal: `owner` and `member`

## 6. Chat Sequencing And Replay

```mermaid
sequenceDiagram
    participant S as Sender
    participant C as chat
    participant R as Receiver

    S->>C: send message
    C->>C: increment conversation last_seq
    C-->>S: message(seq=n)

    R->>C: ack(conversation_id, ack_seq)
    C->>C: store monotonic read cursor
    C-->>R: updated cursor

    R->>C: replay(after_seq)
    C-->>R: messages where seq > after_seq
```

### Notes

- ordering is local to each conversation
- replay is based on seq windows, not wall-clock timestamps
- future gateway push should rely on chat-owned sequencing

## Cross-Flow Implementation Details

- identity is upstream of almost every player-facing flow
- presence is upstream of future runtime-sensitive flows
- invite is upstream of shared join authorization
- chat is isolated from invite and membership mutation, but depends on stable player context
