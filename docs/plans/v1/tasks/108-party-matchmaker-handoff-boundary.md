# 108 Party Matchmaker Handoff Boundary

## Status

`done`

## Goal

Define an executable handoff boundary between `party` queue state and a future external matchmaker so queue integration can evolve without leaking internal party implementation details.

## Scope

- add a queue handoff snapshot owned by `party`
- expose an HTTP and proto surface for reading that handoff while the party is queued
- include queue metadata, leader identity, member ids, and resolved member runtime state
- cover the handoff flow with service, handler, and integration tests

## Non-Goals

- implementing an actual matchmaker service
- match assignment callbacks or ready-check resolution after match creation
- queue analytics or ticket search APIs

## Acceptance

- queued parties can produce a stable handoff payload for an external consumer
- the handoff payload includes the active queue, leader, members, and a deterministic ticket id
- `go test ./services/party/... ./services/integration/...` passes

## Completion Notes

- `party` now exposes `/v1/parties/{partyID}/queue/handoff` as the future matchmaker boundary
- handoff payloads include queue metadata plus resolved member runtime state
- integration coverage now proves a queued party can be turned into a handoff snapshot without introducing a real matchmaker dependency
