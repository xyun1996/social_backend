# Contract Inventory

## Purpose

Provide a human-friendly inventory and gate for the repository's current HTTP, proto, and TCP contract surfaces.

## Entry Points

- `make check-contracts`
- `go run ./scripts/dev/cmd/check_contract_inventory`

## What It Checks

- the HTTP contract file set under `api/http`
- the proto contract file set under `api/proto`
- the realtime TCP contract file set under `api/tcp`
- the current surface indexes in `api/http/README.md` and `api/tcp/README.md`

## Expected Surface Sets

- Control plane: `chat`, `gateway`, `guild`, `identity`, `invite`, `ops`, `party`, `presence`, `social`, `worker`
- Realtime: `chat`, `gateway`

## Failure Meaning

- unexpected HTTP or proto contract set
  - a runtime service lost or added a contract file without updating the expected repository inventory
- unexpected TCP contract set
  - realtime transport docs drifted from the current gateway/chat baseline
- missing README surface entry
  - the contract file exists, but the directory index was not updated
