# Task 05 — Integration Review and Final Rescue Gate

Perform a full integration review after the rescue tasks. Do not assume prior agents completed their work correctly.

## Architecture

Verify:

```text
Source
  ↓
Source Adapter
  ↓
Canonical Series/Event
  ↓
Expression Engine
  ↓
Scene Binding
  ↓
Scene State
  ↓
SVG Renderer
```

Look for duplicated or bypassed logic.

## Determinism

Confirm same data + timestamp + config = same scene state; timeline replay reconstructs state; trend comes from data; animation endpoints come from scene state; interpolation does not fabricate values.

## Source-Agnostic Test

Drive the same conceptual metric through at least two source paths, preferably fixture/CSV plus raw/plaintext/REST and Prometheus fixture if available. Confirm equivalent behavior.

## Expression Builder

Verify visual editing, YAML editing, round trip, min/max mapping, x/y selection, threshold duration, simulation, event binding, and action dry-run.

## Visual Review

Open the application and capture screenshots. Reject the result if stick figures remain as finished assets, the scene is secondary to chrome, the app resembles a generic admin dashboard, metric-to-motion relationships are unclear, exact values are missing, or severe/recovery states are visually indistinguishable.

## Database Replication Acceptance

Run healthy -> lag rising -> severe lag -> incident -> recovery -> healthy. Scrub backward/forward and confirm annotations and scene state stay synchronized.

## Deployment / Security Regression

Verify no Java, no accidental microservice sprawl, no mandatory GPU, no mandatory external CDN, Go modular-monolith target remains viable, local/Docker deployment works, and security boundaries were not weakened.

## Final Deliverable

Create `docs/rescue-final-review.md` with PASS/FAIL per acceptance area, remaining technical debt, remaining visual debt, remaining product gaps, screenshots, and concrete next milestones.

Do not write complete if major placeholders remain.
