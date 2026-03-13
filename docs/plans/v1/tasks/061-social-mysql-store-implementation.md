# 061 Social MySQL Store Implementation

## Goal

Implement the `social` MySQL repository as a real `FriendRequestStore`, `FriendshipStore`, and `BlockStore` so durable-path startup is just configuration and bootstrap.

## Scope

- implement friend request save/list/get persistence methods
- implement friendship and block persistence methods
- add schema bootstrap helper
- add repository tests

## Non-Goals

- wiring the repository into `social` runtime startup
- changing request acceptance semantics
- migration tooling beyond owned schema bootstrap

## Acceptance

- social MySQL repository satisfies the service store interfaces
- repository tests cover save/load behavior and bootstrap execution
