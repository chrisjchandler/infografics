# Project Crayons — Codex Rescue Pack

This pack is intended for an **existing, unsatisfactory implementation**.

## Recommended Order

1. `00-AUDIT-AND-RESCUE.md`
2. `01-DATA-SCENE-BINDING.md`
3. `02-SVG-SCENE-QUALITY.md`
4. `03-EXPRESSION-BUILDER.md`
5. `04-DB-REPLICATION-DEMO.md`
6. `05-INTEGRATION-REVIEW.md`

`AGENTS.md` stays at the repository root and applies to all Codex work.

## Workflow

Run the audit first. Do not start parallel agents until the audit has identified the actual source model, canonical series/event model, expression model, scene-state model, renderer path, and any fake/hard-coded animation behavior.

Next fix deterministic source-to-scene bindings. After that, visual-quality and expression-builder work may proceed in parallel if ownership is cleanly separable. Use the database-replication demo as the acceptance fixture, then run the final integration review.

Codex is explicitly allowed to remove or replace Hermes-era implementation choices when they violate `AGENTS.md`.

The existing repository is evidence, not scripture.
