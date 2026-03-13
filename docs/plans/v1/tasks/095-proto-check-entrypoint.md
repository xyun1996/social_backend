# 095 Proto Check Entrypoint

## Goal

Add a repository-local proto validation entrypoint that is useful even on machines where `buf` is not installed.

## Scope

- add a `proto-check` script that always runs proto smoke tests
- run `buf lint` when `buf` is available
- support a strict mode that requires `buf`
- add a `make proto-check` target
- document the difference between `proto-check` and `proto`

## Acceptance

- `make proto-check` works as a lightweight proto validation path
- missing `buf` is non-fatal by default for `proto-check`
- strict mode is available for environments that require full lint coverage
