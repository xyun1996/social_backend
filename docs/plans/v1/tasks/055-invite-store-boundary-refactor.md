# 055 Invite Store Boundary Refactor

## Goal

Refactor `invite` so its main service logic depends on a store interface instead of a hard-coded in-memory map, making later MySQL integration incremental.

## Scope

- add an invite store interface
- move the default in-memory implementation behind that interface
- keep HTTP and behavior unchanged
- add tests proving injected stores are used

## Acceptance

- `InviteService` can be constructed with an injected store
- existing invite tests still pass
- store injection is covered by tests
