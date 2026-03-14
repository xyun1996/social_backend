# 04 Staging Release Readiness

- Status: `in-progress`

## Goal

Turn the rebuilt Phase A package into something that can survive staging validation and release rehearsal.

## Success Criteria

- Release and rollback are rehearsable.
- Support and audit flows exist for the Phase A package.
- Load and fault drills are tied to the rebuilt runtime, not just the prototype runtime.

## Progress

- `ops-worker` now exposes product runtime overview across `api-gateway` and `social-core`.
- A minimal support repair endpoint now exists for Phase A sync workflows and dry-run rehearsal.
