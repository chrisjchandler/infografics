# Infograf Rescue Audit

Date: 2026-09-02

## Scope

This audit covers the repository state after the Task 00 foundational repair pass. The initial local environment did not expose usable Linux `go` or `npm` binaries, but repo-local wrappers and a bootstrap script now provide working toolchains for local verification. Browser-driven screenshot capture remains unavailable in this environment.

## Current runtime architecture

- Runtime shape: Go modular monolith with one HTTP server entrypoint in `cmd/infograf/main.go`.
- Backend surface: REST endpoints in `internal/app/server.go`.
- Persistence: SQLite through `modernc.org/sqlite` in `internal/store/sqlite/store.go`.
- Static UI delivery: embedded HTML from `internal/app/static/index.html`.
- External source path present before rescue: Graphite demo fetch in `internal/app/graphite.go`.
- Frontend build scaffold present but not runtime-authoritative: `frontend/src/main.ts` and `frontend/dist/main.js`.
- Container/process topology: single process, single binary target by design. No microservice split found.

## Expected architecture

The governing documents require this invariant:

```text
Source
  ↓
Source-aware adapter
  ↓
Canonical series / canonical events
  ↓
Expression / rule evaluation
  ↓
Scene binding
  ↓
Deterministic scene state
  ↓
Renderer interpolation
  ↓
Infographic
```

The mathematics must remain authoritative even when the metaphor changes.

## Original implementation findings

### Backend

- `internal/app/server.go`
  - Correct: single-process server shape; config/scenes/revisions REST surface.
  - Incomplete: no canonical source-to-scene pipeline.
  - Placeholder: `/api/v1/demo/replication` returned a seeded presentation object instead of computed scene state.
- `internal/app/graphite.go`
  - Correct: source-specific Graphite fetch exists and summarizes a live/fallback demo.
  - Architecturally wrong: output was presentation-oriented cards rather than canonical scene inputs.
- `internal/store/sqlite/store.go`
  - Correct: SQLite persistence, config seeding, scene persistence, revision append behavior.
  - Incomplete: seeded demo scene payload stored a mock display document rather than a reusable scene definition.

### Frontend

- `internal/app/static/index.html`
  - Architecturally wrong before rescue: scene rendering depended on frontend-generated metaphor logic, local scene nudges, hard-coded SVG positioning, and timer-driven decorative animation.
  - Duplicated logic: the frontend independently computed value-to-position behavior instead of consuming backend scene state.
  - Placeholder: UI resembled a generic admin shell with a metaphor card layer rather than an infographic authoring surface.
- `frontend/src/main.ts`
  - Incomplete and non-authoritative: separate frontend shell existed but was not the actual embedded runtime path.

### Visual assets

- `internal/app/static/assets/generated/sisyphus/*`
  - Potentially reusable later for scene-pack work.
  - Not integrated into the current deterministic replication renderer.

## Datapoint trace

The repaired Task 00 trace for the replication-lag datapoint at `2026-08-30T10:22:00Z` is:

1. Source connection
   - `internal/demo/replication.go`
   - Fixture rows are loaded from a deterministic in-repo dataset.
   - Classification: correct
2. Source-aware adapter
   - `replicationSeries(...)` in `internal/demo/replication.go`
   - Fixture-specific replication fields are converted into canonical metric series.
   - Classification: correct
3. Sample representation
   - `model.CanonicalSeries` and `model.CanonicalPoint` in `internal/model/scene.go`
   - Timestamped numeric samples now exist as source-agnostic structures.
   - Classification: correct
4. Semantic interpretation
   - `model.SceneBinding.Interpretation` in `internal/model/scene.go`
   - Each bound metric carries type, preferred direction, unit, and range.
   - Classification: correct
5. Expression / rule evaluation
   - `internal/scenes/engine.go`
   - Trend, rate, acceleration, volatility, and active-event windows are evaluated.
   - Threshold expression parsing remains incomplete.
   - Classification: incomplete
6. Scene binding
   - `EvaluateScene(...)` in `internal/scenes/engine.go`
   - `replication_lag_seconds` is normalized and mapped into `replica-db.x` and `replica-db.y`.
   - Classification: correct
7. Scene-state construction
   - `model.SceneState` and `model.SceneObjectState`
   - Reconstructed deterministically for the requested timestamp.
   - Classification: correct
8. X/Y/property mapping
   - `Normalize(...)` and `MapNormalized(...)` in `internal/scenes/engine.go`
   - One backend implementation path now controls normalized property values.
   - Classification: correct
9. Frontend state transport
   - `/api/v1/demo/replication` in `internal/app/server.go`
   - Returns scene definition, fixture, trace, and computed scene state.
   - Classification: correct
