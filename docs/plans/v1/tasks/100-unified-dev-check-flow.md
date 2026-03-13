# 100 Unified Dev Check Flow

## Goal

Add one repository-local command that runs the default developer verification flow instead of requiring people to remember separate test, proto, and contract commands.

## Scope

- add a `dev-check` script under `scripts/dev`
- add `make check-dev`
- run `go test`, `proto-check`, and `check_contract_inventory`
- support opt-in strict modes for required `buf` and local durable gating
- document the developer check workflow

## Acceptance

- `make check-dev` runs the default repository verification flow
- the flow does not require local MySQL or Redis by default
- strict `buf` and local durable modes are available through script flags
