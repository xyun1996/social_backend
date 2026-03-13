# 074 Gateway Redis Session Store

## Goal

Move gateway realtime session and inbox state behind explicit stores and add an optional Redis-backed runtime store.

## Scope

- introduce gateway realtime session and event store interfaces
- move in-memory implementations behind those interfaces
- implement Redis persistence for sessions and buffered events
- add `GATEWAY_STORE=redis` startup selection

## Non-Goals

- changing gateway realtime semantics
- distributed delivery coordination
- durable chat replay history beyond current session inbox state

## Acceptance

- gateway realtime logic no longer depends directly on in-memory maps
- Redis-backed runtime state can be enabled by configuration
- existing realtime, ack, and replay tests remain green
