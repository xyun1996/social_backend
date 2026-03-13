# HTTP Contracts

This directory is the source of truth for the current control-plane HTTP contracts used by runnable prototypes.

## Rules

- Update these docs whenever a wire-visible endpoint shape changes.
- Handler-local request and response structs should follow these docs, not the reverse.
- Keep the contract focused on path, method, key fields, and semantic rules.

## Current Surfaces

- [identity](identity.md): login, refresh, introspection
- [gateway](gateway.md): authenticated session query
- [social](social.md): friend request, accept, list, block
- [invite](invite.md): create, fetch, respond, cancel, list
- [chat](chat.md): conversation create, list, send, ack, replay, delivery planning
- [party](party.md): create, invite, join, ready, leave, kick, transfer leader, queue join/leave, queue handoff
- [guild](guild.md): create, invite, join, kick, transfer owner
- [presence](presence.md): connect, heartbeat, disconnect, lookup
- [ops](ops.md): operator-facing read snapshots
- [worker](worker.md): async job queue lifecycle
