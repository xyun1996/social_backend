# Bounded Contexts

## Identity

- Owns `account`, `credential binding`, `token`, and `player selection`.
- Exposes authenticated player context to downstream systems.

## Presence

- Owns online status, connection metadata, and short-lived player location context.

## Social Graph

- Owns `friend relationship`, `friend request`, and `block relationship`.
- Supplies relationship checks to chat, guild, and party domains.

## Guild

- Owns `guild`, `guild member`, `guild role`, `guild progression`, and `guild activity`.

## Invite

- Owns cross-domain invitation lifecycle, TTL, and acceptance state.

## Chat

- Owns `conversation`, `channel`, `message`, `message seq`, `read cursor`, and `offline replay`.

## Party

- Owns `party`, `party member`, `ready state`, `party invite`, and `social queue state`.

## Operations

- Owns operator-facing queries, moderation actions, and audit-oriented management flows.

## Cross-Context Rules

- Shared identifiers should include tenant and realm routing context where applicable.
- Contexts communicate through explicit contracts, not direct database coupling.
