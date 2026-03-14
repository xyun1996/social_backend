# 03 Social Graph

- Status: `done`
- Version: `v1`

## Goal

Define and implement friend, blocklist, and relationship semantics with clear state transitions and permission checks.

## Inputs

- Identity/session milestone
- Constraints and glossary

## Outputs

- Social service module
- Friend request state model
- Block precedence rules
- Relationship query contracts
- Local in-memory friend and block prototype

## Acceptance Criteria

- Double-confirm friend model is documented.
- Point-to-point block behavior is enforced consistently.
- Relationship APIs support downstream guild/chat/party checks.
- Core relationship lifecycle is runnable without external dependencies.

## Risks

- Relationship semantics can become inconsistent if copied into multiple services.

## Completion Notes

- Friend, block, and pending social state are implemented and surfaced through `ops` player overview for `v1`.
