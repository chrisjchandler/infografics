# Task 03 — Expression Builder and WYSIWYG/YAML Round Trip

Implement or repair the Project Crayons expression builder for both Crayola-level and advanced users.

## Core Rule

Visual editor, YAML editor, API, and config-as-code manipulate the same underlying model.

## Beginner Workflow

A user can configure source, metric, interpretation, min/max, visual object, target property, mapping direction, thresholds, animation, events, local actions, and external actions.

Example:

```text
WHEN Replication Lag is greater than 10 seconds
FOR 30 seconds
THEN show "Lag High"
AND trigger PagerDuty
```

## Advanced YAML

```yaml
binding:
  id: replication-lag
  source:
    ref: prometheus-prod
  metric:
    name: postgres_replication_lag_seconds
  interpretation:
    type: gauge
    preferred_direction: lower
  boundaries:
    min: 0
    max: 120
  target:
    object: replica-db
    property: x
  states:
    - id: healthy
      when: value < 5
    - id: lag-high
      when: value >= 10
      for: 30s
  actions:
    - when: state == lag-high
      action: screen.flash
    - when: state == lag-high
      action: pagerduty.trigger
```

## Round Trip

Visual Editor -> YAML -> manual YAML edit -> Visual Editor must preserve semantics with actionable validation errors.

## Simulation Mode

Support value input/slider, timestamp progression, increasing/decreasing/stable state, event trigger buttons, threshold preview, and action dry-run.

## Action Safety

Separate local visual actions from external actions. External actions require explicit authorization and dry-run support. Do not create an unrestricted webhook SSRF mechanism.

## Required Tests

Visual↔YAML preservation, min/max effects, x/y selection, threshold duration, invalid-expression handling, and action dry-run.
