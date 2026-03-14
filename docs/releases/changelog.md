# Changelog

## 2026-03-14

- Published `v2.0` as the active implementation line under `guild progression + guild chat integration`.
- Extended guild durable storage with contributions, activity instances, idempotent activity records, and reward bookkeeping.
- Added guild progression, contribution, reward, and instance HTTP reads.
- Wired guild governance and progression events into guild chat through system messages.
- Expanded ops guild snapshot reads with progression, contribution, activity instance, and reward state.
- Added minimal worker handlers for guild activity period initialization and expiry transitions.
- Added local durable integration coverage for guild progression + guild chat.

## Earlier Entries

- See [project-archive-v1.md](project-archive-v1.md) and [release-notes/v1.0.md](release-notes/v1.0.md) for the completed `v1` baseline.
