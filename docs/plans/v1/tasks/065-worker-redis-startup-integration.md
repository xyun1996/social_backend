# 065 Worker Redis Startup Integration

## Goal

Wire the `worker` service startup path to optionally use the Redis-backed store so durable queue state can be enabled by configuration.

## Scope

- add `WORKER_STORE=redis` startup selection
- add startup Redis connectivity validation
- document local configuration defaults for the Redis-backed mode

## Non-Goals

- making Redis the default runtime mode
- redesigning background execution behavior
- adding queue sharding or multi-worker coordination

## Acceptance

- `worker` can still boot in memory mode without Redis
- `worker` can boot against Redis when configured and fails fast on connectivity errors
- local example configuration documents the Redis-backed mode
