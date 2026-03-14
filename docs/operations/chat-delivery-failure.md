# Chat Delivery Failure Triage

## Scenario

Chat send, replay, or offline delivery processing starts failing or lagging.

## Triggers

- Chat request failures rise
- Worker offline delivery jobs accumulate
- Guild or party system events stop appearing in channel history

## Checks

1. Query `/metrics` on chat and worker for error-heavy endpoints.
2. Check worker job state for retry or dead-letter growth.
3. Verify guild and party membership reads are healthy because chat visibility depends on them.

## Recovery

1. Restore chat MySQL connectivity first if durable message writes fail.
2. Restore worker and internal token paths if offline delivery processing fails.
3. Re-run one send plus replay flow against a known conversation.

## Exit Criteria

- Message send succeeds
- Replay returns recent messages
- Worker retry backlog stops growing
