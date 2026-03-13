# 072 Party MySQL Store And Startup

## Goal

Implement the `party` MySQL repository and wire it into optional startup configuration so durable party and ready-state behavior can run without rewriting service logic.

## Scope

- implement MySQL persistence for parties, members, and ready states
- add schema bootstrap helper and repository tests
- add `PARTY_STORE=mysql` startup selection
- add optional `PARTY_AUTO_MIGRATE=true` startup bootstrap

## Non-Goals

- changing invite or ready-state semantics
- gateway or queue integration changes
- migration tooling beyond owned schema bootstrap

## Acceptance

- party MySQL repository satisfies `PartyStore` and `ReadyStateStore`
- repository tests cover bootstrap and save/load behavior
- `party` can boot in either memory or MySQL mode by configuration
