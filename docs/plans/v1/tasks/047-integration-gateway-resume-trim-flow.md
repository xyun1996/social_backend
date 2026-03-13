# 047 Integration Gateway Resume Trim Flow

## Status

`done`

## Goal

Cover the new gateway resume buffer trimming rule with a real cross-service local integration flow.

## Scope

- drive `identity`, `presence`, `chat`, and `gateway` through local HTTP test servers
- verify buffered events exist before resume
- verify resume trims buffered events through `last_server_event_id`

## Acceptance

- integration coverage proves `gateway resume` trims local buffered events without inventing replay data
- `go test ./services/integration/...` passes

## Completion Notes

- local integration coverage now proves resume trims buffered gateway events through `last_server_event_id`
- the integration test keeps replay ownership with chat and verifies gateway only removes already observed buffered events
