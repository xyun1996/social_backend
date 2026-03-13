# 039 Identity MySQL Repo Integration

## Goal

Move `identity` from a repo/mysql skeleton to an executable optional MySQL-backed store while preserving the existing in-memory default.

## Scope

- define store interfaces inside `identity`
- wire the MySQL repository into account and session persistence
- add runtime selection through `IDENTITY_STORE`
- keep HTTP contracts unchanged

## Non-Goals

- migrations
- guaranteed schema bootstrap
- making MySQL mandatory for local runs

## Acceptance

- `identity` still runs with the in-memory default
- `identity` can construct a MySQL-backed auth service when `IDENTITY_STORE=mysql`
- service tests still pass after the store injection refactor
