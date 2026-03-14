# Gateway Disconnect Storm Triage

## Scenario

Large numbers of gateway session requests or reconnect attempts start failing or returning unauthorized responses.

## Triggers

- Gateway error rate spikes
- Client reconnect storm
- Presence session visibility drops unexpectedly

## Checks

1. Confirm `/healthz`, `/readyz`, and `/metrics` on gateway are reachable.
2. Check recent gateway request logs for:
   - `status >= 500`
   - `unauthorized`
   - latency spikes
3. Confirm identity introspection and presence base URLs are reachable.
4. Confirm Redis connectivity for gateway session state if `GATEWAY_STORE=redis`.

## Recovery

1. If identity is unavailable, restore identity first and keep gateway in degraded retry mode.
2. If Redis is unavailable, restart or fail over Redis and watch gateway `/metrics` inflight/error recovery.
3. If request throttling is too aggressive, temporarily lower `*_HTTP_RATE_LIMIT_RPS` only after capturing the incident state.
4. Validate reconnect flow with a single login plus `/v1/session/me`.

## Exit Criteria

- Gateway error rate returns to baseline
- Session handshake and resume succeed
- Presence updates begin flowing again
