# Proto Generation

## Purpose

Define the local workflow for linting and generating Go bindings from `api/proto`.

## Entry Points

- `make proto`
- `make proto-check`
- `powershell -ExecutionPolicy Bypass -File ./scripts/dev/proto-generate.ps1`
- `powershell -ExecutionPolicy Bypass -File ./scripts/dev/proto-generate.ps1 -LintOnly`
- `powershell -ExecutionPolicy Bypass -File ./scripts/dev/proto-check.ps1`

## Repository Conventions

- `buf.yaml` is the source of truth for module and lint configuration.
- `buf.gen.yaml` is the source of truth for code generation output.
- Generated Go bindings are written to `gen/proto/go`.
- Hand-written service code must not be placed under `gen/proto/go`.

## Local Requirements

- Install the `buf` CLI and ensure it is on `PATH`.
- No service code changes are required to lint or generate contracts.

## Expected Flow

1. Update one or more files under `api/proto/`.
2. Run `make proto-check`.
3. Run `make proto`.
4. Review any changes under `gen/proto/go`.
5. Commit the contract change and generated bindings together when bindings are intentionally refreshed.

## Validation Modes

- `make proto-check` always runs the repository-local smoke tests under `api/proto`.
- `make proto-check` also runs `buf lint` when `buf` is available on `PATH`.
- `powershell -ExecutionPolicy Bypass -File ./scripts/dev/proto-check.ps1 -RequireBuf` makes missing `buf` a hard failure.
- `make proto` is the stronger path that always requires `buf` because it performs generation.

## Failure Modes

- `buf is not installed or not on PATH`
  - install `buf` locally, then rerun `make proto`
  - or rerun `make proto-check` without strict mode if only smoke validation is needed
- lint failures
  - fix the `.proto` file before generating bindings
- generation output drift
  - treat generated files as build artifacts tied to the committed proto contract
