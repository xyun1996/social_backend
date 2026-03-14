# 002 Guild Chat Integration

- Status: `completed`
- Milestone: `V2.0-M2`

## Scope

- Publish guild governance and progression system events into guild chat and expose the resulting state through ops and local durable tests.

## Validation

- `go test ./services/chat/... ./services/worker/...`
- `make test-local-durable`
