# Task 027 - Align Ops Proto Contract

## Context

`ops` has continued to grow on the HTTP side with worker queue visibility and player overview aggregation, but `api/proto/ops/v1/ops.proto` still only describes the earlier presence, party, and guild reads. That drift violates the contract-first rule that HTTP and proto baselines should evolve together when they represent the same service semantics.

## Goal

Update `api/proto/ops/v1/ops.proto` so it reflects the current operator-facing read surface.

## Scope

- Add social and player overview messages to `api/proto/ops/v1/ops.proto`
- Add worker snapshot messages to `api/proto/ops/v1/ops.proto`
- Add RPCs for player overview and worker snapshot reads

## Non-Goals

- Wiring code generation
- Replacing HTTP in ops
- Adding every future operator query

## Acceptance Criteria

- `ops.proto` covers the current HTTP read semantics for presence, player overview, party, guild, and worker snapshots

## Status

`done`
