# V2.0 Roadmap

## Objective

Turn guilds from a governance shell into a durable progression system with strong chat linkage and minimum operator/runtime support.

## Scope

- Guild progression and contribution tracking
- Fixed activity templates with period-bound instances
- Reward record skeletons
- Guild channel system events for governance and progression
- Ops reads for progression surfaces
- Minimal worker support for current-period initialization and expiry transitions

## Milestones

1. [01 Guild Progression](milestones/01-guild-progression.md)
2. [02 Guild Chat Integration](milestones/02-guild-chat-integration.md)

## Completed Slice

- Guild progression, contributions, instances, and reward records are implemented in memory and MySQL paths.
- Guild activities are idempotent per provided key and period-limited by template.
- Guild activity and governance events publish into guild chat as system messages.
- Ops guild snapshot includes progression, contributions, instances, and rewards.
- Worker supports minimal guild activity period maintenance handlers.
- Durable integration now includes guild progression + chat coverage.

## Remaining Follow-Ups

- Proto contracts for guild and ops should be updated to match the new HTTP/runtime surface.
- Rich guild activity cards can graduate from text-only messages in a later slice.
- Additional progression rules should stay out until a new `v2.x` scope is explicitly opened.
