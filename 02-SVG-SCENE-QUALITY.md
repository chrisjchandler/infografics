# Task 02 — SVG Scene and Visual Quality Rescue

Repair the visual layer so Project Crayons looks like an infographic authoring application rather than a generated administration dashboard.

## Scene-First UI

Prioritize a large central canvas, restrained toolbars, whitespace, clear hierarchy, obvious metrics, direct visual feedback, and presentation-quality typography.

Avoid endless card grids, generic admin chrome, tiny scene previews, framework-default aesthetics, and excessive borders/badges.

## SVG Requirements

Use SVG groups with stable identifiers, for example:

```xml
<g id="replication-scene">
  <g id="primary-db">...</g>
  <g id="replica-db">...</g>
  <g id="replication-path">...</g>
  <g id="data-packets">...</g>
  <g id="incident-bubble">...</g>
  <g id="metric-labels">...</g>
</g>
```

## Database Replication Art Direction

- recognizable illustrated database/server characters or polished iconographic objects;
- replication path clearly communicates motion;
- data packets/blocks move according to replication rate;
- increasing lag creates visible separation/tension;
- severe lag is obvious;
- recovery visibly restores normal relationship;
- exact values remain readable;
- incident annotations appear as integrated infographic callouts.

The scene may be whimsical without looking childish.

## Scene Pack Structure

```text
scene.svg
manifest.yaml
mappings.yaml
animations.yaml
event-behaviors.yaml
thumbnail.svg
LICENSE
```

## Validation

Run the app and capture screenshots for healthy, lag increasing, severe lag, incident active, and recovery. Compare to supplied Crayons mockups. If the result still reads as a generic generated dashboard, continue working.

## Deliverables

Polished replication scene, improved authoring layout, reusable scene-pack conventions, screenshots under `docs/screenshots/`, and `docs/visual-system.md`.
