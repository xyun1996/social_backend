# V1 Roadmap

- Version: `v1`
- Status: `planned`
- Last updated: `2026-03-12`

## Goal

Prepare a maintainable foundation for a reusable game social backend, then turn that foundation into implementable milestones for the major domains: identity, social graph, chat, guild, and party queue.

## Milestone Order

1. [01 Foundation](milestones/01-foundation.md)
2. [02 Identity Session](milestones/02-identity-session.md)
3. [03 Social Graph](milestones/03-social-graph.md)
4. [04 Chat Offline](milestones/04-chat-offline.md)
5. [05 Guild](milestones/05-guild.md)
6. [06 Party Queue](milestones/06-party-queue.md)

## Planned Deliverables

- Repository and governance scaffold
- Foundational architecture, glossary, and constraints
- ADR baseline for transport, session identity, realm model, and delivery semantics
- Milestone-level breakdown for implementation sequencing

## Dependency Notes

- Foundation unlocks all other milestones.
- Identity/session assumptions influence gateway, presence, chat, guild, and party.
- Social graph and chat can proceed in parallel once identity primitives are settled.
- Guild and party both depend on invite semantics and presence expectations.

## Deferred to Backlog

- Full production deployment automation
- Rich media chat
- Advanced moderation and anti-abuse systems
- Event bus integration