10. SVG/DOM update
   - `internal/app/static/index.html`
   - The renderer now reads backend scene state and updates object transforms and labels directly.
   - Classification: correct
11. Animation interpolation
   - `internal/app/static/index.html`
   - Decorative loops were removed from the main scene path; displayed movement is endpoint-driven by scene state.
   - Classification: correct for Task 00

## Event trace

- Canonical event model now exists as `model.CanonicalEvent`.
- Replication incident fixture:
  - ID: `INC-REPL-001`
  - Start: `2026-08-30T10:22:00Z`
  - End: `2026-08-30T10:24:00Z`
- Active window evaluation:
  - `activeEventsAt(...)` in `internal/scenes/engine.go`
- Visual event state:
  - `incident-callout` object opacity/state in `EvaluateScene(...)`
- Timeline replay:
  - deterministic by requested `at` timestamp on the backend
- Classification:
  - persistence: incomplete
  - historical visual synchronization: correct for the demo fixture path

## Components safe to retain

- `cmd/infograf/main.go`
- core REST server structure in `internal/app/server.go`
- SQLite store/revision foundation in `internal/store/sqlite/store.go`
- Graphite source fetch code as future adapter input material in `internal/app/graphite.go`
- existing docs, ADRs, and asset provenance records

## Components replaced or materially changed

- Replaced seeded replication presentation payload with a deterministic fixture-driven scene path.
- Replaced the embedded HTML shell with a scene-first replication studio that consumes backend scene state.
- Added canonical scene model types in `internal/model/scene.go`.
- Added deterministic scene evaluator in `internal/scenes/engine.go`.
- Added fixture adapter and replication demo constructor in `internal/demo/replication.go`.
- Updated seeded scene payload storage to persist the scene definition rather than a mock status object.

## Critical defects found

- No canonical scene state model existed.
- The frontend owned metric-to-visual behavior that belonged in the backend binding path.
- Decorative/timer-driven animation existed in a metric-bound surface.
- The replication demo endpoint returned mock presentation data instead of deterministic scene state.
- The embedded runtime UI and the `frontend/` TypeScript shell diverged.
- No boundary tests covered normalized mapping or deterministic replay.

## Foundational corrections completed in Task 00

- Added canonical series, events, bindings, scene definition, and scene state models.
- Added backend normalization and mapping helpers with explicit `min == max` failure behavior.
- Added deterministic timestamp reconstruction for metric samples.
- Added trend/rate/acceleration/volatility derivation from timestamped data.
- Added active event window evaluation.
- Added a fixture-backed replication demo response that exposes the real scene path.
- Added test coverage for value normalization, inversion, clamping, invalid bounds, determinism, and demo endpoint shape.

## Remaining defects after Task 00

- No real expression parser or beginner/advanced rule authoring workflow yet.
- Graphite path still terminates in `GraphiteOverview` cards instead of canonical series.
- `frontend/src/main.ts` remains an obsolete parallel frontend surface and should be reconciled or removed in later rescue work.
- No screenshot capture or browser-verified visual acceptance was possible in this environment.
- Scene-pack file conventions, YAML round-trip, and export surfaces are still incomplete.
- Security features beyond the current control-plane baseline have not yet been audited in depth.

## Correction sequence after Task 00

1. Complete Task 01 by generalizing binding validation and extending tests to every target property and inverse mapping path.
2. Convert the Graphite demo path into canonical series so source-agnostic behavior is proven across more than one adapter.
3. Complete Task 02 by replacing the current replication SVG with a fuller scene pack and capture screenshots.
4. Complete Task 03 by introducing the shared editor/YAML model and action safety model.
5. Complete Task 04 using the replication fixture as the release acceptance demo.
6. Complete Task 05 with integration review, screenshots, and deployment/security regression validation.

## Acceptance status for Task 00

- One source sample traceable to one scene property: PASS
- No random animation required for the bound property: PASS
- Same timestamp/data/config yields same scene state: PASS by implementation and tests authored
- Scene state is source-agnostic: PASS for the model boundary
- X/Y mapping has one defined implementation path: PASS
- Core model-boundary tests exist: PASS
- `docs/rescue-audit.md` reflects the repaired state: PASS

## Verification constraints

- On September 2, 2026, local Linux Go and Node toolchains were bootstrapped successfully and verification was rerun in this workspace.
- `go test ./...`: PASS after redirecting Go caches to writable `/tmp` paths and allowing module download.
- `npm run build`: PASS with the bootstrapped local Node toolchain.
- `go run ./cmd/infograf`: PASS for server startup and bind verification on `:8080` when run outside the sandbox; sandboxed socket binding remains restricted.
- UI/browser screenshots: still blocked locally because no runnable browser path is available in this environment.
