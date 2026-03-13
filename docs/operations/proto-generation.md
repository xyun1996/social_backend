# Proto Generation

## Purpose

Define the local workflow for linting and generating Go bindings from `api/proto`.

## Entry Points

- `make proto`
- `make proto-check`
- `make proto-lint`
- `powershell -ExecutionPolicy Bypass -File ./scripts/dev/proto-generate.ps1`
- `powershell -ExecutionPolicy Bypass -File ./scripts/dev/proto-generate.ps1 -LintOnly`
- `powershell -ExecutionPolicy Bypass -File ./scripts/dev/proto-generate.ps1 -SkipLint`
- `powershell -ExecutionPolicy Bypass -File ./scripts/dev/proto-check.ps1`

## Repository Conventions

- `buf.yaml` is the source of truth for module and lint configuration.
- `buf.gen.yaml` is the source of truth for code generation output.
- Generated Go bindings are written to `api/proto/<service>/v1`.
- Hand-written service code must not be placed under those generated package directories.

## Local Requirements

- Install the `buf` CLI and ensure it is on `PATH`.
- No service code changes are required to lint or generate contracts.

## Expected Flow

1. Update one or more files under `api/proto/`.
2. Run `make proto-check`.
3. Run `make proto`.
4. Review any changes under `api/proto/<service>/v1`.
5. Commit the contract change and generated bindings together when bindings are intentionally refreshed.

## Validation Modes

- `make proto-check` always runs the repository-local smoke tests under `api/proto`.
- `make proto-check` does not require `buf` and skips lint by default.
- `make proto-lint` runs the smoke tests and then requires `buf lint` to pass.
- `powershell -ExecutionPolicy Bypass -File ./scripts/dev/proto-check.ps1 -RequireBuf` makes missing `buf` a hard failure.
- `make proto` is the stronger path that always requires `buf` because it performs generation.
- `powershell -ExecutionPolicy Bypass -File ./scripts/dev/proto-generate.ps1 -SkipLint` is the escape hatch for the current prototype-era proto layout when generation is needed before lint cleanup is complete.

## Failure Modes

- `buf is not installed or not on PATH`
  - install `buf` locally, then rerun `make proto`
  - or rerun `make proto-check` without strict mode if only smoke validation is needed
- current lint rules fail on prototype-era multi-package layout
  - use `-SkipLint` only when generation is needed immediately
  - treat that as temporary debt, not the steady-state workflow
- lint failures
  - fix the `.proto` file before generating bindings
- generation output drift
  - treat generated files as build artifacts tied to the committed proto contract
