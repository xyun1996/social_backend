# 051 Chat MySQL Repo Foundation

## Goal

Add a service-local MySQL repository foundation for `chat` so durable conversation, message, and cursor work has an explicit schema owner in code.

## Scope

- add `services/chat/internal/repo/mysql`
- define chat-owned schema statements for conversations, members, messages, and cursors
- align persistence documentation

## Acceptance

- chat has a MySQL repository foundation with explicit schema ownership
- persistence docs point to the new foundation path
