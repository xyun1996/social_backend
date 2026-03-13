# 057 Invite MySQL Store Implementation

## Goal

Implement the `invite` MySQL repository as a real `InviteStore` so later durable-path wiring is mostly startup configuration.

## Scope

- implement invite save/list/get persistence methods
- add schema bootstrap helper
- add repository tests

## Non-Goals

- wiring the repository into `invite` runtime startup
- TTL scheduler redesign
- migration tooling beyond owned schema bootstrap

## Acceptance

- invite MySQL repository satisfies the service store interface
- repository tests cover save/load, ordering, and bootstrap execution
