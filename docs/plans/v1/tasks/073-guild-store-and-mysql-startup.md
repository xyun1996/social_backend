# 073 Guild Store And MySQL Startup

## Goal

Refactor `guild` around a store boundary, implement the MySQL repository, and wire optional MySQL startup so durable guild state can run without changing membership logic.

## Scope

- introduce `GuildStore`
- move the in-memory guild map behind that interface
- implement MySQL persistence for guilds and members
- add `GUILD_STORE=mysql` startup selection
- add optional `GUILD_AUTO_MIGRATE=true` startup bootstrap

## Non-Goals

- role redesign
- queue or gateway integration changes
- migration tooling beyond owned schema bootstrap

## Acceptance

- guild service logic no longer depends directly on an in-memory map
- MySQL repository satisfies `GuildStore`
- guild can boot in either memory or MySQL mode by configuration
