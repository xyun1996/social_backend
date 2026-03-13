# 106 Party Social Queue Prototype

## Status

`done`

## Goal

Extend the `party` prototype with an explicit social queue lifecycle so the project covers queue entry, exit, and current queue reads before any external matchmaker exists.

## Scope

- add party-owned queue state under the service layer
- require all current members to be online and ready before queue join
- expose HTTP endpoints for queue join, leave, and current queue state
- block member-changing party operations while a queue state is active
- extend the party MySQL repository with queue persistence and tests
- align HTTP and proto contracts with the new queue surface

## Non-Goals

- implementing combat matchmaking
- queue timeout, backfill, or retry policies
- multi-queue search, ranking rules, or ticket merging

## Acceptance

- party leader can join a named queue once all current members are online and ready
- queued parties can read and leave their active queue state
- party membership mutation and ready toggles are rejected while queued
- `go test ./services/party/...` passes

## Completion Notes

- `party` now owns an explicit queue state with leader-driven join and leave semantics
- queue join enforces online and ready validation across the whole party
- queue state is available through both in-memory and MySQL-backed party stores
- handler and service tests cover queue lifecycle and queued-party guardrails
