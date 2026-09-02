# Infografics Visual Asset & Animation Skill

## Purpose

Create or acquire production-quality visual assets for **Infografics**, a metrics-to-animated-infographics platform.

This skill exists to prevent programmer-art placeholders, stick figures, generic icons, and decorative animation from being treated as finished scene assets.

## Core Rule

Visual assets must support deterministic data-driven animation.

The asset layer does not decide what metrics mean.

The scene engine provides semantic state such as:

```text
value
normalized_value
trend
rate
acceleration
severity
event_state
recovery_state
```

This skill turns those states into visually expressive assets and animation-ready components.

## Asset Source Priority

Use this order:

1. existing approved project assets;
2. public-domain assets;
3. CC0 assets;
4. permissively licensed assets suitable for redistribution/modification;
5. original assets created specifically for Infografics.

Do not use assets with unclear licensing.

Do not copy a copyrighted modern visual interpretation of a public-domain character.

Record provenance for every externally sourced asset.

## Required Asset Metadata

Each sourced asset must record:

```yaml
asset:
  name:
  source:
  original_author:
  source_url:
  license:
  license_url:
  public_domain_basis:
  modifications:
  attribution_required:
```

## Supported Visual Styles

Create scene assets in styles appropriate to the intended audience.

Supported styles include:

- polished SVG/vector infographic art;
- professional illustration;
- 16-bit pixel art;
- 32-bit pixel art;
- retro game-style animation;
- diagrammatic technical illustration;
- stylized public-domain characters;
- user-supplied branded artwork.

Do not default every scene to the same visual style.

## Finished Art Standard

Finished artwork must not consist primarily of:

- stick figures;
- circles and rectangles used as characters;
- framework-default icons;
- emoji;
- crude programmer art;
- temporary placeholders;
- unstyled stock SVGs.

If temporary placeholders are necessary during implementation, clearly mark them as placeholders and do not declare visual work complete.

## Animation-Ready Decomposition

Whenever practical, artwork must be decomposed into meaningful addressable components.

Example:

```text
sisyphus
├── body
├── head
├── left_arm
├── right_arm
├── left_leg
├── right_leg
├── boulder
├── facial_expression
├── annotation_anchor
└── motion_anchor
```

Example:

```text
icarus
├── body
├── left_wing
├── right_wing
├── arms
├── legs
├── expression
├── incident_bubble
├── recovery_prop
└── motion_anchor
```

## Vector Assets

Prefer SVG when appropriate.

SVG components must:

- use stable element/group IDs;
- support translation;
- support rotation;
- support scaling;
- support opacity changes;
- support visibility/state changes;
- support path-based movement where appropriate;
- remain resolution independent.

Do not flatten an animation-ready SVG into one opaque image unless required.

## Pixel-Art Assets

For 16-bit or 32-bit style scenes:

- create intentional sprite sheets or layered sprites;
- use a consistent visual scale;
- maintain transparent backgrounds where appropriate;
- provide named animation states;
- preserve nearest-neighbor scaling;
- do not create unnecessary frame counts.

Example states:

```text
idle
walk
run
push
slip
fall
recover
warning
success
```

Animation states must be selected by deterministic scene state.

## Animation Principle

Never create arbitrary looped animation for a metric-bound behavior.

Wrong:

```text
character randomly walks left and right forever
```

Correct:

```text
normalized metric -> x position
rate of change -> animation speed
trend sign -> direction
event state -> reaction animation
```

Animation frames interpolate or illustrate data-derived state.

They do not manufacture state.

## Data-to-Art Mapping Examples

### Sisyphus

```text
remaining workload -> boulder size
progress -> hill position
rate of progress -> push/walk speed
negative rate -> slipping/backward movement
blocker event -> obstacle
recovery -> resumed climb
```

### Icarus

```text
availability -> altitude
availability rate -> vertical velocity
degradation -> unstable wing animation
outage -> dive/fall
incident -> speech/callout
recovery -> repair/recovery animation
```

### Database Replication

```text
replication rate -> packet movement speed
lag -> distance/separation
lag trend -> worsening/recovery movement
errors -> warning animation
incident -> callout
recovery -> restoration animation
```

## Visual Quality Review

Before marking visual work complete:

1. run the actual application;
2. inspect the scene at multiple values;
3. inspect increasing and decreasing trends;
4. inspect event states;
5. capture screenshots;
6. verify exact values remain readable;
7. verify the artwork looks intentional and presentation-ready.

If the result resembles programmer art or generic dashboard decoration, continue working.

## Relationship to Scene Engine

This skill may create:

- assets;
- sprites;
- SVG groups;
- animation definitions;
- scene-pack metadata.

It must not:

- rewrite datasource semantics;
- introduce source-specific calculations into artwork;
- bypass the canonical scene-state model;
- hard-code arbitrary metric behavior inside assets.

The scene engine owns data semantics.

The asset system owns visual expression.
