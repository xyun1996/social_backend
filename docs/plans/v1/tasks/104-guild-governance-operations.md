# 104 Guild Governance Operations

## Goal

Extend the `guild` prototype with core owner-driven governance operations so the service supports more than create, invite, and join.

## Scope

- add owner-driven member kick
- add guild ownership transfer
- cover the new behavior in service and HTTP tests

## Non-Goals

- multi-role permission systems
- guild announcements
- guild activity systems

## Acceptance

- owners can transfer ownership to a current member
- owners can kick non-owner members
- guild reads reflect the updated owner and member state
