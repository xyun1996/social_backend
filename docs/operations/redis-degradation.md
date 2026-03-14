# Redis Degradation Handling

## Scenario

Redis becomes unavailable or slow, impacting presence, gateway realtime state, or worker runtime coordination.

## Triggers

- Redis ping failures
- Durable summary shows Redis runtime unavailable
- Presence or worker retries spike together

## Checks

1. Confirm Redis is reachable at the configured address.
2. Query `GET /v1/ops/runtime/redis` and `GET /v1/ops/durable/summary`.
3. Check gateway, presence, and worker logs for repeated Redis connection errors.

## Recovery

1. Restart Redis or shift traffic to the standby instance if available.
2. Restart affected services only after Redis responds to `PING`.
3. Re-run `make check-local-durable-status` or the equivalent staging status probe.

## Exit Criteria

- Redis runtime summary is healthy
- Presence and worker request failures fall back to baseline
- Gateway session and presence-dependent flows recover
