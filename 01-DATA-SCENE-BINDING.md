# Task 01 — Deterministic Data-to-Scene Binding

Implement or repair the binding pipeline so numerical data actually drives visual properties.

## Required Mapping

```text
normalized = clamp((value - min) / (max - min), 0, 1)
```

Support inversion with `mapped = 1 - normalized`.

Support target properties: x, y, scale, rotation, opacity, path progress, animation speed, label text, and discrete state.

Use normalized/resolution-independent scene coordinates. Do not bind raw metric values directly to browser pixels.

## Trend Model

Calculate delta, rate of change, increasing, decreasing, stable, acceleration/deceleration, and volatility where configured. Trend must derive from timestamped data, not frontend guesses or timers.

## Animation Model

Separate data update frequency from visual interpolation. If scene state changes from `x: 0.25` to `x: 0.70`, the browser may tween that movement; the endpoints are data, the tween is presentation.

## Example Binding

```yaml
binding:
  id: replica-lag-position
  source:
    ref: db-prometheus
  metric:
    name: replication_lag_seconds
  interpretation:
    type: gauge
    preferred_direction: lower
    range:
      min: 0
      max: 120
  target:
    object: replica
    property: x
  mapping:
    direction: inverse
    output:
      min: 0.60
      max: 0.90
  animation:
    duration_ms: 500
    easing: ease-out
```

## Validation Fixture

For `min=0`, `max=100`, `output x=0.20..0.80`, verify:

```text
0   -> 0.20
25  -> 0.35
50  -> 0.50
75  -> 0.65
100 -> 0.80
```

Also verify clamping, inversion, and clear failure when min == max.

Trend fixtures must prove increasing, decreasing, and stable behavior.

## UI Acceptance

Provide a simulation value. Changing it must visibly move the object. Changing target from x to y must change the responding axis. Changing min/max must immediately alter mapping.

## Deliverables

Canonical binding implementation, tests, validation errors, one interactive demo binding, and concise technical documentation. Do not add scene-specific hacks.
