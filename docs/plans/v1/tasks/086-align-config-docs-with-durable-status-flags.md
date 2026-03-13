# 086 Align Config Docs With Durable Status Flags

## Goal

Bring config and environment documentation back in sync with the current durable status toggles and remove outdated flags.

## Scope

- document `OPS_MYSQL_STATUS` and `OPS_REDIS_STATUS`
- remove outdated references that no longer match runtime behavior

## Non-Goals

- changing config keys
- new runtime features

## Acceptance

- config docs mention current ops durable status flags
- outdated runtime toggle references are removed
