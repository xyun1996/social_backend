# 109 Guild Announcement Prototype

## Status

`done`

## Goal

Add a minimal guild announcement capability so guild governance has a stable, owner-managed communication surface before activity and log systems arrive.

## Scope

- extend the guild aggregate with announcement fields
- add an owner-scoped announcement update service path
- expose an HTTP endpoint for announcement updates
- persist announcement state in both memory and MySQL-backed guild stores
- align HTTP and proto contracts with the new guild shape

## Non-Goals

- announcement history
- rich text or attachment support
- member-targeted notification fanout

## Acceptance

- guild owners can update the announcement through HTTP
- `GET /v1/guilds/{guildID}` returns the latest announcement state
- `go test ./services/guild/...` passes

## Completion Notes

- guilds now persist announcement text and update timestamp as part of the core aggregate
- announcement updates are owner-scoped and returned by the existing guild read surface
- MySQL-backed guild storage now owns a dedicated announcement migration step
