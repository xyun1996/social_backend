# 050 Invite MySQL Repo Foundation

## Goal

Add a service-local MySQL repository foundation for `invite` so durable invite lifecycle work has an explicit schema owner in code.

## Scope

- add `services/invite/internal/repo/mysql`
- define invite-owned schema statements
- align persistence documentation

## Acceptance

- invite has a MySQL repository foundation with explicit schema ownership
- persistence docs point to the new foundation path
