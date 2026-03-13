# 070 Local Durable Worker Runtime Flows

## Goal

Extend the opt-in local durable integration coverage to worker-driven cross-service flows so the current MySQL and Redis durable paths are exercised together.

## Scope

- add a durable `invite -> worker -> expire` integration flow
- add a durable `chat -> worker -> offline delivery` integration flow

## Non-Goals

- background worker daemons
- distributed worker coordination
- persistent offline delivery receipts

## Acceptance

- local durable integration tests cover worker execution against durable invite state
- local durable integration tests cover worker execution against durable chat state
