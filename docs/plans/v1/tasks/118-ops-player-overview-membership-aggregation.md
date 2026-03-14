# 118 Ops Player Overview Membership Aggregation

- Title: Expand ops player overview with current party and guild membership reads
- Status: `done`
- Version: `v1`
- Milestone: `05 Guild`

## Background

Ops player overview already aggregated presence and social state, but it still could not answer the most common runtime support question: “what party and guild is this player currently in?”

## Goal

Extend the player overview aggregate so ops can read the player's current party, guild, guild role, and queue status from one endpoint.

## Scope

- add party and guild membership readers to ops aggregation
- enrich player overview with current party ID, guild ID, guild role, and queue status
- add party and guild player-membership HTTP client reads
- align ops proto and HTTP contracts
- add service and handler tests

## Non-Goals

- historical membership timelines
- guild activity contribution breakdowns
- party match history

## Dependencies

- [026 Ops Player Overview Aggregation](026-ops-player-overview-aggregation.md)
- [114 Ops Guild Governance Visibility](114-ops-guild-governance-visibility.md)
- [Ops HTTP contract](../../../api/http/ops.md)

## Acceptance Criteria

- `GET /v1/ops/players/{playerID}/overview` includes current party and guild membership when present
- queue status is surfaced when the current party is queued or assigned
- ops proto conversion includes the new overview fields
- `go test ./...` passes

## Related Docs / ADRs

- [Current plan](../../current.md)
- [02 Identity Session milestone](../milestones/02-identity-session.md)

## Completion Notes

- ops player overview now aggregates current party and guild membership from the party and guild services
- queue status is included when the player's current party is in a queue lifecycle
- player overview now exposes current guild role alongside the guild identifier
