# 116 Guild Growth Activity Template Skeleton

- Title: Add guild progression and the first durable activity templates
- Status: `done`
- Version: `v1`
- Milestone: `05 Guild`

## Background

Guild governance had reached create, invite, join, and ownership operations, but the v1 milestone still lacked the promised growth model and first activity template skeleton.

## Goal

Introduce a lightweight guild progression model with durable activity records and a fixed first batch of activity templates that can grow guild experience.

## Scope

- add `level` and `experience` to the guild aggregate
- define fixed activity templates for `sign_in`, `donate`, and `task`
- add activity submission and list endpoints
- persist activity records in memory and MySQL-backed flows
- append governance log entries when activity is submitted
- add service, handler, and repository tests

## Non-Goals

- seasonal guild activities
- reward inventory payout
- complex contribution caps or reset schedules

## Dependencies

- [110 Guild Governance Log Prototype](110-guild-governance-log-prototype.md)
- [05 Guild milestone](../milestones/05-guild.md)
- [Guild HTTP contract](../../../api/http/guild.md)

## Acceptance Criteria

- guild reads include level and experience
- supported activity templates are discoverable through HTTP
- activity submission persists a durable record and increases guild progression
- `go test ./...` passes

## Related Docs / ADRs

- [Current plan](../../current.md)
- [Bounded contexts](../../../architecture/bounded-contexts.md)

## Completion Notes

- guild progression now tracks experience and level directly on the guild aggregate
- durable activity records are stored for the first three template types
- activity submission now contributes to guild growth and writes a governance log entry
