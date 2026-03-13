# Task 036 - Worker Background Runner

## Context

`worker` can now execute registered job handlers, but execution still depends on explicit HTTP calls such as `run-once` or `run-until-empty`. There is no background runner to continuously drain queued work.

## Goal

Add an optional background runner that periodically executes queued jobs until the worker process stops.

## Scope

- Add background runner methods to `worker`
- Add optional process startup wiring in `worker` main
- Update docs, examples, and tests

## Non-Goals

- Backoff or jitter policies
- Multi-worker coordination
- Durable leases

## Acceptance Criteria

- Worker can start a background runner with a configurable interval
- The runner processes jobs until context cancellation
- Tests cover background execution

## Status

`done`
