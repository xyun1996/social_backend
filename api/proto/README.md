# Proto Contracts

This directory contains the first baseline for internal service-to-service contracts.

## Scope

Current proto baselines cover boundaries that already have explicit cross-service HTTP clients:

- `gateway`
- `identity`
- `social`
- `presence`
- `invite`
- `chat`
- `party`
- `guild`
- `ops`
- `worker`

## Rules

- Proto contracts should describe service boundaries, not storage schemas.
- Field additions must remain backward-compatible.
- HTTP and proto contracts should evolve together when the same boundary semantics change.
- Transport-specific concerns such as websocket frames do not belong here.

## Current Status

- These proto files are design baselines only.
- Code generation is not wired yet.
- `api/proto/` now covers all currently executable control-plane service boundaries.
- `worker.proto` now includes execution result semantics so executor and background-drain flows do not live only in HTTP docs.
