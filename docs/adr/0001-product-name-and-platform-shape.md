# ADR 0001: Product Name and Platform Shape

## Status
Accepted

## Context
The source material alternated between several names while describing the same product concept. The implementation also consistently favored a single deployable control plane rather than a microservice graph.

## Decision
- The product name used in the repository, docs, and UI is **Infograf**.
- The existing `Crayons` folder name is retained as the workspace/container name only.
- The runtime shape is a **Go modular monolith** with an embedded frontend and a small number of bounded internal packages.
- The first implementation target is a WSL-friendly developer setup that can run locally as a single process.

## Consequences
- All new user-facing docs should say Infograf.
- Source connectors, scene packs, expressions, actions, revisions, and audit remain separated as internal modules, not services.
- Packaging can later expand to Docker and other deployment modes without changing the core layout.

## Notes
This decision does not settle plugin isolation details permanently; it only fixes the product name and the top-level runtime shape.
