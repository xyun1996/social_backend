# 048 Identity MySQL Startup Bootstrap

## Goal

Turn the optional MySQL-backed `identity` store into a safer startup path by validating connectivity and optionally applying the owned schema.

## Scope

- ping MySQL on startup when `IDENTITY_STORE=mysql`
- add an explicit `IDENTITY_AUTO_MIGRATE` switch
- add repository bootstrap helpers and unit tests for statement application
- update example config and runflow notes

## Non-Goals

- production migration tooling
- rollback logic
- changing the in-memory default path

## Acceptance

- `identity` fails fast if the configured MySQL connection is unavailable
- `IDENTITY_AUTO_MIGRATE=true` applies `identity`-owned schema statements at startup
- repository tests cover schema application ordering and failure behavior
