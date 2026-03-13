# 069 Local Durable Gateway Auth Flow

## Goal

Extend the opt-in local durable integration coverage to the login and handshake path by combining MySQL-backed identity with Redis-backed presence behind gateway.

## Scope

- add a durable identity testkit
- add a local durable integration test for `identity -> gateway -> presence`

## Non-Goals

- durable gateway session storage
- chat delivery in the same test
- multi-node gateway coordination

## Acceptance

- local durable integration coverage includes a real login followed by gateway handshake
- the durable handshake test verifies presence state written through the current gateway path
