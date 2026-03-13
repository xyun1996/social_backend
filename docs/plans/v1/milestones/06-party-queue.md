# 06 Party Queue

- Status: `planned`
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
- External matchmaker integration remains an explicit future boundary.

## Acceptance Criteria

- Party leader operations are explicit.
- Queue lifecycle is documented with reconnect and timeout behavior.
- External matchmaker integration contract is defined.

## Risks

- Queue state can become tightly coupled to combat systems if scope is not guarded.
