# 096 TCP Contract Smoke Tests

## Goal

Bring the realtime protocol docs under the same repository-local smoke coverage model already used for HTTP and proto contracts.

## Scope

- add tests that assert `api/tcp` covers the expected realtime surfaces
- add tests that assert `api/tcp/README.md` lists those surfaces explicitly
- keep the realtime surface set aligned with the current gateway/chat transport baseline

## Acceptance

- `go test ./...` fails if a realtime TCP contract file goes missing
- `go test ./...` fails if `api/tcp/README.md` stops listing the current realtime surfaces
