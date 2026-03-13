# 076 Expand Durable Runtime Coverage

## Goal

Extend opt-in local durable integration coverage to the newly durable `party`, `guild`, and `gateway` runtime paths.

## Scope

- add durable testkits for `party` and `guild`
- add local durable integration flows for party join/ready and guild join/member reads
- add a Redis-backed gateway persistence test across server restart

## Non-Goals

- multi-process load testing
- production failover semantics
- persistent chat history beyond current runtime checks

## Acceptance

- local durable integration tests cover `party` and `guild` durable flows
- local durable integration tests verify gateway session state survives server recreation when Redis-backed
