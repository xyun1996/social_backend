# 043 Integration Local Flow Tests

## Goal

Add cross-service local integration coverage for the flows that now span multiple HTTP boundaries and worker execution paths.

## Scope

- cover `invite -> worker -> expire`
- cover `chat -> worker -> offline delivery`
- cover `gateway -> chat delivery -> session inbox`
- use `httptest` servers and real service HTTP clients instead of isolated stubs

## Non-Goals

- docker-compose or external infra
- browser or websocket integration
- performance or load testing

## Acceptance

- integration tests execute the full local flow for the three target paths
- `go test ./...` remains green with the new integration package
