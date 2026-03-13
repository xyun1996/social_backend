# Task 034 - MySQL Repo Foundation

## Context

The architecture and persistence docs now describe MySQL ownership, and local MySQL defaults are documented, but the repository still lacks any shared MySQL configuration helper or service-local repo/mysql skeleton to anchor future durable state work.

## Goal

Add the first MySQL repository foundation with shared DSN configuration and an `identity`-local repo/mysql skeleton.

## Scope

- Add shared MySQL config and DSN helpers under `pkg/db`
- Add `services/identity/internal/repo/mysql` foundation code
- Update docs and example config notes

## Non-Goals

- Opening real DB connections in runtime code
- Migrations
- Replacing the in-memory identity service

## Acceptance Criteria

- The repo has a reusable MySQL config helper
- `identity` has a service-local repo/mysql skeleton with explicit schema ownership
- Tests cover the shared DSN helper

## Status

`done`
