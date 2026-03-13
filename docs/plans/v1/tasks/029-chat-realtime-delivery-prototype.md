# Task 029 - Chat Realtime Delivery Prototype

## Context

`chat` now computes delivery planning and creates offline job intent, while `gateway` now owns a realtime session state machine. But there is still no executable prototype for the online push leg where gateway consumes chat delivery planning and routes events into active session targets.

## Goal

Add a gateway-side realtime delivery prototype that consumes chat delivery planning, writes online events into gateway session inboxes, and leaves offline targets on the replay path.

## Scope

- Add a chat delivery planning client to gateway
- Add gateway-owned event inbox storage per realtime session
- Expose an HTTP prototype endpoint for chat delivery dispatch
- Expose an HTTP prototype endpoint to inspect session event inboxes

## Non-Goals

- Actual TCP or WebSocket push
- Delivery acknowledgement from clients
- Final durable push retry logic

## Acceptance Criteria

- Online targets are enqueued into gateway session event inboxes
- Offline targets remain on the replay path and are reported as deferred
- Gateway tests cover delivery dispatch and event inspection

## Status

`done`
