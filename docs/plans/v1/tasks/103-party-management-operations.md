# 103 Party Management Operations

## Goal

Extend the `party` prototype beyond create, join, and ready flows by adding core leader and membership management operations.

## Scope

- add party leave for non-leader members
- add leader-driven member kick
- add leader transfer
- cover the new behavior in service and HTTP tests

## Non-Goals

- party dissolve
- queue or matchmaker flows
- moderation or audit trails

## Acceptance

- non-leader members can leave a party
- leaders can kick members and transfer leadership
- kicked or departed members no longer appear in party membership or ready-state reads
