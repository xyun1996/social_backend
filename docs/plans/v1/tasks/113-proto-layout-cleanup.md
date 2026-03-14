# 113 Proto Layout Cleanup

- Title: Reorganize source proto files into versioned service directories and restore `buf lint`
- Status: `done`
- Version: `v1`
- Milestone: `01 Foundation`

## Background

The repository had generated Go bindings under `api/proto/<service>/v1`, but the source `.proto` files were still all placed in `api/proto/`. That layout kept generation working only when lint was skipped and left the contract tree inconsistent.

## Goal

Move source proto contracts into the same versioned service directories as their generated bindings so the tree is self-consistent and `buf lint` passes without exceptions for multi-package source layout.

## Scope

- move source proto files into `api/proto/<service>/v1/<service>.proto`
- update cross-proto imports to the new paths
- update proto inventory tests and contract tooling to scan the new layout
- refresh proto generation output and README references

## Non-Goals

- changing protobuf package names
- introducing a second API version
- redesigning the actual RPC surfaces

## Dependencies

- [093 Proto Generation Baseline](093-proto-generation-baseline.md)
- [095 Proto Check Entrypoint](095-proto-check-entrypoint.md)
- [101 Consume Generated Proto Bindings](101-consume-generated-proto-bindings.md)

## Acceptance Criteria

- `buf lint` passes without `-SkipLint`
- `make proto` or the PowerShell generation script runs without layout overrides
- proto inventory tests and contract tooling recognize the versioned source layout
- `go test ./...` passes

## Related Docs / ADRs

- [Current plan](../../current.md)
- [Proto contracts README](../../../api/proto/README.md)

## Completion Notes

- source and generated proto artifacts now live together under versioned service directories
- cross-service imports now reference versioned source paths such as `invite/v1/invite.proto`
- `buf lint` is now green on the repository default layout
