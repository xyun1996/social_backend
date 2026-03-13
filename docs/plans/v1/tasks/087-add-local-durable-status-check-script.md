# 087 Add Local Durable Status Check Script

## Goal

Add a single local script entrypoint that queries the `ops` durable status endpoints and prints the current MySQL and Redis snapshots.

## Scope

- add a small dev script under `scripts/dev/cmd`
- add a `Makefile` target for the script
- document the new local check shortcut

## Non-Goals

- service startup orchestration
- interactive dashboards

## Acceptance

- developers can query local durable status with one command
- README and local runflow docs mention the new shortcut
