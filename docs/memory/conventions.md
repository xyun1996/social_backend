# Conventions

## Repository

- Use lowercase kebab-case for documentation file names unless a numbered prefix is intentionally used.
- Keep governance docs under `docs/`; do not scatter planning files at the repo root.

## Status Values

- Plans and tasks use: `planned`, `in-progress`, `done`, `blocked`
- ADRs use: `proposed`, `accepted`, `superseded`, `deprecated`

## Cross-Reference Rules

- Tasks must link to their milestone and any relevant ADRs.
- ADRs should link to affected architecture docs and current plan when relevant.
- Session notes should link to any decisions promoted into ADRs.

## Naming

- Service names should match their directory names.
- Public contract directories under `api/` should remain transport-specific.

## Documentation Hygiene

- Update the highest-priority source of truth first.
- Prefer concise summaries at the top of long documents.
