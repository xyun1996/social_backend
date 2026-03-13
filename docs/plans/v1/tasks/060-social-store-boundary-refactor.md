# 060 Social Store Boundary Refactor

## Goal

Refactor the `social` service around explicit persistence interfaces so later durable-path work can swap storage implementations without rewriting friendship and block semantics.

## Scope

- introduce `FriendRequestStore`, `FriendshipStore`, and `BlockStore`
- move the in-memory implementation behind those interfaces
- update service tests to exercise injected stores

## Non-Goals

- changing the HTTP surface
- changing the existing friend request or block semantics
- wiring MySQL into service startup

## Acceptance

- `social` service logic no longer depends directly on in-memory maps
- in-memory behavior remains unchanged for current tests
- service tests cover the injected store path
