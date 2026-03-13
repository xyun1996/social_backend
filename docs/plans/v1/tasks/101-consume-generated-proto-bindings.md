# 101 Consume Generated Proto Bindings

## Status

`done`

## Goal

Move generated proto bindings from passive artifacts to active repository code by introducing the first adapter layer that imports and uses them.

## Scope

- generate Go bindings into module-local import paths
- add the first hand-written package that converts runtime shapes into generated proto messages
- document that generated bindings are now part of active repository code, not only contract output

## Acceptance

- generated bindings exist under `api/proto/<service>/v1`
- at least one hand-written package imports and uses generated bindings
- the generated packages compile under `go test`

## Completion Notes

- generated bindings are present under versioned `api/proto/<service>/v1` packages
- `services/ops/internal/protoconv` now imports and converts into generated `opsv1` messages
- generated packages compile as part of the default repository test flow
