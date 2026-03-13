# 091 Align Ops Proto With Durable Summary

## Goal

Bring `api/proto/ops.proto` back in line with the current operator-facing HTTP and service surfaces after the durable runtime work.

## Scope

- add the durable summary messages and RPC to `api/proto/ops.proto`
- add MySQL bootstrap and Redis runtime snapshot messages and RPCs
- align existing operator shapes with current fields
  - social pending inbox/outbox
  - player overview pending fields and counts
  - party/guild snapshot counts
  - worker snapshot count
- update the active plan references

## Acceptance

- `api/proto/ops.proto` covers the current `ops` HTTP/service read surface
- no current `ops` HTTP endpoint is missing from the proto contract baseline
- `go test ./...` passes
