# 107 Ops Party Queue Visibility

## Status

`done`

## Goal

Expose party queue enrollment through the existing ops party snapshot so operator reads can see whether a party is currently queued without making a second direct party call.

## Scope

- extend the ops-facing party snapshot with optional queue state
- have the ops party client read `/v1/parties/{partyID}/queue` in addition to member state
- treat missing queue state as an empty optional field rather than an error
- align HTTP and proto contracts with the expanded party snapshot

## Non-Goals

- queue analytics or list-all-queued-party reads
- matchmaker inspection APIs
- changing party queue write semantics

## Acceptance

- `GET /v1/ops/parties/{partyID}` includes active queue state when present
- parties without active queue state still return a valid snapshot with `queue` omitted
- `go test ./services/ops/... ./services/integration ...` passes

## Completion Notes

- ops party snapshots now surface the active queue state produced by the party service
- missing queue state is treated as a normal empty condition, not a read failure
- HTTP and proto contracts now describe the optional queue field on party snapshots
