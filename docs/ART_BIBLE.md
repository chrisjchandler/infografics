# Infograf Game-Art Bible

Status: approved research baseline; scene art has **not** yet been approved.

## 1. Product visual target

Infograf uses high-detail **32-bit-era pixel art** for animated telemetry stories. The target is modern, readable game art with mythic silhouettes—not procedural SVG figures, clip art, or literal 32×32 icons.

Inspiration only (do not copy shipped art):

- Dead Cells production pipeline: https://www.gamedeveloper.com/production/art-design-deep-dive-using-a-3d-pipeline-for-2d-animation-in-i-dead-cells-i-
- Blasphemous mythic silhouettes: https://www.hollywoodreporter.com/movies/movie-news/blasphemous-designer-purist-approach-games-pixel-art-style-1238370/
- Owlboy character acting: https://www.gamedeveloper.com/art/using-graphics-to-express-character-and-theme-in-i-owlboy-i-
- Shovel Knight sprite process: https://old.yachtclubgames.com/2020/03/creating-a-shovel-knight-character-sprite/
- Saint11/Pedro Medeiros tutorials: https://saint11.art/blog/pixel-art-tutorials/

## 2. Fixed scene coordinate system

- Logical scene canvas: **480×270**.
- Scale to the display at integer multiples only: 960×540, 1440×810, etc.
- `image-rendering: pixelated`; renderer antialiasing disabled.
- Integer sprite positions; no subpixel sprite placement.
- Fixed side-on orthographic camera with mild 2.5D layering.
- No zooming when metrics spike.
- Bottom 18–20% is reserved for time/source metadata and foreground footing.
- Important silhouettes stay inside X 7–93%, Y 10–78%.

## 3. Character scale and sprite cells

| Scene | Visible art size | Atlas frame cell | Anchor |
|---|---:|---:|---|
| Sisyphus | human 38–44×64 px; boulder 48–56 px diameter | 128×96 | rear foot on telemetry point |
| Icarus | body 34–40×64 px; wingspan 112–144 px | 160×112 | pelvis/chest on telemetry point |
| Hydra | body 112–144×80 px; 3–5 necks up to 144 px high | 192×144 | body root on telemetry point |
| Medusa option | 56–72×88 px; snake halo 72–88 px | 96×112 | feet/body root |
| Kraken option | mantle 96–128×80 px; total reach 160–192×128 px | 192×144 | waterline/root |

Scale is fixed per scene class. **Metric magnitude never changes character scale.** Position, pose, effects, and environment communicate the data.

## 4. Required silhouettes

### Sisyphus

- Forward-pointing triangular silhouette.
- Planted rear foot, bowed back, straight pushing arms.
- Hands visibly touch the boulder: 0–2 logical pixels of separation.
- Actor and boulder are authored as one coordinated animation or joined through frame-authored hand/contact sockets.
- Rising: forward lean, strain, determined expression.
- Falling: boulder leads downhill, body recoils/leans back, alarmed expression.
- Steady: braced pose and controlled breathing.

### Icarus

- Recognizable human torso suspended between two feathered wings.
- Wingspan 112–144 logical pixels; body remains readable at 34–40×64.
- Wings are part of authored frames or attached through invariant shoulder sockets.
- Never animate independent DOM/SVG wing transforms.
- Rising: powered flap/euphoric expression.
- Steady: asymmetric glide.
- Falling: folded/chaotic wings, panic, falling feathers/wax accents.

### Hydra

- One connected low body mass.
- **3–5** serpentine necks with clear negative space between heads.
- No smiling faces, repeated bean shapes, flower stems, or equal-height neck rows.
- Heads have reptilian wedges, brows, jaws, teeth/tongue accents, and varied orientation.
- Neck sway is phase-offset; the body root remains fixed to telemetry.
- Error/retry multiplication may add temporary ghost necks or venom effects, but base anatomy stays coherent.

## 5. Shared palette

Master palette: 24–32 colors; target 16–22 visible colors per scene.

Neutral ramp:

- Ink `#0B1020`
- Deep blue-black `#171D2C`
- Slate `#27344A`
- Mid slate `#3D5269`
- Pale stone `#AAB7C4`
- Warm light `#E8E3D5`

Scene ramps:

- Sisyphus: `#3B2F2A`, `#6A4B3B`, `#A26D4F`, `#D7A06A`
- Icarus: `#173B5F`, `#246B8E`, `#40A7B5`, `#89D4D1`, `#F2B84B`, `#F8E7A1`
- Hydra/sea monsters: `#0E3040`, `#15535A`, `#238073`, `#D0E85B`, `#B93B4B`, `#F06A5B`

Rules:

- Brightest colors are reserved for face/hands, telemetry core, and alert effects.
- Backgrounds are lower-saturation and lower-contrast than actors.
- Telemetry line: 1 px dark keyline plus 2 px bright core at logical resolution.
- Use hue-shifted ramps, not gray-only shading.

## 6. Animation clips

Authored character frames are stepped. World/path interpolation remains timestamp-based at display refresh rate.

| Clip | Frames | Authored playback |
|---|---:|---:|
| Idle/breath | 4–6 | 6–8 fps |
| Sisyphus push/walk | 8 | 10–12 fps |
| Sisyphus strain/brace | 6 | 8 fps |
| Boulder roll | 8 hand-pixelled orientations | distance synchronized |
| Icarus powered flap | 8 | 10–12 fps |
| Icarus glide | 4 | 6 fps |
| Icarus panic/tumble | 8 | 10–12 fps |
| Hydra neck sway | 6–8 per neck | 6–8 fps, phase offset |
| Hydra strike/recoil | 6 | 10–12 fps |
| Dust/feathers/venom/spray | 4–8 | 10–12 fps |

