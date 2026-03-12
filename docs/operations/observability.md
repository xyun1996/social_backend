# Observability

## Logging

- Use structured logs.
- Include request, session, player, tenant, and realm correlation identifiers when available.

## Metrics

- Track connection counts, login success/failure, invite throughput, queue sizes, message send/replay counts, and error rates.

## Tracing

- Trace cross-service flows for login, friend actions, guild actions, message send, and queue transitions.

## Alerts

- Define alerts for abnormal disconnect spikes, Redis/MySQL failures, queue backlog growth, and worker retry storms.

## Ownership

- Every production-facing service should define its core health signals before launch.
