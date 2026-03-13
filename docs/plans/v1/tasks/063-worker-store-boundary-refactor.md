# 063 Worker Store Boundary Refactor

## Goal

Refactor the `worker` service around an explicit `JobStore` so queue execution logic no longer depends directly on in-memory job maps.

## Scope

- introduce `JobStore`
- move the in-memory implementation behind that interface
- update service tests to exercise injected store behavior

## Non-Goals

- changing worker job semantics
- changing the HTTP surface
- wiring Redis into service startup

## Acceptance

- worker execution logic no longer depends directly on in-memory storage
- current in-memory behavior remains stable
- service tests cover the injected store path
