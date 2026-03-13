# 056 Chat MySQL Store Implementation

## Goal

Implement the `chat` MySQL repository as a real `ConversationStore`, `MessageStore`, and `ReadCursorStore` so the later durable-path wiring step is mostly configuration and bootstrap.

## Scope

- implement conversation, membership, message, and cursor persistence methods
- add schema bootstrap helper
- add repository tests

## Non-Goals

- wiring the repository into `chat` runtime startup
- migrations beyond the owned schema bootstrap
- offline delivery persistence changes

## Acceptance

- chat MySQL repository satisfies the service store interfaces
- repository tests cover save/load behavior and bootstrap execution
