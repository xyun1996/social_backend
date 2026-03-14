# 03 Phase A Core Social Package

- Status: `in-progress`

## Goal

Deliver the first genuinely product-grade capability slice:

- login and session
- friends and blocks
- invites
- private chat
- guild basics
- party basics

## Success Criteria

- The reduced feature set is complete end to end in the new runtime.
- Staging and operator flows are based on the new runtime, not the frozen prototype services.

## Progress

- `social-core` now exposes an explicit module registry for the phase A domains:
  identity, social, invite, private chat, guild basics, and party basics.
- This registry is now the canonical inventory for migration into the consolidated runtime.
- `social-core` now serves product-rebuild auth and social-graph endpoints directly for
  login/refresh/introspect and friends/blocks basics, instead of only documenting them.
- `social-core` also serves invite lifecycle endpoints directly for
  create/get/respond/cancel/list, giving the rebuilt runtime its first cross-domain flow.
