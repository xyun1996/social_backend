# 097 HTTP Contract README Smoke Tests

## Goal

Make the HTTP contract index itself part of repository-local verification so the control-plane surface list cannot silently drift from the actual files.

## Scope

- add tests that assert `api/http` covers the expected control-plane surfaces
- add tests that assert `api/http/README.md` lists those surfaces explicitly
- keep the surface set aligned with the current runnable prototype set

## Acceptance

- `go test ./...` fails if a control-plane HTTP contract file goes missing
- `go test ./...` fails if `api/http/README.md` stops listing one of the current control-plane surfaces
