# 098 Contract Inventory Checker

## Goal

Add a repository-local script that prints and validates the current HTTP, proto, and TCP contract inventory without relying on developers to infer the surface set from tests.

## Scope

- add a `check_contract_inventory` command under `scripts/dev/cmd`
- validate the current control-plane and realtime surface sets
- validate the contract README index entries
- add a `make check-contracts` target
- document the contract inventory workflow

## Acceptance

- `make check-contracts` prints the current contract inventory
- `make check-contracts` exits non-zero when the repository contract sets or README indexes drift
- the checker logic is unit-tested
