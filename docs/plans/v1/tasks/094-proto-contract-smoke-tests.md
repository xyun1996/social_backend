# 094 Proto Contract Smoke Tests

## Goal

Add repository-local tests that catch the most basic protocol contract drift even when local code generation tooling is not installed.

## Scope

- add tests that assert every executable runtime service has both HTTP and proto contract files
- add tests that assert every proto file declares `syntax`, `package`, and `go_package`
- keep the test baseline aligned with the current runtime service set

## Acceptance

- `go test ./...` fails if a runtime service loses its HTTP or proto contract file
- `go test ./...` fails if a proto file loses its required top-level declarations
- the smoke tests do not require `buf` to be installed
