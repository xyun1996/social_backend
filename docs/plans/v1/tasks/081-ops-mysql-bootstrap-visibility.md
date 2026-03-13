# 081 Ops MySQL Bootstrap Visibility

## Goal

Expose recorded MySQL bootstrap state through `ops` so service-owned migration progress is visible over the operator read surface.

## Scope

- add an optional `ops` MySQL bootstrap reader over `schema_migrations`
- expose `GET /v1/ops/bootstrap/mysql`
- cover the read path in unit and durable integration tests

## Non-Goals

- MySQL write operations from `ops`
- production admin auth

## Acceptance

- `ops` can return recorded migration ids when MySQL status reading is enabled
- the new read surface is documented in `api/http/ops.md`
- `go test ./...` remains green
