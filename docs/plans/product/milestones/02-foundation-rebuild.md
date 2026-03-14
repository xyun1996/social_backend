# 02 Foundation Rebuild

- Status: `in-progress`

## Goal

Rebuild the shared runtime layer for product use:

- configuration
- authn/authz
- request validation
- audit
- persistence boundaries
- jobs
- observability

## Success Criteria

- Product runtime no longer depends on prototype-era assumptions for wiring or auth.
- New foundation packages become the default path for `api-gateway`, `social-core`, and `ops-worker`.

## Progress

- `social-core` now has explicit product foundation contracts for authz, audit, transactions, and jobs.
- Rebuild runtime inventory is served from new `internal/app` wiring instead of command-level hardcoding.
