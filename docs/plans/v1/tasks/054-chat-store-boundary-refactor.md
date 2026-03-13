# 054 Chat Store Boundary Refactor

## Goal

Refactor `chat` so its main service logic depends on store interfaces instead of hard-coded in-memory maps, making the next durable path integration incremental rather than structural.

## Scope

- add conversation, message, and cursor store interfaces
- move the default in-memory implementation behind those interfaces
- keep HTTP and behavior unchanged
- add tests proving injected stores are used

## Non-Goals

- full MySQL-backed chat runtime
- offline delivery persistence changes
- API contract changes

## Acceptance

- `ChatService` can be constructed with injected stores
- existing chat tests still pass
- store injection is covered by tests
