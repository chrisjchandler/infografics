# Infograf Design Doc

## Overview

Infograf is a greenfield monitoring-visualization platform that turns metrics, events, and historical data into animated infographic scenes for live monitoring, replay, and explanation.

The goal is to let a beginner upload a spreadsheet and get something useful quickly, while an advanced operator can use connectors, YAML, revision history, RBAC, and external hooks without switching to a different product.

Working product name in prose: **Infograf**.
Existing folder name: **Crayons**.

## Product principles

- Anything in, visual explanation out.
- One canonical semantic model across all inputs.
- WYSIWYG and YAML are two views of the same object.
- Local visual effects and external actions are separate.
- The platform should be useful for both small installs and enterprise deployments.
- The scene should be previewable, replayable, and exportable.

## First-slice scope

### In scope
- Go modular monolith backend.
- Embedded frontend asset serving.
- SQLite-backed metadata/control plane.
- Revision history and named snapshots.
- Basic auth scaffolding.
- One demo replication scene.
- One source adapter path.
- One expression/rule model.
- Scene preview and timeline concepts.
- One action hook path.

### Out of scope for the first slice
- Full marketplace packaging.
- Complete plugin sandboxing implementation.
- Production-grade multi-tenancy.
- Every possible source connector.
- Every possible scene template.
- Deep analytics and advanced transforms.

## User stories

### Beginner
- Upload a CSV/XLSX sheet.
- Pick a field that represents current value.
- Choose a visual metaphor.
- Preview and export.

### Operator
- Connect Prometheus, Graphite, REST, or OTEL.
- Define thresholds and durations.
- Attach alerting hooks.
- Replay history and inspect the timeline.

### Enterprise admin
- Configure local accounts, LDAP, OIDC, and mTLS.
- Control access by workspace and resource.
- Track immutable audit history.
- Restore snapshots and validate backups.

## Architecture

### Backend
- Go modular monolith.
- REST control plane.
- Optional SSE/WebSocket for live scene state.
- SQLite for metadata and control data.
- Source registry + semantic model + expression engine + revision engine + action engine.

### Frontend
- TypeScript.
- SVG-first scene rendering.
- Normalized coordinates rather than raw pixels.
- Simulation mode with test sliders and event triggers.
- Direct YAML editor and round-trip parsing.

### Data path
1. Source connector ingests a sample.
2. Semantic layer normalizes it into canonical fields.
3. Expression engine evaluates thresholds and state transitions.
4. Scene engine maps state into SVG properties and animation cues.
5. Action engine emits local or external hooks.
6. Timeline engine preserves replayable state.

## Canonical entities

For the first slice, focus on:

- Sample
- Series
- SeriesDescriptor
- MetricSemantics
- Event
- Annotation
- Expression
- Rule
- SceneBinding
- SceneState
- ActionInvocation

## Demo scene: database replication

The initial demo scene uses replication because the signals are easy to understand:

- Primary node
- Replica A: healthy
- Replica B: warning
- Replica C: critical
- Suggested actions: flash, sound alarm, PagerDuty trigger, annotation

Example rule:

```text
when replica_lag_ms > 100 for 2m -> warning
when replica_lag_ms > 500 for 30s -> critical
```

The same rule should be editable in WYSIWYG controls and representable in YAML.

## Suggested repo layout

```text
Crayons/
  cmd/infograf/
  internal/
    app/
    model/
    store/
    security/
    rules/
    scenes/
    connectors/
  frontend/
    src/
    dist/
  docs/
  tests/
```

## First MVP milestone

1. Start the Go server.
2. Serve the embedded landing page.
3. Return health/version/demo replication JSON.
4. Render a usable replication mock scene.
5. Add the first persistence layer for revisions and scene definitions.
6. Add one source adapter and one action hook.

## Open questions

- Which frontend framework should own the production editor shell?
- Which authentication backend should be first-class for the first release: local, LDAP, or OIDC?
- Should plugins start as WASM or out-of-process gRPC?
- Which real datasource should become the first connector after the demo scene?
- Which exporter should be prioritized first: SVG, PNG, or PDF?

## Next engineering moves

- Freeze the canonical model interfaces.
- Build the Go control plane skeleton.
- Connect one live sample source.
- Wire the replication demo scene.
- Replace hardcoded data with persisted scene state.
- Then move to auth, revisioning, and plugin contracts.
