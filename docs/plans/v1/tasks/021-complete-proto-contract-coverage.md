# Task 021 - Complete Proto Contract Coverage

## Context

The repository already has HTTP contract baselines for `gateway` and `social`, and the executable prototypes are using those boundaries. However, `api/proto/` still does not cover those two services, which leaves the internal contract baseline incomplete and keeps `api/proto/README.md` out of sync with the actual contract-first direction.

## Goal

Complete the first-round proto contract coverage by adding `api/proto/gateway/v1/gateway.proto` and `api/proto/social/v1/social.proto`, then update plan and contract documentation so current status matches the repository.

## Scope

- Add `api/proto/gateway/v1/gateway.proto`
- Add `api/proto/social/v1/social.proto`
- Update `api/proto/README.md`
- Update `docs/plans/current.md`

## Non-Goals

- Wiring protobuf generation
- Replacing existing HTTP clients with generated code
- Adding websocket or realtime transport contracts

## Acceptance Criteria

- `api/proto/` covers every currently executable control-plane service boundary
- Proto README no longer describes already-completed work as future work
- Current plan references the expanded proto contract coverage

## Status

`done`
