# Runbooks

Use one file per meaningful operator or incident workflow once implementation begins.

## Expected Topics

- Gateway disconnect storm triage
- Redis degradation handling
- MySQL write backlog triage
- Message replay lag investigation
- Worker retry backlog recovery
- Operator mute or system broadcast procedure

## Rules

- Every runbook should follow the standard template under `docs/templates/runbook.md`.
- Incident learnings should link back to runbooks and ADRs where relevant.
