# 04 Chat Offline

- Status: `planned`
- Version: `v1`

## Goal

Define and implement private, guild, party, world, system, and custom channel messaging with offline retrieval semantics.

## Inputs

- Transport strategy ADR
- Message delivery ADR
- Identity and social primitives

## Outputs

- Chat service module
- Channel model
- Seq/ack/offline retrieval rules
- Hot/cold message persistence approach

## Acceptance Criteria

- At-least-once delivery model is documented and testable.
- Offline window policy is explicit.
- Channel permission checks are defined for all built-in channel types.

## Progress Notes

- Chat already supports conversation creation, seq ordering, ack cursors, replay, delivery planning, and offline delivery recording.
- Resource-backed built-in kinds now use a stable `kind + resource_id` channel model with a readable descriptor surface.

## Risks

- Message ordering, ack behavior, and replay windows are easy sources of drift.
