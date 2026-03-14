# Worker Retry Backlog Recovery

## Scenario

Worker jobs accumulate in retry or dead-letter states faster than they are drained.

## Triggers

- Worker snapshot shows retry growth
- Dead-letter jobs start appearing for invite/chat/guild maintenance
- Retry storms correlate with Redis/MySQL or downstream service failures

## Checks

1. Query `GET /v1/ops/jobs`.
2. Separate retries by job type to isolate the failing downstream service.
3. Check worker logs for repeated `last_error` patterns.

## Recovery

1. Fix the dependent service first.
2. Use the worker run endpoints to drain a controlled subset after the dependency recovers.
3. Re-check dead-letter growth before re-enabling background drain loops.

## Exit Criteria

- Retry queue returns to normal size
- No new dead-letter accumulation
- A manual drain cycle succeeds
