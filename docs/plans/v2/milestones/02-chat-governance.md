# 02 Chat Governance

- Status: `planned`
- Version: `v2`

## Goal

Build on the `v1` chat baseline with stronger channel governance, moderation hooks, and clearer policy boundaries for system, world, and custom channels.

## Inputs

- `v1.0` chat baseline
- resource-backed channel model
- backlog themes for advanced channel governance

## Outputs

- stronger send and visibility policies by channel kind
- moderation-ready extension points
- better governance for world/system/custom channels

## Acceptance Criteria

- advanced chat policy remains contract-first and testable
- built-in channels have clearer governance semantics than `v1`

## Risks / Blockers

- moderation and channel policy can become too broad if runtime rules and product policy are not separated
