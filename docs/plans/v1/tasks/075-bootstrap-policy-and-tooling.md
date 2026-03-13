# 075 Bootstrap Policy And Tooling

## Goal

Normalize early schema bootstrap so initialization is explicit, repeatable, and documented before a full migration framework exists.

## Scope

- add `BOOTSTRAP_ONLY=true` mode to MySQL-backed services
- add a local bootstrap script for owned MySQL schemas
- document bootstrap policy and local usage

## Non-Goals

- introducing a production migration framework
- replacing service-owned schema boundaries
- handling Redis data seeding

## Acceptance

- MySQL-backed services can run bootstrap and exit without serving traffic
- local schema bootstrap is available as a documented explicit command
- bootstrap expectations are documented under operations
