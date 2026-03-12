# Risk Register

## Active Risks

### R-001 Transport Complexity

- Summary: supporting TCP as the primary real-time protocol while maintaining WebSocket compatibility can create drift.
- Mitigation: centralize envelope semantics and compatibility tests.

### R-002 Session Ambiguity

- Summary: account-vs-player identity confusion can cascade across domains.
- Mitigation: keep session ownership player-scoped and document it everywhere.

### R-003 Plan Drift

- Summary: governance docs can diverge if not updated in order.
- Mitigation: follow source-of-truth rules and link every task to milestones and ADRs.

### R-004 Messaging Semantics Drift

- Summary: seq, ack, replay, and dedupe semantics are easy to redefine inconsistently.
- Mitigation: maintain one protocol truth and test around it early.
