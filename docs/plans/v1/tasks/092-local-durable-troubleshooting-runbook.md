# 092 Local Durable Troubleshooting Runbook

## Goal

Document the actual operator workflow for recovering local durable startup and status failures now that the repository has multiple durable bootstrap and verification entrypoints.

## Scope

- add a runbook for local durable bootstrap and status triage
- cover MySQL bootstrap, migration verification, `ops` durable readers, and status gate failures
- document the expected local service set and how to intentionally relax checks for partial topologies
- link the runbook from the operations index, local durable runflow, and repository entrypoints

## Acceptance

- there is a dedicated runbook for local durable startup and status failures
- the runbook includes concrete commands already available in the repository
- the runbook explains the most likely local failure signatures and what they mean
