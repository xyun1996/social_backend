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

This repository currently provides the project scaffold, governance documents, and the first reusable Go service bootstrap.

- `go.mod` initializes the repository as the root Go module.
- `go.work` is pre-created as the future monorepo entrypoint.
- `pkg/app`, `pkg/config`, and `pkg/logging` provide a minimal shared runtime layer.
- `services/gateway/cmd/gateway` and `services/identity/cmd/identity` are runnable starter binaries.
- `Makefile` exposes placeholder targets so later automation has a stable home.
- Most leaf directories include `.gitkeep` placeholders so the intended shape survives in version control.

## Quick Start

1. Review [docs/plans/current.md](docs/plans/current.md)
2. Review [docs/plans/v1/roadmap.md](docs/plans/v1/roadmap.md)
3. Check active constraints and ADRs under [docs/memory](docs/memory)
4. Start creating service modules under `services/` as milestones move into implementation

## Local Runflow

Example environment files live under [configs/examples](configs/examples).

- `make run-identity`
- `make run-gateway`
- `make run-social`

Starter service defaults:

- `identity` listens on `:8081`
- `gateway` listens on `:8080`
- `social` listens on `:8082`
