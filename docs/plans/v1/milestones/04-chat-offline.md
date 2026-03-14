# 04 Chat Offline

- Status: `done`
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
- Chat now exposes conversation summaries with unread counts, ack cursors, and last-message snapshots.
- Guild and party channel access now aligns to current resource membership instead of only the stored conversation member list.

## Risks

- Message ordering, ack behavior, and replay windows are easy sources of drift.

## Completion Notes

- Chat now meets the `v1` line for conversation creation, delivery, replay, unread summaries, resource channels, and guild/party permission alignment.
