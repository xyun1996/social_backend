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

- These proto files are now wired to a repository-local `buf` generation baseline.
- Generated Go bindings belong under `gen/proto/go`.
- `api/proto/` now covers all currently executable control-plane service boundaries.
- `worker.proto` now includes execution result semantics so executor and background-drain flows do not live only in HTTP docs.
- `gateway.proto` now tracks the realtime prototype surface, including handshake, inbox delivery, ack, and replay handoff.

## Generation

- `buf.yaml` defines the module and lint rules for `api/proto/`.
- `buf.gen.yaml` defines Go and gRPC-Go output into `gen/proto/go`.
- `make proto` runs `scripts/dev/proto-generate.ps1`, which lints first and then generates bindings.
- `powershell -ExecutionPolicy Bypass -File ./scripts/dev/proto-generate.ps1 -LintOnly` runs lint only.

## Notes

- The current environment may still need a local `buf` installation; the repository now defines the generation baseline even when the CLI is not present.
- Generated bindings are intentionally kept out of the hand-written service packages so contract and transport code stay separable.
