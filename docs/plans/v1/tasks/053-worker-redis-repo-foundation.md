# 053 Worker Redis Repo Foundation

## Goal

Add a service-local Redis repository foundation for `worker` so future queue claim and backlog visibility state has an explicit owner in code.

## Scope

- add `services/worker/internal/repo/redis`
- define canonical Redis key conventions for queue and claim state
- align persistence documentation

## Acceptance

- worker has a Redis repository foundation with explicit key ownership
- persistence docs point to the new foundation path
