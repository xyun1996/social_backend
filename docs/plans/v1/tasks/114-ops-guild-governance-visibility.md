# 114 Ops Guild Governance Visibility

- Title: Expand ops guild snapshots with announcement and governance log visibility
- Status: `done`
- Version: `v1`
- Milestone: `05 Guild`

## Background

The guild service now owns announcements and governance logs, but the operator-facing `ops` read surface still only exposes guild members. That makes guild governance state harder to inspect during debugging and local durable verification.

## Goal

Extend the `ops` guild snapshot so operators can read guild identity, announcement state, and recent governance logs from the same aggregated boundary.

## Scope

- expand the ops guild snapshot shape with guild aggregate fields and governance logs
- have the ops guild HTTP client read guild aggregate, members, and logs from the guild service
- align ops proto and HTTP contracts with the richer guild snapshot
- add client, service, and handler tests

## Non-Goals

- log filtering or pagination
- guild activity, contribution, or progression reads
- write-side governance operations in ops

## Dependencies

- [109 Guild Announcement Prototype](109-guild-announcement-prototype.md)
- [110 Guild Governance Log Prototype](110-guild-governance-log-prototype.md)
- [Ops HTTP contract](../../../api/http/ops.md)

## Acceptance Criteria

- `GET /v1/ops/guilds/{guildID}` includes guild identity, announcement, member count, and governance log count
- ops proto conversion includes the expanded guild snapshot fields
- `go test ./...` passes

## Related Docs / ADRs

- [Current plan](../../current.md)
- [05 Guild milestone](../milestones/05-guild.md)

## Completion Notes

- ops guild snapshots now aggregate guild basics, announcement state, members, and governance logs
- the guild ops client now fans into `/v1/guilds/{guildID}`, `/members`, and `/logs`
- ops proto alignment now covers the richer guild governance read surface
