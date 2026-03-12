# Dependencies

## Service Dependencies

- `gateway` depends on `identity`, `presence`, and domain service APIs.
- `presence` depends on Redis-style short-lived state storage.
- `social`, `guild`, `invite`, `chat`, and `party` depend on identity-provided player context.
- `worker` depends on domain-owned job definitions and storage access.
- `ops` depends on stable service APIs and audit data.

## Storage Dependencies

- MySQL is the default long-lived relational store.
- Redis is the default short-lived cache, presence, queue, and hot-message store.

## Documentation Dependencies

- Current plan depends on roadmap, ADRs, and architecture docs remaining aligned.
- Tasks depend on milestone and current plan references.

## Future External Dependencies

- Matchmaker boundary for party queue handoff
- Optional event bus in future versions
- Optional moderation or analytics integrations
