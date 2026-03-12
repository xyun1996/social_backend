# 03 Social Graph

- Status: `planned`
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

## Acceptance Criteria

- Double-confirm friend model is documented.
- Point-to-point block behavior is enforced consistently.
- Relationship APIs support downstream guild/chat/party checks.

## Risks

- Relationship semantics can become inconsistent if copied into multiple services.
