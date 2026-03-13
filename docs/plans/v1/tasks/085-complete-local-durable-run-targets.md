# 085 Complete Local Durable Run Targets

## Goal

Bring local durable run shortcuts in `Makefile` and docs up to the current durable service surface.

## Scope

- add local durable run targets for `party`, `guild`, `gateway`, and durable `ops`
- update README and local durable runflow docs to list the new shortcuts

## Non-Goals

- changing default service addresses
- adding deployment manifests

## Acceptance

- local docs and `Makefile` cover the full current durable service surface
- `ops` has a documented shortcut for combined MySQL and Redis status reads
