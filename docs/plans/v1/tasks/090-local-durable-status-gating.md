# 090 Local Durable Status Gating

## Status

`done`

## Goal

Turn the local durable status checker into a real gate that can fail fast when the expected MySQL and Redis topology is not visible through `ops`.

## Scope

- extend `scripts/dev/cmd/check_local_durable_status` to read typed durable summary data
- support env-driven expectations for required MySQL and Redis summaries
- support env-driven expected MySQL service names
- make `make check-local-durable-status` enforce the full local durable topology by default
- add unit coverage for config loading and summary validation
- update local runflow documentation

## Acceptance

- the status checker exits non-zero when required durable readers are missing
- the status checker exits non-zero when expected MySQL services are absent from the bootstrap snapshot
- `make check-local-durable-status` is opinionated for the full local topology but can still be relaxed with env overrides
- `go test ./...` passes

## Completion Notes

- `make check-local-durable-status` now succeeds on the validated local setup and fails when durable readers or required MySQL services are absent
- the gate is strict by default but still supports env overrides for partial topologies
