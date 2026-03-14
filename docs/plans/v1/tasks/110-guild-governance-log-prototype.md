# 110 Guild Governance Log Prototype

## Status

`done`

## Goal

Add a minimal guild governance log so key owner and membership actions leave an inspectable history before the broader activity system exists.

## Scope

- define a guild-owned governance log entry model
- record create, join, announcement update, kick, and owner transfer events
- expose an HTTP read endpoint for guild logs
- persist governance logs in memory and MySQL-backed guild stores
- align HTTP and proto contracts with the new log surface

## Non-Goals

- full audit export
- filtering, pagination, or retention policies
- activity logs for progression systems

## Acceptance

- guild reads can return governance logs through HTTP
- key governance actions create readable log entries
- `go test ./services/guild/...` passes

## Completion Notes

- guild now records governance events as first-class guild-owned log entries
- the prototype log surface is readable through `GET /v1/guilds/{guildID}/logs`
- MySQL-backed guild storage now owns a dedicated log table and migration
