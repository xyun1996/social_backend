# Task 030 - Worker Job Executor Prototype

## Context

`worker` can enqueue, claim, complete, fail, and list jobs, but it still has no execution loop or handler registry. That means all async intent created by `invite` and `chat` stops at queue visibility and never becomes runnable work.

## Goal

Add a generic worker executor prototype that can claim queued jobs, dispatch them to registered handlers, and complete or fail them automatically.

## Scope

- Add a handler registry to `worker`
- Add `run-once` and `run-until-empty` execution methods
- Expose HTTP endpoints for execution control
- Update docs and tests

## Non-Goals

- Background polling loops
- Durable scheduling or backoff
- Domain-specific job handlers

## Acceptance Criteria

- Worker can execute a queued job through a registered handler
- Successful handlers complete jobs
- Handler errors fail jobs with a visible last error

## Status

`done`
