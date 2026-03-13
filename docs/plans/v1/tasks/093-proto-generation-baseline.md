# 093 Proto Generation Baseline

## Goal

Move `api/proto` from a set of static contract files to a repository-local generation baseline so future gRPC work is anchored to an executable toolchain shape.

## Scope

- add repository `buf` configuration
- add repository generation output convention for Go bindings
- wire `make proto` to a real generation script
- document the local proto workflow and failure modes
- keep generation concerns outside hand-written service packages

## Acceptance

- the repository includes `buf.yaml` and `buf.gen.yaml`
- `make proto` points to a concrete generation entrypoint
- generated bindings have a reserved output directory
- documentation explains how to lint and generate local proto bindings
