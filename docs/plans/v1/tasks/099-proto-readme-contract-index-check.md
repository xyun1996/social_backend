# 099 Proto README Contract Index Check

## Goal

Bring `api/proto/README.md` under the same repository-local index validation as the HTTP and TCP contract directories.

## Scope

- add a proto README smoke test
- extend the contract inventory checker to validate the proto README service list
- update the active plan references

## Acceptance

- `go test ./...` fails if `api/proto/README.md` stops listing one of the current runtime services
- `make check-contracts` fails if the proto README service index drifts from the expected control-plane set
