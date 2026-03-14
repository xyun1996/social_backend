# Observability

## Logging

- Use structured logs.
- Include request, session, player, tenant, and realm correlation identifiers when available.
- Every HTTP service emits request completion logs by default.
- Required fields for production triage:
  - `service`
  - `env`
  - `request_id`
  - `trace_id`
  - `method`
  - `route`
  - `status`
  - `duration_ms`
  - `remote_addr`
- Audit-sensitive writes should emit a distinct `audit event` log line.

## Metrics

- Track connection counts, login success/failure, invite throughput, queue sizes, message send/replay counts, and error rates.
- Every HTTP service now exposes `/metrics` with:
  - inflight request count
  - per-route request counts
  - per-route response bytes
  - per-route average latency
- Production dashboards should at minimum chart:
  - `gateway` request errors and disconnect-adjacent failures
  - `chat` send/replay error rates
  - `worker` backlog, retries, and dead-letter growth
  - `ops` durable summary failures
  - MySQL and Redis connectivity failures

## Tracing

- Trace cross-service flows for login, friend actions, guild actions, message send, and queue transitions.
- Current production baseline uses propagated `X-Request-ID` and `X-Trace-ID` headers.
- External tracing backends may be added later, but all services must preserve these identifiers today.

## Alerts

- Define alerts for abnormal disconnect spikes, Redis/MySQL failures, queue backlog growth, and worker retry storms.
- Minimum single-region alert set:
  - health endpoint failures or repeated startup failures
  - Redis/MySQL unavailable
  - worker retry/dead-letter growth
  - chat send failures above baseline
  - gateway unauthorized spikes or disconnect spikes
  - ops durable summary read failures

## Ownership

- Every production-facing service should define its core health signals before launch.
- Runbooks must link to the metrics and alerts they expect operators to watch.
