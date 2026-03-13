# 078 Normalize Service-Owned MySQL Migrations

## Goal

Move `identity`, `chat`, `invite`, `social`, `party`, and `guild` onto versioned service-owned MySQL migrations.

## Scope

- add migration lists to each MySQL repository
- keep existing schema ownership intact while routing bootstrap through the shared runner
- update bootstrap policy and current plan documentation

## Non-Goals

- changing service store selection
- introducing a separate migration binary

## Acceptance

- each MySQL repository exposes ordered migration metadata
- service bootstrap uses recorded migration ids instead of raw statement loops
- docs mention the shared `schema_migrations` ownership model
