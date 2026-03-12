# ADR-004 Message Delivery Semantics

- Status: `accepted`
- Date: `2026-03-12`

## Context

The chat system needs a practical delivery model for multiplayer social features without turning v1 into a heavy exactly-once messaging platform.

## Decision

Use at-least-once delivery semantics with deduplication based on stable identifiers and conversation sequencing.

## Alternatives Considered

- Strong commit-before-push delivery
- Best-effort unordered delivery without stable dedupe

## Consequences

- Client and gateway need dedupe support.
- Seq, ack, and replay rules must remain centralized in protocol and chat docs.
