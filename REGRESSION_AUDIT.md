# STOP IMPLEMENTATION — REGRESSION AUDIT ONLY

Do not modify application code.

The current repository has suffered significant regressions during recent agent-driven refactoring.

Your only task is to identify them.

Use Git history, current source, screenshots, tests, documentation, and existing configuration examples.

Compare the current HEAD against the last revision before the recent rescue/refactor work.

Identify every capability that was:

- removed;
- simplified;
- stubbed;
- converted into a placeholder;
- disconnected;
- hard-coded;
- visually degraded;
- made datasource-specific.

At minimum investigate:

1. Expression builder
2. Datasource configuration
3. Prometheus support
4. Graphite support
5. OpenTelemetry support
6. REST/generic source support
7. CSV/XLSX support
8. Raw streaming input
9. Scene bindings
10. X/Y coordinate mapping
11. Trendline behavior
12. Animation
13. Timeline playback
14. Annotation/event handling
15. YAML round trip
16. Export behavior

For every regression produce:

| Capability | Previous state | Current state | Commit/change responsible | Required restoration |
|------------|----------------|---------------|---------------------------|----------------------|

Create:

docs/REGRESSION-AUDIT.md

Do not write implementation code.

Do not decide that removed functionality was unnecessary.

Do not reclassify implemented capabilities as future features.

When uncertain whether something previously worked, treat it as requiring investigation rather than assuming it was placeholder functionality.
