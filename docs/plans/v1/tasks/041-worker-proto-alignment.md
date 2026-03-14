# 041 Worker Proto Alignment

## Goal

Align `api/proto/worker/v1/worker.proto` with the already implemented execution surface so background-drain semantics are not documented only through HTTP.

## Scope

- add proto messages for execution requests and results
- add proto RPCs for `ExecuteNext` and `ExecuteUntilEmpty`
- document that background runner semantics now exist in the proto baseline

## Non-Goals

- generated proto bindings
- runtime gRPC transport
- process lifecycle RPCs for starting or stopping the background runner

## Acceptance

- `api/proto/worker/v1/worker.proto` describes execution requests and responses
- the proto baseline covers both queue lifecycle and executor lifecycle semantics
