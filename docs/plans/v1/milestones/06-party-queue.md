# 06 Party Queue

- Status: `done`
- Version: `v1`

## Goal

Define and implement party management and social queue orchestration while keeping combat matchmaking out of scope.

## Inputs

- Identity/session model
- Invite rules
- Presence model

## Outputs

- Party service module
- Party invite and ready-state flow
- Queue entry/exit orchestration
- Matchmaker integration boundary

## Progress Notes

- Party creation, shared invite joins, ready-state updates, leader transfer, leave, and kick flows are implemented.
- Party now owns a social queue prototype with queue join, current queue reads, and queue leave semantics.
- Active queue enrollment blocks party membership mutation until the leader leaves queue.
- Ops party snapshots now surface active queue enrollment for operator visibility.
- Party now exposes a stable queue handoff snapshot as the future external matchmaker boundary.
- Party now accepts post-handoff match assignment callbacks and locks queue ownership after assignment.
- Party now accepts match resolution callbacks that clear assignment ownership and unlock follow-up party mutations.

## Acceptance Criteria

- Party leader operations are explicit.
- Queue lifecycle is documented with reconnect and timeout behavior.
- External matchmaker integration contract is defined.

## Risks

- Queue state can become tightly coupled to combat systems if scope is not guarded.

## Completion Notes

- Party now meets the `v1` line for management, queue orchestration, handoff, assignment, and resolution cleanup.
