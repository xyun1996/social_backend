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

## Audit Requirements

- Sensitive or operator-driven actions should be auditable.
- Domain models should not assume audit trails can be reconstructed from logs alone.
