# 085 Complete Local Durable Run Targets

## Status

`done`

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

## Completion Notes

- `Makefile` now exposes durable run shortcuts for `party`, `guild`, `gateway`, and `ops`
- Windows-compatible PowerShell wrappers back the durable targets so `make` works consistently on local Windows development hosts
