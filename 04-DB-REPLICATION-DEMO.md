# Task 04 — Database Replication Acceptance Demo

Create a deterministic database-replication demo that proves the Project Crayons architecture works.

## Dataset

```csv
timestamp,replication_lag_seconds,replication_rate_mb_s,write_tps,last_commit_age_seconds,errors_5m
2026-08-30T10:20:00Z,0.42,12.4,1254,2,0
2026-08-30T10:21:00Z,12.87,8.1,1020,22,0
2026-08-30T10:22:00Z,78.35,3.2,320,150,1
2026-08-30T10:23:00Z,55.12,5.7,540,95,0
2026-08-30T10:24:00Z,2.31,11.0,1201,3,0
2026-08-30T10:25:00Z,0.38,12.6,1310,2,0
```

The sequence must visibly pass through healthy, lag rising, severe lag, incident active, recovery, healthy again.

## Event

```yaml
event:
  id: INC-REPL-001
  type: replication_incident
  start: 2026-08-30T10:22:00Z
  end: 2026-08-30T10:24:00Z
  title: Replication lag exceeded recovery threshold
  attributes:
    ticket: INC-REPL-001
    severity: SEV2
```

## Scene Behavior

```text
PRIMARY DB       REPLICATION PATH       REPLICA DB
```

Bind replication lag to separation/stress/replica position, replication rate to packet motion speed, lag trend to worsening/recovery behavior, errors to warning state, last commit age to supporting metric, and the incident to an integrated callout.

## Timeline

At each fixture timestamp the same state must reconstruct. Support scrub, play, pause, jump to incident, and return to live/test mode.

## Manual Test

Provide manual lag override and test 0%, 25%, 50%, 75%, 100%. Bound x/y/property values must update immediately and predictably.

## Action Demo

```text
lag > 60 seconds for 30 seconds
  -> show Lag High state
  -> flash scene
  -> PagerDuty action in dry-run/test mode
```

Do not actually page anyone.

## Export

Produce a static scene snapshot with no editor chrome, readable values, incident callout, and presentation-ready composition.

## Automated Tests

Verify value→normalized mapping, normalized→scene property, rate direction, timeline determinism, event active window, recovery state, and dry-run action behavior.
