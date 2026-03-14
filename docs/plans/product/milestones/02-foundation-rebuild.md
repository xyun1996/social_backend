# 02 Foundation Rebuild

- Status: `planned`

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
