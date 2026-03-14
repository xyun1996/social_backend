# 05 Guild

- Status: `done`
- Version: `v1`

## Goal

Define and implement guild organization, governance, growth, and the first activity templates.

## Inputs

- Social relationship rules
- Invite semantics
- Presence expectations

## Outputs

- Guild service module
- Member and role model
- Guild log and progression model
- Activity framework with initial templates

## Progress Notes

- Guild create, invite, join, kick, transfer-owner, and presence-aware member reads are implemented.
- Guilds now support owner-managed announcement updates on the core guild aggregate.
- Guild governance logs are now readable and progression now tracks level and experience.
- Guild activity templates now include `sign_in`, `donate`, and `task`, with durable activity records and guild XP growth.
- Ops guild snapshots now aggregate announcement state and governance logs for operator visibility.

## Acceptance Criteria

- Role and permission boundaries are explicit.
- Join/apply/approve flows are consistent with relationship checks.
- Activity templates have documented lifecycle and reward records.

## Risks

- Guild activity scope can expand quickly unless held to the roadmap.

## Completion Notes

- Guild now meets the `v1` line for governance, announcements, logs, growth, and the first durable activity template skeleton.
