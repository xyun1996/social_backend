# Social Backend

Go-based social backend scaffold for medium-light multiplayer games. This repository is organized as both:

- a codebase for services, shared packages, protocols, tooling, and tests
- a governance system for plans, architecture records, operational notes, and long-term memory

## Current Focus

The current source of truth lives in [docs/plans/current.md](docs/plans/current.md).

Read in this order when joining or resuming work:

1. [docs/plans/current.md](docs/plans/current.md)
2. [docs/memory/constraints.md](docs/memory/constraints.md)
3. [docs/memory/glossary.md](docs/memory/glossary.md)
4. relevant ADRs under [docs/memory/adr](docs/memory/adr)
5. [docs/architecture/overview.md](docs/architecture/overview.md)

## Repository Layout

```text
api/        Protocol contracts for gRPC, HTTP, TCP, and shared errors
services/   Domain services such as gateway, identity, social, chat, guild, party
pkg/        Shared infrastructure packages only
configs/    Environment-specific config templates and examples
deploy/     Local/dev deployment assets
scripts/    Workflow automation scripts
tools/      Debugging, mock, load, and fixture tooling
test/       Integration, protocol, end-to-end, and load tests
docs/       Plans, architecture, operations, releases, memory, templates
```

## Documentation Map

- Plans: [docs/plans](docs/plans)
- Architecture: [docs/architecture](docs/architecture)
- Operations: [docs/operations](docs/operations)
- Memory and ADRs: [docs/memory](docs/memory)
- Release history: [docs/releases](docs/releases)
- Templates: [docs/templates](docs/templates)

## Tooling Status

This repository currently provides the project scaffold, governance documents, and the first reusable Go service bootstrap plus early in-memory service prototypes.

- `go.mod` initializes the repository as the root Go module.
- `go.work` is pre-created as the future monorepo entrypoint.
- `pkg/app`, `pkg/config`, and `pkg/logging` provide a minimal shared runtime layer.
- `services/gateway`, `services/identity`, `services/social`, `services/invite`, `services/chat`, `services/party`, `services/guild`, and `services/presence` now have runnable starter binaries or in-memory prototypes.
- `Makefile` exposes placeholder targets so later automation has a stable home.
- Most leaf directories include `.gitkeep` placeholders so the intended shape survives in version control.

## Quick Start

1. Review [docs/plans/current.md](docs/plans/current.md)
2. Review [docs/plans/v1/roadmap.md](docs/plans/v1/roadmap.md)
3. Check active constraints and ADRs under [docs/memory](docs/memory)
4. Start creating service modules under `services/` as milestones move into implementation

## Local Runflow

Example environment files live under [configs/examples](configs/examples).

Gateway depends on both `identity` and `presence` base URLs in local runs.
Gateway now also includes an HTTP-form realtime session prototype for handshake, heartbeat, resume, and close flows.
Gateway can also prototype chat direct delivery by writing planned online events into per-session inboxes.
Shared local infrastructure defaults live in `configs/examples/local-infra.env.example`.
Chat also depends on `presence` for delivery planning in local runs.
Chat can optionally depend on `worker` for offline delivery job intent in local runs.
Party also depends on `presence` for runtime-ready checks in local runs.
Guild also depends on `presence` for member runtime views in local runs.
Ops depends on `presence`, `party`, and `guild` for operator read aggregation in local runs.
Ops also depends on `worker` for queue visibility in local runs.
Ops also depends on `social` for player overview aggregation in local runs.
Invite can optionally depend on `worker` for async expiry job intent in local runs.
Worker can depend on `invite` and `chat` for executable async job handling in local runs.
Worker also supports an optional background drain loop via `WORKER_AUTO_RUN=true`.
`pkg/db` now includes a shared MySQL foundation, and `identity` can optionally run with a MySQL-backed store via `IDENTITY_STORE=mysql`; `IDENTITY_AUTO_MIGRATE=true` applies the owned schema at startup.
`pkg/db` now also includes a shared Redis foundation, and `presence` can optionally run with a Redis-backed store via `PRESENCE_STORE=redis`.

- `make run-identity`
- `make run-gateway`
- `make run-social`
- `make run-invite`
- `make run-chat`
- `make run-party`
- `make run-guild`
- `make run-presence`
- `make run-ops`
- `make run-worker`

Starter service defaults:

- `identity` listens on `:8081`
- `gateway` listens on `:8080`
- `social` listens on `:8082`
- `invite` listens on `:8083`
- `chat` listens on `:8084`
- `party` listens on `:8085`
- `guild` listens on `:8086`
- `presence` listens on `:8087`
- `ops` listens on `:8088`
- `worker` listens on `:8089`

Local infrastructure defaults:

- MySQL: `localhost:3306`, user `root`, password `1234`, database `social_backend`
- Redis: `localhost:6379`, no username, no password, database `0`
