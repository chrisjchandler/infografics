# Task 00 — Audit and Rescue the Existing Implementation

Take ownership of the existing Project Crayons repository. Do not begin by adding features.

## Known Symptoms

- animation objects do not reliably follow source values;
- visual position does not correctly map to configured x/y bounds;
- increasing/decreasing trends do not reliably alter direction;
- velocity/rate-of-change is not consistently reflected in motion;
- animations appear canned or decorative;
- visual assets are crude stick figures/placeholders;
- UI resembles a generic generated Node/React admin application;
- prior fixes changed symptoms without correcting the underlying state model.

Treat these as probable architecture defects.

## Repository Audit

Document runtime/deployment, languages/frameworks, backend/frontend entrypoints, build system, process/container topology, persistence, websocket/SSE paths, and external dependencies.

Trace one datapoint from source to visible scene through:
1. source connection;
2. source adapter;
3. sample representation;
4. semantic interpretation;
5. expression/rule evaluation;
6. scene binding;
7. scene-state construction;
8. x/y/property mapping;
9. frontend state transport;
10. SVG/DOM update;
11. animation interpolation.

Classify each step as correct, incomplete, placeholder, duplicated, hard-coded, or architecturally wrong.

Also trace manual/API events, event persistence, historical lookup, visual event state, timeline replay behavior, and whether visual/YAML configuration share a real model.

## Required Output

Create `docs/rescue-audit.md` containing current architecture, expected architecture, findings, critical defects, components safe to retain, components to replace, and correction sequence.

## Foundational Corrections

After the audit, implement only enough to establish:

```text
SourceAdapter
  ↓
Canonical Series/Event
  ↓
Expression Engine
  ↓
Scene Binding
  ↓
Scene State
  ↓
Renderer
```

Do not polish visuals until this path is real.

## Acceptance Criteria

- one source sample can be traced through code to one scene property;
- no random animation is required for that property;
- same timestamp/data/config yields same scene state;
- scene state is source-agnostic;
- x/y mapping has one defined implementation path;
- core model-boundary tests exist;
- `docs/rescue-audit.md` reflects the repaired state.

Do not declare completion merely because the app builds.
