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
- Generated Go bindings now live under `api/proto/<service>/v1`.
- `api/proto/` now covers all currently executable control-plane service boundaries.
- `api/proto/worker/v1/worker.proto` now includes execution result semantics so executor and background-drain flows do not live only in HTTP docs.
- `api/proto/gateway/v1/gateway.proto` now tracks the realtime prototype surface, including handshake, inbox delivery, ack, and replay handoff.
- `services/ops/internal/protoconv` now consumes generated `opsv1` bindings as the first repository-local adapter layer.

## Generation

- `buf.yaml` defines the module and lint rules for `api/proto/`.
- Source proto files now live under `api/proto/<service>/v1/<service>.proto`.
- `buf.gen.yaml` defines Go and gRPC-Go output into module-local import paths under `api/proto/<service>/v1`.
- `make proto` runs `scripts/dev/proto-generate.ps1`, which lints first and then generates bindings.
- `powershell -ExecutionPolicy Bypass -File ./scripts/dev/proto-generate.ps1 -LintOnly` runs lint only.
- `powershell -ExecutionPolicy Bypass -File ./scripts/dev/proto-generate.ps1 -SkipLint` forces generation when lint is intentionally deferred.

## Notes

- The current environment may still need a local `buf` installation; the scripts now also resolve `buf` from `GOPATH/bin` when it is not on `PATH`.
- Generated bindings now live beside the versioned source contract files, so each service contract directory stays self-contained under `api/proto/<service>/v1`.
