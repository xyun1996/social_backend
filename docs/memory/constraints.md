# Constraints

These are durable project constraints unless explicitly superseded by a newer ADR or current plan update.

## Active Constraints

- Default deployment model is realm-isolated, with future global-graph support preserved in identifiers and routing.
- TCP with Protobuf is the primary real-time transport.
- WebSocket remains a supported compatibility transport.
- HTTP is the login and control surface.
- Player identity, not account identity, is the primary real-time session unit.
- Message delivery targets at-least-once semantics with deduplication.
- MySQL plus Redis is the default storage baseline.
- The repository uses governance-first docs so structural decisions are not lost between sessions.