Shared semantic state contract:

- `steady_idle`
- `rising_move`
- `falling_move`
- `rising_arrive`
- `falling_arrive`
- `steady_arrive`
- optional `enter`, `exit`, myth-specific accents
- `reduced_static`

State classification uses metric values over a 3–5-sample window. New states hold at least 180 ms; direct rising↔falling reversal requires 1.5× the entry threshold.

## 7. Renderer architecture

Use **one shared PixiJS v8 application** and one ticker.

```text
HTML application shell
├── semantic title, value, trend, timestamps, axes, controls
└── one PixiJS canvas
    ├── static/background layer
    ├── telemetry-path layer (refresh/resize only)
    ├── actorRoot (telemetry position only)
    │   └── visualContainer
    │       ├── AnimatedSprite
    │       ├── frame-authored props/sockets
    │       └── pooled effects
    └── foreground/effects layer
```

- Only the selected/expanded scene animates at full rate.
- Inactive scenes use cached poster frames.
- SVG is limited to static icons/export—not articulated characters.
- Animated WebP/AVIF is limited to non-interactive ambient loops.
- WebGL failure shows a static poster frame rather than a second animation engine.

### Anchor contract

- `actorRoot.position` is the only telemetry-path transform.
- Ground pivot: approximately `(0.50, 0.88)` normalized.
- Flight pivot: approximately `(0.50, 0.55)`.
- Export sockets: `leftHand`, `rightHand`, `head`, `propContact`, `effectOrigin`.
- Facing is mirrored on `visualContainer`, never `actorRoot`.
- Character scale and breathing never modify `actorRoot`.

## 8. Telemetry and chart integrity

- Time maps monotonically to X.
- Value maps numerically to Y.
- Every scene displays numerical Y ticks with units and UTC X timestamps.
- Actor movement is linear in data time; no per-segment easing.
- Binary-search surrounding samples and interpolate within the real timestamp interval.
- Actor contact anchor stays on the actual polyline.
- At pause/end, actor is at the latest/rightmost sample.
- Tangent-derived sprite rotation is limited to ±12°; stronger lean is authored in the clip.
- Time/source/refresh metadata remains semantic HTML and is never hidden by actor art.

## 9. Performance budget

- 60 fps target; degraded 30 fps mode.
- Animation update ≤2 ms typical, ≤4 ms p95.
- Complete frame ≤10 ms typical, ≤12 ms p95.
- No animation-caused task over 50 ms.
- Prefer ≤200 visible Pixi objects and ≤20 draw calls.
- No per-frame allocations in the hot loop.
- Pause ticker when hidden/offscreen/no active scene.
- Default atlas max 2048×2048; prefer 1024×1024 where practical.
- Current scene: at most two atlases; transition: one extra atlas.
- Resident textures ≤64 MiB desktop, ≤32 MiB constrained mode.
- Initial scene asset transfer target 1.5–2 MiB.
- Preload current and next scene only; unload obsolete assets.

If p95 frame time exceeds 12 ms for two seconds:

1. Disable particles/filters.
2. Reduce renderer resolution toward 1.
3. Render world at 30 fps while keeping timestamp-based movement.
4. Fall back to static latest-state pose.

## 10. Layer order

1. Sky/abyss field
2. Far silhouettes
3. Midground landmarks
4. Telemetry terrain ribbon
5. Numerical axes, grid, telemetry line, samples
6. Actor and attached prop/body
7. Appendages/effects
8. Sparse foreground occluders
9. Semantic UI/time band

## 11. Vertical-slice acceptance gate

Build **Sisyphus only** before converting Icarus or Hydra.

It passes only if:

- silhouette reads as a man pushing a boulder at thumbnail size;
- boulder remains connected to both hands throughout all clips;
- actor occupies 24–27% of logical scene height;
- numerical X/Y scales are readable and unobscured;
- rising, falling, and steady moods read without explanatory text;
- latest-frame placement is correct;
- no full DOM rebuild occurs on play/pause;
- scene transition p95 is below 100 ms after assets are warm;
- active frame p95 is ≤12 ms;
- close/hidden state stops the ticker and releases outgoing assets;
- reduced-motion mode preserves all telemetry information.

Do not convert Icarus/Hydra until this vertical slice receives visual approval.

## 12. Authoritative runtime references

- MDN Canvas optimization: https://developer.mozilla.org/en-US/docs/Web/API/Canvas_API/Tutorial/Optimizing_canvas
- MDN requestAnimationFrame: https://developer.mozilla.org/en-US/docs/Web/API/Window/requestAnimationFrame
- MDN WebGL best practices: https://developer.mozilla.org/en-US/docs/Web/API/WebGL_API/WebGL_best_practices
- PixiJS performance: https://pixijs.com/8.x/guides/concepts/performance-tips
- PixiJS assets: https://pixijs.com/8.x/guides/components/assets
- PixiJS AnimatedSprite: https://pixijs.download/release/docs/scene.AnimatedSprite.html
- MDN reduced motion: https://developer.mozilla.org/en-US/docs/Web/CSS/@media/prefers-reduced-motion
- WCAG animation from interactions: https://www.w3.org/WAI/WCAG22/Understanding/animation-from-interactions.html
