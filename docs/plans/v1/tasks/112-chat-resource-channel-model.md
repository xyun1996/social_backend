# 112 Chat Resource Channel Model

- Title: Add a resource-backed channel model for built-in chat kinds
- Status: `done`
- Version: `v1`
- Milestone: `04 Chat Offline`

## Background

The chat prototype already supports multiple built-in `kind` values, but those kinds still behave mostly like generic conversations. `guild`, `party`, `world`, `system`, and `custom` need explicit channel semantics before downstream services can rely on them as stable chat surfaces.

## Goal

Turn built-in resource-backed chat kinds into a real channel model with explicit validation, stable reuse by `kind + resource_id`, and a readable descriptor surface that explains channel policy.

## Scope

- tighten conversation creation validation by kind
- require `resource_id` for resource-backed channel kinds
- reuse and reconcile existing channels for repeated `kind + resource_id` creation
- add a channel descriptor read endpoint and proto contract
- add service and handler tests for the new channel rules

## Non-Goals

- guild/party membership sourcing from external services
- channel moderation and mute rules
- unread summary aggregation
- world broadcast fanout optimizations

## Dependencies

- [008 Chat HTTP Prototype](008-chat-http-prototype.md)
- [04 Chat Offline milestone](../milestones/04-chat-offline.md)
- [Chat HTTP contract](../../../api/http/chat.md)

## Acceptance Criteria

- `guild`, `party`, `world`, `system`, and `custom` conversations reject missing `resource_id`
- repeated creation for the same resource-backed channel reuses the existing conversation
- `GET /v1/conversations/{conversationID}/channel` returns the resolved channel policy
- `go test ./...` passes

## Related Docs / ADRs

- [Current plan](../../current.md)
- [Bounded contexts](../../../architecture/bounded-contexts.md)

## Completion Notes

- built-in resource channels now behave as stable `kind + resource_id` bindings instead of free-form conversation rows
- repeated resource-channel creation reconciles membership scope without creating duplicate channels
- channel descriptors now expose scope, membership mode, and send policy for downstream runtime consumers
