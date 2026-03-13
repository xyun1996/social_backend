# 064 Worker Redis Store Implementation

## Goal

Implement the `worker` Redis repository as a real `JobStore` so durable queue state can back the current worker execution semantics.

## Scope

- persist job snapshots in Redis
- keep an ordered Redis index for deterministic listing and claiming
- add repository tests

## Non-Goals

- redesigning worker retry or backoff semantics
- adding distributed claim locks beyond the current prototype rules
- changing job handler contracts

## Acceptance

- worker Redis repository satisfies the service store interface
- repository tests cover save/load/list behavior
- job ordering remains stable by created time then id
