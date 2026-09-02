# ADR 0002: SQLite Control Plane and Revision Semantics

## Status
Accepted

## Context
Infograf needs durable local state for configuration, scenes, and change history. The control plane must work offline, support replayable revisions, and be easy to verify in local development.

## Decision
- Use **SQLite** for the control plane store.
- Persist the following core objects:
  - app configuration
  - scenes
  - immutable revision records
- Seed the database on first launch with:
  - a default Infograf workspace config
  - a demo replication scene
- Expose control-plane endpoints for:
  - config read/update
  - scene list/create/read/update
  - revision listing
  - health/readiness/version

## Consequences
- The backend can now survive process restarts without losing the canonical scene/config state.
- Revision history becomes a first-class API surface rather than an in-memory log.
- Tests can target the real control plane behavior instead of only generated demo JSON.

## Revision behavior
- Every config update increments the stored config revision.
- Every scene create/update increments the scene revision and emits a revision record.
- Revisions are append-only and ordered newest-first by timestamp.

## Notes
The actual source adapter storage for external TSDB data remains separate from the SQLite control plane. SQLite stores canonical platform state, not the source histories themselves.
