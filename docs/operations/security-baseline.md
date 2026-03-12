# Security Baseline

## Authentication

- Control-surface login issues tokens.
- Real-time transports validate tokens before establishing player sessions.

## Authorization

- Operator actions require stronger auditability than player actions.
- Domain services should not infer privilege purely from transport origin.

## Data Handling

- Do not commit secrets.
- Sensitive identifiers should be masked in logs where full values are unnecessary.

## Audit

- Key operator actions, guild governance changes, and sensitive account operations should be auditable.

## Future Work

- Add explicit threat model, secret rotation policy, and moderation workflow once implementation starts.
