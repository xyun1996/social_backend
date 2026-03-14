# 01 Runtime Consolidation

- Status: `in-progress`

## Goal

Shift the target runtime from many thin prototype services to three product-facing binaries:

- `api-gateway`
- `social-core`
- `ops-worker`

## Success Criteria

- New runtime entrypoints compile and run.
- Current plan and architecture docs reference the new target topology.
- Existing prototype services are explicitly marked as frozen reference implementations.

## Progress

- `social-core` now carries all six Phase A core domains directly.
- `api-gateway` now exposes runtime/upstream visibility and proxies Phase A product routes into `social-core`.
