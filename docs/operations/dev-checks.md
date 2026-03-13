# Dev Checks

## Purpose

Provide one repository-local command that runs the default developer verification flow.

## Entry Points

- `make check-dev`
- `powershell -ExecutionPolicy Bypass -File ./scripts/dev/dev-check.ps1`

## Default Flow

1. `go test ./...`
2. `proto-check`
3. `check_contract_inventory`

## Optional Strict Modes

- `powershell -ExecutionPolicy Bypass -File ./scripts/dev/dev-check.ps1 -RequireBuf`
  - requires `buf` to be available and makes proto lint mandatory
  - equivalent intent to adding `proto-lint` into the dev flow
- `powershell -ExecutionPolicy Bypass -File ./scripts/dev/dev-check.ps1 -RunLocalDurableGate`
  - also runs `check_local_durable_status`
  - use this only when the local durable stack is intentionally up

## Notes

- `check-dev` is the default local verification path for repository structure, contracts, generated bindings, and tests.
- `check-dev` does not require local MySQL or Redis by default.
- The durable status gate remains opt-in because it depends on running local infrastructure.
- `make check-dev` was verified successfully on Windows after the durable target wrapper update, so `make` is now the preferred entrypoint when available.
