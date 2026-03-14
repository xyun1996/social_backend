# Product Rebuild

- Status: `active`
- Last updated: `2026-03-14`

## Purpose

Treat the current repository as a high-fidelity prototype asset base and rebuild the runtime into a smaller, product-oriented deployment shape.

## What Stays

- `docs/` remains the primary product memory and decision system.
- `api/http` and `api/proto` remain the contract reference base.
- Existing domain models and integration scenarios remain useful as reference material.
- Durable local verification patterns remain valid and should be preserved.

## What Freezes

- Existing `services/{gateway,identity,social,invite,chat,party,guild,ops,worker,presence}` stay as prototype implementations.
- They remain runnable for reference and regression comparison, but are no longer the target production architecture.

## What Changes

- New active runtime target becomes:
  - `api-gateway`
  - `social-core`
  - `ops-worker`
- Product development should prioritize module depth and operational quality over service count.
- Empty or placeholder infra directories should no longer be treated as implicit future capability.

## Rebuild Phases

1. Runtime consolidation
2. Product foundation rebuild
3. Phase A core social package
4. Staging and release readiness
5. Selective expansion after core package is operational

## Reuse Rules

- Reuse freely:
  - plans
  - ADRs
  - architecture docs
  - contract docs
  - durable flow scenarios
- Reuse carefully:
  - domain structs
  - validation rules
  - persistence ownership boundaries
- Do not assume direct reuse:
  - current HTTP handlers
  - current repo implementations
  - current service wiring

## Immediate Deliverables

- Switch the active plan from `production` hardening to `product-rebuild`.
- Add the new runtime entrypoints under `services/api-gateway`, `services/social-core`, and `services/ops-worker`.
- Add a dedicated product roadmap and milestones.
- Update architecture docs so the runtime target is explicit.
