# Runbooks

This directory now carries the minimum on-call baseline for the single-region production target.

## Active Runbooks

- [Gateway disconnect storm triage](gateway-disconnect-storm.md)
- [Redis degradation handling](redis-degradation.md)
- [MySQL write backlog triage](mysql-write-backlog.md)
- [Chat delivery failure triage](chat-delivery-failure.md)
- [Worker retry backlog recovery](worker-retry-backlog.md)
- [Local durable bootstrap and status triage](local-durable-troubleshooting.md)
- [Proto generation workflow](proto-generation.md)
- [Contract inventory workflow](contract-inventory.md)
- [Developer check workflow](dev-checks.md)

## Rules

- Every runbook should follow the standard template under `docs/templates/runbook.md`.
- Incident learnings should link back to runbooks and ADRs where relevant.
