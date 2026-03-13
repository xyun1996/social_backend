# Task 026 - Ops Player Overview Aggregation

## Context

`ops` can already read player presence and separate party, guild, and worker snapshots, but operator workflows still need multiple calls to understand a player's social runtime context. `social` already exposes friend and block reads, so `ops` should aggregate them with presence.

## Goal

Add a player-centric overview endpoint in `ops` that aggregates presence plus social relationship state.

## Scope

- Add a social read client for `ops`
- Expose `GET /v1/ops/players/{playerID}/overview`
- Return presence, friends, and blocks in one operator-facing payload
- Update docs, example config, and tests

## Non-Goals

- Party or guild membership lookup by player
- Mutating social state through ops
- Audit trail or historical snapshots

## Acceptance Criteria

- Ops exposes a player overview endpoint
- The overview aggregates presence and social reads through explicit service boundaries
- Ops tests cover the new endpoint and service behavior

## Status

`done`
