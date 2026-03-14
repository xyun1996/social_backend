# MySQL Write Backlog Triage

## Scenario

MySQL writes fail or back up for identity, social, invite, chat, party, or guild durable paths.

## Triggers

- MySQL ping or migration verification failures
- Durable flow tests or production writes start timing out
- Service logs show repeated write failures

## Checks

1. Confirm MySQL connectivity and account permissions.
2. Run the migration verification tool against the target database.
3. Inspect recent writes for guild progression, chat persistence, invite state changes, and social metadata.

## Recovery

1. Restore MySQL connectivity before restarting writers.
2. If schema drift is suspected, run the migration verification step and compare expected rows.
3. Restart services with durable stores only after MySQL health is confirmed.

## Exit Criteria

- MySQL responds to pings and migration verification
- Durable write paths succeed again
- No service remains stuck in repeated write-failure loops
