# V1 Roadmap

- Version: `v1`
- Status: `done`
- Last updated: `2026-03-14`

## Goal

Deliver a maintainable, locally durable `v1` social backend that covers the core identity, social, invite, chat, guild, party, and ops flows for a reusable game middle platform.

## Milestone Order

1. [01 Foundation](milestones/01-foundation.md)
2. [02 Identity Session](milestones/02-identity-session.md)
3. [03 Social Graph](milestones/03-social-graph.md)
4. [04 Chat Offline](milestones/04-chat-offline.md)
5. [05 Guild](milestones/05-guild.md)
6. [06 Party Queue](milestones/06-party-queue.md)

## Delivered

- Repository and governance scaffold
- Foundational architecture, glossary, constraints, and ADR baseline
- Durable local runtime for MySQL and Redis-backed services
- Contract-first HTTP/proto/TCP surfaces with generated bindings
- Core `v1` social backend modules and operator read surfaces

## Completion Notes

- `v1` is now intentionally freeze-scoped by [freeze.md](freeze.md).
- The release line is verified by `go test ./...`, `make check-dev`, and `make test-local-durable`.
- Future feature-deepening work should be scheduled through backlog and `v2`, not by reopening `v1`.

## Deferred to Backlog

- Full production deployment automation
- Rich media chat
- Advanced moderation and anti-abuse systems
- Event bus integration
