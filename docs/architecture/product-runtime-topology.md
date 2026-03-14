# Product Runtime Topology

## Purpose

Define the reduced runtime target for the product rebuild.

## Runtime Units

### api-gateway

- Owns external HTTP/WebSocket ingress
- Owns auth boundary, request attribution, client rate limits, and external session policy
- Does not own durable domain truth

### social-core

- Owns the first product-grade implementation of:
  - identity/session
  - social graph
  - invites
  - private chat
  - guild basics
  - party basics
- Keeps module boundaries internally, but is a single runtime boundary early on

### ops-worker

- Owns support reads, repair/admin workflows, and async task execution
- Serves product support before broader GM platformization

## Why This Topology

- It reduces operational complexity while implementation depth is still low.
- It keeps business boundaries without paying full microservice tax too early.
- It gives the project a realistic path to a real staging and release cycle.

## Relationship To Existing Services

- Existing per-domain services remain useful as prototypes and references.
- They are not the target production architecture.
- Product implementation should move into the new runtime line first.
