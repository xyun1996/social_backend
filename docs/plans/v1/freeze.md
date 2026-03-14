# V1 Freeze

- Status: `active`
- Version: `v1`
- Updated: `2026-03-14`

## Purpose

This document defines the smallest acceptable `v1` release line for the Social Backend. Anything outside this line should be treated as `v2` unless it blocks the stated acceptance checks.

## V1 Must-Have Capabilities

### Identity

- login
- refresh
- introspection
- player-scoped session identity

### Social

- friend request send/accept
- friend list
- block list and basic block enforcement

### Invite

- create
- accept
- reject
- cancel
- expire

### Chat

- create conversation
- send message
- ack
- replay
- conversation summary and unread count
- resource channel model
- guild/party channel permission alignment

### Guild

- create
- invite and join
- owner transfer
- kick
- announcement
- governance logs
- baseline growth (`level`, `experience`)
- first activity template skeleton

### Party

- create
- invite and join
- ready state
- leave / kick / transfer leader
- queue join / leave / state
- match handoff
- assignment callback
- assignment resolution cleanup

### Ops

- player overview
- party snapshot
- guild snapshot
- worker queue view
- durable summary

## Release Gates

- `go test ./...`
- `make check-dev`
- `make test-local-durable`
- current milestones and release notes reflect actual implementation state

## Explicitly Deferred

- advanced guild progression design
- advanced chat moderation and channel governance
- social recommendations and notes
- advanced worker retry policy
- full matchmaker lifecycle and game-server orchestration
- production deployment automation

## Remaining V1-Freeze Work

1. Close the last runtime-alignment gaps that block a credible release line
2. Re-run durable verification after final changes
3. Publish release notes, known gaps, and `v2` carry-over items
