# 105 Invite Cancel Lifecycle

## Goal

Close the invite lifecycle gap by allowing senders to cancel pending invites through the shared invite boundary.

## Scope

- add invite cancel service logic
- expose `POST /v1/invites/{inviteID}/cancel`
- add tests for sender-only cancellation

## Non-Goals

- invite recall side effects in downstream domains
- bulk cancellation
- operator cancellation APIs

## Acceptance

- senders can cancel pending invites
- non-senders cannot cancel invites
- terminal invites remain non-cancelable
