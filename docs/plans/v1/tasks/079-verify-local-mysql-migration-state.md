# 079 Verify Local MySQL Migration State

## Goal

Make local MySQL bootstrap observable by verifying required `schema_migrations` rows after service-owned bootstrap completes.

## Scope

- add a local verification script for `schema_migrations`
- wire verification into local bootstrap and Makefile targets
- add real local MySQL integration coverage for repeated migration registration

## Non-Goals

- production migration promotion
- rollback tooling

## Acceptance

- local bootstrap reports missing owned migrations instead of silently succeeding
- repeated service-owned bootstrap keeps `schema_migrations` stable
- local durable integration coverage includes migration verification
