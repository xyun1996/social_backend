# Security Baseline

## Authentication

- Control-surface login issues tokens.
- Real-time transports validate tokens before establishing player sessions.
- `IDENTITY_ACCESS_TOKEN_TTL` and `IDENTITY_REFRESH_TOKEN_TTL` now define token lifecycle defaults.
- Access-token introspection rejects expired sessions.
- Refresh rotation rejects expired refresh lineage and invalidates consumed refresh tokens.

## Authorization

- Operator actions require stronger auditability than player actions.
- Domain services should not infer privilege purely from transport origin.
- `/v1/internal/*` endpoints are protected by `APP_INTERNAL_TOKEN` when configured.
- `/v1/ops/*` endpoints are protected by `OPS_API_TOKEN` when configured.
- Public mutating HTTP routes can be throttled through shared request rate limits.

## Data Handling

- Do not commit secrets.
- Sensitive identifiers should be masked in logs where full values are unnecessary.
- Internal and operator tokens must come from runtime environment injection, never from committed config.

## Audit

- Key operator actions, guild governance changes, and sensitive account operations should be auditable.
- Shared middleware now emits `audit event` logs for internal writes, operator calls, guild governance changes, chat governance changes, and progression writes.

## Future Work

- Add explicit threat model, secret rotation policy, and moderation workflow once implementation starts.
- Add secret rotation automation, external authorization backends, and richer moderation policies in a future production phase.
