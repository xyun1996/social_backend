# Environments

## Local

- Purpose: fast developer iteration
- Expected dependencies: local MySQL, Redis, optional tracing backend
- Tolerance: mock integrations acceptable

## Dev

- Purpose: integration validation across services
- Expected dependencies: shared MySQL/Redis, basic observability

## Staging

- Purpose: release validation and protocol compatibility testing
- Expected dependencies: production-like topology where practical

## Prod

- Purpose: live game traffic
- Requirements: audited changes, stronger alerting, stable runbooks, and release notes

## Rules

- Document environment differences before relying on environment-specific behavior.
- Avoid introducing runtime assumptions that only exist in local development.
