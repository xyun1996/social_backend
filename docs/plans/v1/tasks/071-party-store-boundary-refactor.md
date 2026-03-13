# 071 Party Store Boundary Refactor

## Goal

Refactor the `party` service around explicit persistence interfaces so future durable rollout can replace storage without changing invite and ready-state logic.

## Scope

- introduce `PartyStore` and `ReadyStateStore`
- move the in-memory implementations behind those interfaces
- update service tests to exercise injected stores

## Non-Goals

- changing invite semantics
- adding MySQL storage yet
- changing presence-aware ready rules

## Acceptance

- party service logic no longer depends directly on in-memory maps
- current in-memory behavior remains unchanged
- tests cover the injected store path
