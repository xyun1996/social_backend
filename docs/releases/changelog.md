# Changelog

## Unreleased

- Initialized repository scaffold for the Social Backend project
- Added governance documents, milestone structure, ADR baseline, and doc templates
- Verified local durable MySQL bootstrap, durable status gating, and durable integration test flow
- Fixed Windows `make` durable targets by routing them through PowerShell wrappers
- Marked generated proto consumption and gateway replay alignment tasks as completed in project docs
- Added party leave, kick, and transfer-leader operations
- Added party social queue join, leave, and queue-state reads with ready/online validation
- Added ops visibility for active party queue state
- Added party queue handoff snapshots as the future matchmaker integration boundary
- Added party match assignment callbacks and durable queue assignment snapshots
- Added a resource-backed chat channel model with reusable built-in channel bindings
- Added guild owner-managed announcement updates with durable storage support
- Added guild governance logs with durable storage support
- Added guild kick and transfer-owner operations
- Added invite cancellation and aligned HTTP/task docs with gateway ack compaction and resume trimming flows
