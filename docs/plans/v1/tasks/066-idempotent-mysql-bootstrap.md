# 066 Idempotent MySQL Bootstrap

## Goal

Make service-owned MySQL bootstrap safe to run on repeated startup so `*_AUTO_MIGRATE=true` remains usable across local restarts.

## Scope

- make owned `CREATE TABLE` statements idempotent
- apply the change across identity, invite, chat, and social repositories
- keep existing repository bootstrap tests passing

## Non-Goals

- introducing a migration framework
- changing table ownership boundaries
- altering runtime query behavior

## Acceptance

- repeated bootstrap runs do not fail just because the owned table already exists
- existing repository tests still pass
