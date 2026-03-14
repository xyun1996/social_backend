# Party HTTP Contract

Base purpose: party membership, ready state, queue orchestration, and v2 runtime-hardening queue lifecycle rules.

## Health

- `GET /healthz`

## Core Flows

- `POST /v1/parties`
- `GET /v1/parties/{partyID}`
- `POST /v1/parties/{partyID}/invites`
- `POST /v1/parties/{partyID}/join`
- `POST /v1/parties/{partyID}/ready`
- `POST /v1/parties/{partyID}/leave`
- `POST /v1/parties/{partyID}/kick`
- `POST /v1/parties/{partyID}/transfer-leader`
- `GET /v1/party-memberships/{playerID}`
- `GET /v1/parties/{partyID}/ready`
- `GET /v1/parties/{partyID}/members`

## Queue Flows

- `POST /v1/parties/{partyID}/queue/join`
- `GET /v1/parties/{partyID}/queue`
- `GET /v1/parties/{partyID}/queue/handoff`
- `POST /v1/parties/{partyID}/queue/assignment`
- `GET /v1/parties/{partyID}/queue/assignment`
- `POST /v1/parties/{partyID}/queue/assignment/resolve`
- `POST /v1/parties/{partyID}/queue/leave`

Queue state now also includes:
- `expires_at`

## V2 Runtime Hardening Addition

### Sweep Expired Queue Ownership

- `POST /v1/internal/parties/queue/sweep-expired`

Response:

```json
{
  "removed_party_ids": ["party-1"],
  "removed_count": 1,
  "swept_at": "2026-03-14T12:00:00Z"
}
```

## Queue Rules

- queue joins attach an `expires_at` timeout window
- expired queue state is treated as no longer owned by the party
- sweeping expired queues clears both queue state and any stale assignment ownership
- assigned parties still require explicit resolution to complete the match lifecycle
