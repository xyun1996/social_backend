# Config Strategy

## Directory Layout

- `configs/local/`
- `configs/dev/`
- `configs/staging/`
- `configs/prod/`
- `configs/examples/`

## Rules

- Commit templates and examples, not secrets.
- Prefer environment variables for secrets and environment-specific overrides.
- Document every config key that affects runtime behavior.

## Ownership

- Service-owned keys should be documented close to the service and summarized here.
- Shared infrastructure keys belong here first and may be referenced elsewhere.

## Future Work

- Decide on exact config format and loading library when Go modules are introduced.
