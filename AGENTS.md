## Project Rules

### NEVER

* introduce Java;
* add a microservice without explicit approval and a demonstrated technical need;
* use random, canned, timer-driven, or decorative animation for metric-bound behavior;
* substitute stick figures, programmer art, or placeholder assets for finished scene assets;
* introduce a GPU requirement for normal operation;
* add externally hosted CDN or runtime dependencies required for production operation;
* duplicate semantic calculations independently in the frontend and backend;
* hard-code scene movement where a metric, expression, event, trend, or configured rule should determine it;
* hide broken or incomplete metric-to-visual mappings behind decorative animation;
* sacrifice source portability, scene portability, or deterministic replay for implementation convenience.

### ALWAYS

* derive scene state deterministically from source data, semantic interpretation, expressions, events, and explicit configuration;
* provide automated tests for value → normalized value → scene coordinate/property mappings;
* keep scenes datasource-agnostic;
* preserve WYSIWYG/YAML/API/config-as-code round-trip compatibility;
* run the actual UI before declaring visual work complete;
* inspect screenshots after visual changes and compare them against the intended visual-quality standard;
* create polished, production-quality visual assets suitable for an infographic product rather than developer placeholders;
* support reusable SVG/vector artwork as the primary scene format, with optional 16-bit, 32-bit, pixel-art, illustrated, or other visual styles where appropriate;
* when adapting public-domain artwork or characters, preserve license/provenance information and create animation-ready derivatives with addressable components;
* remember that this product is **Infografics**, a metrics-to-animated-infographics platform, not a conventional dashboard product;
* preserve the mathematical relationship between time, value, trend, and visual state even when a scene does not visibly display conventional X/Y axes;
* where a scene uses conventional chart semantics, ensure X and Y coordinates accurately represent their configured dimensions, typically time and value;
* allow metaphorical scenes to transform those same underlying values into position, distance, altitude, scale, rotation, speed, path progress, state, opacity, or other explicit visual properties;
* support as many practical ingestion methods as possible through source-aware adapters and a stable connector interface;
* provide first-class paths for both conventional telemetry systems and simple/raw data ingestion;
* design every authoring workflow for a user spectrum ranging from novice to advanced;
* expose simple semantic choices to novice users while retaining expressions, YAML, APIs, and advanced transformations for expert users;
* infer source semantics where the datasource supplies them rather than forcing users to understand source-specific functions;
* maintain exact numeric values and meaningful labels alongside metaphorical visualization where appropriate;
* treat annotations and events as first-class timeline data and keep them synchronized with playback, snapshots, and live state;
* ensure visual state can be reconstructed deterministically at an arbitrary historical timestamp;
* benchmark performance where relevant, run regression tests after material changes, and avoid degrading existing functionality while implementing new features;
* prefer correcting an underlying abstraction over repeatedly patching symptoms;
* treat visual quality, usability, deployability, security, performance, and deterministic data representation as product requirements rather than optional polish.

## Visual Asset Standard

Finished scene artwork must look intentionally designed.

Acceptable styles include:

* polished vector illustration;
* professional infographic illustration;
* 16-bit or 32-bit pixel art;
* appropriately stylized public-domain characters;
* user-supplied/custom scene artwork.

Finished scene artwork must not consist of:

* stick figures;
* generic geometric placeholders presented as final art;
* framework-default icons used as the principal scene artwork;
* crude programmer-generated characters;
* unrelated decorative animation.

Artwork intended for animation should expose meaningful addressable components such as:

```text
character
├── body
├── head
├── left-arm
├── right-arm
├── expression
├── carried-object
├── annotation-anchor
└── motion-anchor
```

The renderer should animate those components from deterministic scene state rather than cycling through canned animations unrelated to data.

## Data Visualization Principle

Infografics may visually replace a conventional chart, but it must not discard the underlying quantitative model.

Conceptually:

```text
time/value/events
        ↓
semantic interpretation
        ↓
normalized quantitative state
        ↓
scene mapping
        ↓
infographic representation
```

For a conventional chart:

```text
time  → X
value → Y
```

For Sisyphus:

```text
time              → timeline/playback position
remaining work    → boulder size
progress          → hill position
rate of progress  → movement velocity
negative progress → backward movement
```

For Icarus:

```text
time                  → timeline/playback position
availability          → altitude
availability trend    → vertical velocity
severity/event state  → character behavior
incident annotation   → callout
```

The metaphor may change.

The mathematics may not.
