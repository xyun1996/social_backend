# 052 Social MySQL Repo Foundation

## Goal

Add a service-local MySQL repository foundation for `social` so durable friend request, friendship, and block ownership is explicit in code.

## Scope

- add `services/social/internal/repo/mysql`
- define social-owned schema statements
- align persistence documentation

## Acceptance

- social has a MySQL repository foundation with explicit schema ownership
- persistence docs point to the new foundation path
