# 01 Social Depth

- Status: `planned`
- Version: `v2`

## Goal

Extend the social graph beyond the `v1` friend/block baseline with richer read models and relationship metadata that can support downstream product features.

## Inputs

- `v1.0` social baseline
- player overview aggregation patterns from ops
- backlog themes for recommendation and relationship reads

## Outputs

- richer relationship queries
- optional friend remarks / metadata
- better pending-state reads and aggregation

## Acceptance Criteria

- relationship reads can answer more than simple friend/block presence
- the new social surface remains compatible with the shipped `v1` flows

## Risks / Blockers

- social deepening can sprawl quickly if recommendation logic and metadata are mixed too early
