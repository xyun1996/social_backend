# Data Model Guidelines

## Identifier Strategy

- Plan for identifiers to carry or be associated with:
  - `tenant_id`
  - `region_id`
  - `realm_id`
  - `account_id`
  - `player_id`
- Choose a stable ID generation strategy before service code is implemented.

## Common Fields

Use consistent metadata fields where applicable:

- `created_at`
- `updated_at`
- `deleted_at` for soft-delete capable entities
- `created_by`
- `updated_by`

## Table and Index Conventions

- Use singular or plural naming consistently; define the final rule in `docs/memory/conventions.md`.
- Name indexes predictably to support maintenance.
- Reserve routing fields early even if physical sharding is deferred.

## Data Placement Principles

- MySQL for durable relational state
- Redis for hot state, presence, queue metadata, and short-lived messaging caches
- Service ownership for durable versus hot state is detailed in `docs/architecture/persistence.md`

## Candidate Durable Entities

- `identity`: account, player, refresh token lineage
- `social`: friend_request, friendship_edge, block_edge
- `invite`: invite
- `chat`: conversation, conversation_member, message, read_cursor
- `party`: party, party_member
- `guild`: guild, guild_member, guild_role_assignment
- `worker`: job_intent when durable retries are required

## Audit Requirements

- Sensitive or operator-driven actions should be auditable.
- Domain models should not assume audit trails can be reconstructed from logs alone.
