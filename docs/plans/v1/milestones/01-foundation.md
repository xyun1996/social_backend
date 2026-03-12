# 01 Foundation

- Status: `in-progress`
- Version: `v1`

## Goal

Establish repository structure, documentation governance, templates, and the first architectural source-of-truth documents.

## Inputs

- Project governance plan
- Initial architecture decisions

## Outputs

- Repo scaffold
- Plan and memory structure
- ADR baseline
- Architecture entrypoints
- Initial Go module and reusable service bootstrap packages

## Acceptance Criteria

- New collaborators can find active scope, constraints, and ADRs quickly.
- Empty directories required for future implementation are preserved in version control.
- Templates exist for the core doc types.
- The repository can run basic service processes from a shared bootstrap path.
- Starter services have standard local run entrypoints and example config files.

## Risks

- Structure may drift unless contributors follow source-of-truth rules.
