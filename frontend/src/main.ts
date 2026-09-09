type PeriodKey = '1h' | '6h' | '24h' | '7d';

interface PeriodOption {
  key: string;
  label: string;
  window: string;
}

interface GraphiteCard {
  label: string;
  metric: string;
  value_display: string;
  delta_display: string;
  detail: string;
  status: string;
  progress: number;
  live: boolean;
  points?: number[];
}

interface GraphiteOverview {
  connected: boolean;
  source_url: string;
  source_label: string;
  period: string;
  period_label: string;
  window: string;
  captured_at: string;
  cards: GraphiteCard[];
  notes: string[];
  periods: PeriodOption[];
}

interface SceneTemplate {
  id: string;
  title: string;
  type: string;
  summary: string;
  noteBig: string;
  noteSmall: string;
  joke: string;
  art: 'hill' | 'storm' | 'rope';
  focusMetric: string;
  metricOrder: string[];
  beats: [string, string][];
  narrative: string;
  hook: string;
}

const scenes: SceneTemplate[] = [
  {
    id: 'sisyphus-growth',
    title: 'Sisyphus Growth',
    type: 'Monitor',
    summary: 'A boulder climbs the hill of throughput, because growth is apparently still an uphill commute.',
    noteBig: 'Growth is legible when it looks like a stubborn little myth.',
    noteSmall: 'We turn 200s, ingest volume, and error drift into a hill that a human can read in one glance.',
    joke: 'Sisyphus would absolutely label the axes.',
    art: 'hill',
    focusMetric: '',
    metricOrder: [],
    beats: [
      ['Observe', 'Growth is the hill.'],
      ['Normalize', 'Throughput becomes a climb.'],
      ['Compose', 'Boulder, slope, and a caption do the work.'],
      ['Publish', 'The metaphor makes the curve memorable.'],
    ],
    narrative: 'This scene makes a growth metric feel less like a chart and more like a story with gravity, effort, and a tiny bit of absurdity.',
    hook: 'A growth line becomes a hill; the boulder is your throughput; the joke is that the climb is the product.',
  },
  {
    id: 'weather-404',
    title: '404 Weather Report',
    type: 'Decide / Learn',
    summary: 'Missing routes fall as rain while success clears the sky. The joke is weather; the point is that errors become obvious.',
    noteBig: 'The wrong routes are just weather with bad manners.',
    noteSmall: 'A 404 spike becomes rain, while the 200s clear the sky. It is hard to ignore weather when it is drawn like weather.',
    joke: 'The cloud is not angry. It is just failing loudly.',
    art: 'storm',
    focusMetric: '',
    metricOrder: [],
    beats: [
      ['Observe', 'Errors show up as rain.'],
      ['Normalize', 'Mistakes get a forecast.'],
      ['Compose', 'Cloud, drops, and a warning badge.'],
      ['Publish', 'A weather metaphor is easier to brief.'],
    ],
    narrative: 'This scene keeps error-rate visibility without making the page look like a spreadsheet with a mood problem.',
    hook: 'A 404 rainstorm is a friendlier way to explain missing routes than a wall of red numbers.',
  },
  {
    id: 'tightrope-latency',
    title: 'Latency Tightrope',
    type: 'Operate',
    summary: 'CPU and disk walk a wire above patience itself. The image says: we are close to the edge, but still upright.',
    noteBig: 'Every request is balancing on a wire with a cup of coffee.',
    noteSmall: 'Operational metrics become a tightrope act: if the rope sags, you know where to look.',
    joke: 'The trapeze act is powered by cache misses.',
    art: 'rope',
    focusMetric: '',
    metricOrder: [],
    beats: [
      ['Observe', 'Load becomes a balancing act.'],
      ['Normalize', 'CPU and disk share the rope.'],
      ['Compose', 'A wire, a walker, and a safety net.'],
      ['Publish', 'The caution is visible without reading a tooltip.'],
    ],
    narrative: 'This one is for when you want a lived-in operational metaphor instead of a cold traffic graph.',
    hook: 'A tightrope makes load feel precarious in a human way without hiding the actual numbers.',
  },
];

const periodLabels: Record<PeriodKey, string> = {
  '1h': 'Last hour',
  '6h': 'Last 6 hours',
  '24h': 'Last 24 hours',
  '7d': 'Last 7 days',
};

const storageKey = (name: string) => `infograf.${name}`;
const clamp = (n: number, min: number, max: number) => Math.max(min, Math.min(max, n));

const state: {
  period: PeriodKey;
  sceneId: string;
  selectedMetric: string;
  overview: GraphiteOverview | null;
  loading: boolean;
  surpriseIndex: number;
} = {
  period: (localStorage.getItem(storageKey('period')) as PeriodKey | null) || '24h',
  sceneId: localStorage.getItem(storageKey('scene')) || scenes[0].id,
  selectedMetric: localStorage.getItem(storageKey('metric')) || '',
  overview: null,
  loading: false,
  surpriseIndex: 0,
};

function byId<T extends HTMLElement>(id: string): T {
  const el = document.getElementById(id);
  if (!el) throw new Error(`Missing element: ${id}`);
  return el as T;
}

function selectedScene(): SceneTemplate {
  return scenes.find((scene) => scene.id === state.sceneId) || scenes[0];
}

function selectedPeriodOption(): PeriodOption {
  const period = state.overview?.periods?.find((p) => p.key === state.period);
  return period || { key: state.period, label: periodLabels[state.period], window: state.period };
}

function prioritizedCards(scene: SceneTemplate, cards: GraphiteCard[]): GraphiteCard[] {
  const lookup = new Map(cards.map((card) => [card.metric, card] as const));
  const ordered: GraphiteCard[] = [];
  for (const metric of scene.metricOrder) {
    const card = lookup.get(metric);
    if (card) {
      ordered.push(card);
      lookup.delete(metric);
    }
  }
  return ordered.concat([...lookup.values()]);
}

function getSelectedCard(scene: SceneTemplate, cards: GraphiteCard[]): GraphiteCard {
  if (state.selectedMetric) {
    const found = cards.find((card) => card.metric === state.selectedMetric);
    if (found) return found;
  }
  return cards.find((card) => card.metric === scene.focusMetric) || cards[0];
}

function sparklinePath(points: number[] | undefined, width = 220, height = 64): string {
  if (!points || points.length === 0) return '';
  const min = Math.min(...points);
  const max = Math.max(...points);
  const range = max - min || 1;
  const step = points.length === 1 ? 0 : width / (points.length - 1);
  return points
    .map((value, idx) => {
      const x = idx * step;
      const y = height - ((value - min) / range) * (height - 8) - 4;
      return `${idx === 0 ? 'M' : 'L'} ${x.toFixed(2)} ${y.toFixed(2)}`;
    })
    .join(' ');
}

function statusClass(status: string): string {
  return status === 'critical' ? 'critical' : status === 'warning' ? 'warning' : status === 'neutral' ? 'neutral' : 'healthy';
}

function formatTrendLabel(card: GraphiteCard): string {
  return `${card.delta_display || 'flat'} · ${card.points?.length || 0} samples`;
}

function renderPeriodButtons(): void {
  const host = byId<HTMLDivElement>('period-controls');
  const options = state.overview?.periods || Object.entries(periodLabels).map(([key, label]) => ({ key, label, window: key }));
  host.innerHTML = options.map((opt) => `<button class="period-btn ${opt.key === state.period ? 'active' : ''}" data-period="${opt.key}">${opt.label}</button>`).join('');
  host.querySelectorAll<HTMLButtonElement>('[data-period]').forEach((btn) => {
    btn.addEventListener('click', () => {
      const next = btn.dataset.period as PeriodKey | undefined;
      if (next && next !== state.period) {
        state.period = next;
        localStorage.setItem(storageKey('period'), next);
        void loadOverview();
      }
    });
  });
}

function renderSceneList(): void {
  const host = byId<HTMLDivElement>('scene-list');
  const current = selectedScene();
  host.innerHTML = scenes.map((scene) => `
    <button class="scene-item ${scene.id === current.id ? 'active' : ''}" data-scene="${scene.id}">
      <div class="title">${scene.title}</div>
      <div class="meta">${scene.summary}</div>
      <div class="foot">
        <span class="badge">${scene.type}</span>
        <span class="badge">${scene.art}</span>
      </div>
    </button>
  `).join('');
  host.querySelectorAll<HTMLButtonElement>('[data-scene]').forEach((btn) => {
    btn.addEventListener('click', () => {
      const next = btn.dataset.scene;
      if (!next) return;
      state.sceneId = next;
      state.selectedMetric = '';
      localStorage.setItem(storageKey('scene'), state.sceneId);
      localStorage.removeItem(storageKey('metric'));
      render();
    });
  });
}

function renderMetricGrid(scene: SceneTemplate, cards: GraphiteCard[]): void {
  const host = byId<HTMLDivElement>('metric-grid');
  const ordered = prioritizedCards(scene, cards);
  const selected = getSelectedCard(scene, cards);
  host.innerHTML = ordered.map((card) => `
    <article class="metric-card ${statusClass(card.status)} ${card.metric === selected.metric ? 'active' : ''}" data-metric="${card.metric}">
      <div class="metric-head">
        <div>
          <div class="metric-label">${card.label}</div>
          <div class="metric-value">${card.value_display}</div>
        </div>
        <div class="metric-delta">${formatTrendLabel(card)}</div>
      </div>
      <svg class="sparkline" viewBox="0 0 220 64" preserveAspectRatio="none" aria-hidden="true">
        <path d="${sparklinePath(card.points)}" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round"></path>
      </svg>
      <div class="metric-detail">${card.detail}</div>
      <div class="meter"><span style="width:${clamp(card.progress, 0, 100)}%;"></span></div>
    </article>
  `).join('');
  host.querySelectorAll<HTMLElement>('[data-metric]').forEach((card) => {
    card.addEventListener('click', () => {
      const next = card.dataset.metric;
      if (!next) return;
      state.selectedMetric = next;
      localStorage.setItem(storageKey('metric'), state.selectedMetric);
      render();
    });
  });
}

function renderBeatList(scene: SceneTemplate): void {
  const host = byId<HTMLDivElement>('beat-list');
  host.innerHTML = scene.beats.map(([step, text]) => `<div class="beat"><strong>${step}</strong><span>${text}</span></div>`).join('');
}

function renderSceneArt(scene: SceneTemplate, cards: GraphiteCard[], overview: GraphiteOverview | null): string {
  const focus = getSelectedCard(scene, cards);
  const support = cards.find((card) => card.metric !== focus.metric) || focus;
  const period = selectedPeriodOption();
  const progress = clamp(focus.progress || 0, 8, 95);
  const x = 120 + progress * 4.6;
  const y = 290 - progress * 1.45;
  const note = scene.joke;
  const caption = `${focus.label}: ${focus.value_display} (${focus.delta_display || 'flat'})`;
  const tail = (focus.points || []).slice(-12);
  const trail = tail.length ? sparklinePath(tail, 220, 64) : '';
  const accent = scene.art === 'storm' ? 'rgba(251, 191, 36, 0.9)' : scene.art === 'rope' ? 'rgba(244, 114, 182, 0.92)' : 'rgba(34, 211, 238, 0.92)';

  if (scene.art === 'storm') {
    const drops = Array.from({ length: Math.max(5, Math.round(progress / 12)) }).map((_, i) => `<rect x="${140 + i * 104}" y="${160 + (i % 2) * 12}" width="14" height="44" rx="8" transform="rotate(${i * 6 - 8} ${140 + i * 104} ${160 + (i % 2) * 12})" opacity="0.92" />`).join('');
    return `
      <svg class="scene-art" viewBox="0 0 900 420" role="img" aria-label="${scene.title}">
        <defs><linearGradient id="sky" x1="0" x2="0" y1="0" y2="1"><stop offset="0%" stop-color="#0b1120" /><stop offset="100%" stop-color="#111827" /></linearGradient></defs>
        <rect width="900" height="420" rx="24" fill="url(#sky)" />
        <circle cx="748" cy="82" r="46" fill="#fde68a" opacity="0.9" />
        <g opacity="0.95"><ellipse cx="320" cy="116" rx="120" ry="48" fill="#64748b" /><ellipse cx="420" cy="100" rx="145" ry="58" fill="#475569" /><ellipse cx="520" cy="122" rx="110" ry="44" fill="#64748b" /></g>
        <g fill="${accent}">${drops}</g>
        <path d="M 90 300 C 230 240, 330 250, 460 296 S 680 370, 860 298" fill="none" stroke="#0f172a" stroke-width="36" stroke-linecap="round" />
        <path d="M 90 300 C 230 240, 330 250, 460 296 S 680 370, 860 298" fill="none" stroke="#334155" stroke-width="4" stroke-linecap="round" stroke-dasharray="10 10" opacity="0.85" />
        <g transform="translate(118 230)"><circle cx="18" cy="18" r="18" fill="#fde68a" /><path d="M18 36 L18 78 M0 52 L36 52 M6 72 L30 72" stroke="#fde68a" stroke-width="7" stroke-linecap="round" /><path d="M18 78 L6 112 M18 78 L30 112" stroke="#fde68a" stroke-width="7" stroke-linecap="round" /></g>
        <g transform="translate(${x} ${y})"><circle cx="0" cy="0" r="28" fill="#fb7185" /><circle cx="0" cy="0" r="18" fill="#fecdd3" opacity="0.35" /><text x="0" y="6" fill="#fff" font-size="14" text-anchor="middle" font-weight="800">${focus.value_display}</text></g>
        <rect x="540" y="228" width="300" height="112" rx="18" fill="rgba(2,6,23,0.75)" stroke="rgba(148,163,184,0.2)" />
        <text x="562" y="258" fill="#94a3b8" font-size="14" letter-spacing="2">FORECAST</text>
        <text x="562" y="290" fill="#e5eefb" font-size="28" font-weight="800">${caption}</text>
        <text x="562" y="320" fill="#94a3b8" font-size="16">${note}</text>
        <rect x="88" y="44" width="260" height="68" rx="16" fill="rgba(15,23,42,0.9)" stroke="rgba(34,211,238,0.25)" />
        <text x="108" y="73" fill="#22d3ee" font-size="16" font-weight="800">${scene.title}</text>
        <text x="108" y="96" fill="#94a3b8" font-size="12">${period.label} · ${overview?.connected ? 'live Graphite' : 'source unavailable'}</text>
        <rect x="88" y="340" width="370" height="48" rx="14" fill="rgba(15,23,42,0.88)" stroke="rgba(148,163,184,0.18)" />
        <text x="108" y="370" fill="#e5eefb" font-size="14" font-weight="700">${support.label}: ${support.value_display} — ${support.delta_display}</text>
      </svg>
    `;
  }

  if (scene.art === 'rope') {
    return `
      <svg class="scene-art" viewBox="0 0 900 420" role="img" aria-label="${scene.title}">
        <defs><linearGradient id="ropeSky" x1="0" x2="0" y1="0" y2="1"><stop offset="0%" stop-color="#08111f" /><stop offset="100%" stop-color="#0f172a" /></linearGradient></defs>
        <rect width="900" height="420" rx="24" fill="url(#ropeSky)" />
        <circle cx="734" cy="78" r="42" fill="#fde68a" opacity="0.88" />
        <path d="M80 312 C 260 282, 410 284, 820 232" fill="none" stroke="#6b7280" stroke-width="4" stroke-dasharray="8 10" opacity="0.5" />
        <path d="M74 312 C 250 276, 420 274, 824 230" fill="none" stroke="#e2e8f0" stroke-width="8" stroke-linecap="round" opacity="0.9" />
        <g transform="translate(${x} ${y})"><circle cx="0" cy="0" r="18" fill="#f8fafc" /><path d="M0 18 L-10 54 M0 18 L12 54 M-12 34 L12 34" stroke="#f8fafc" stroke-width="5" stroke-linecap="round" /><rect x="-24" y="54" width="48" height="24" rx="12" fill="#94a3b8" opacity="0.9" /></g>
        <path d="M86 348 C 190 254, 314 318, 460 244 S 660 176, 826 220" fill="none" stroke="rgba(244,114,182,0.75)" stroke-width="4" stroke-linecap="round" />
        <rect x="104" y="48" width="322" height="72" rx="16" fill="rgba(15,23,42,0.92)" stroke="rgba(244,114,182,0.30)" />
        <text x="126" y="78" fill="#f9a8d4" font-size="16" font-weight="800">${scene.title}</text>
        <text x="126" y="102" fill="#94a3b8" font-size="12">${period.label} · request balance</text>
        <rect x="536" y="70" width="276" height="96" rx="18" fill="rgba(2,6,23,0.8)" stroke="rgba(148,163,184,0.18)" />
        <text x="558" y="102" fill="#94a3b8" font-size="14" letter-spacing="2">TENSION</text>
        <text x="558" y="134" fill="#e5eefb" font-size="26" font-weight="800">${focus.label}: ${focus.value_display}</text>
        <text x="558" y="160" fill="#94a3b8" font-size="15">${scene.joke}</text>
        <rect x="112" y="258" width="180" height="78" rx="16" fill="rgba(15,23,42,0.88)" stroke="rgba(34,211,238,0.20)" />
        <text x="130" y="290" fill="#22d3ee" font-size="15" font-weight="800">${support.label}</text>
        <text x="130" y="314" fill="#94a3b8" font-size="12">${support.value_display} · ${support.delta_display}</text>
      </svg>
    `;
  }

  return `
    <svg class="scene-art" viewBox="0 0 900 420" role="img" aria-label="${scene.title}">
      <defs>
        <linearGradient id="hillSky" x1="0" x2="0" y1="0" y2="1"><stop offset="0%" stop-color="#09111f" /><stop offset="100%" stop-color="#0f172a" /></linearGradient>
        <linearGradient id="hillBody" x1="0" x2="1" y1="0" y2="0"><stop offset="0%" stop-color="#1f2937" /><stop offset="100%" stop-color="#0f172a" /></linearGradient>
      </defs>
      <rect width="900" height="420" rx="24" fill="url(#hillSky)" />
      <circle cx="748" cy="82" r="44" fill="#fde68a" opacity="0.9" />
      <path d="M82 346 C 178 264, 276 188, 378 164 S 612 126, 828 74" fill="none" stroke="#334155" stroke-width="86" stroke-linecap="round" opacity="0.95" />
      <path d="M82 346 C 178 264, 276 188, 378 164 S 612 126, 828 74" fill="none" stroke="url(#hillBody)" stroke-width="70" stroke-linecap="round" opacity="0.95" />
      <path d="M82 346 C 178 264, 276 188, 378 164 S 612 126, 828 74" fill="none" stroke="#475569" stroke-width="4" stroke-linecap="round" stroke-dasharray="9 11" opacity="0.9" />
      <g transform="translate(118 282)"><circle cx="18" cy="18" r="18" fill="#f8fafc" /><path d="M18 36 L18 76 M0 52 L36 52 M6 72 L30 72" stroke="#f8fafc" stroke-width="6" stroke-linecap="round" /><path d="M18 76 L8 112 M18 76 L30 112" stroke="#f8fafc" stroke-width="6" stroke-linecap="round" /></g>
      <g transform="translate(${x} ${y})"><circle cx="0" cy="0" r="32" fill="#fb7185" /><circle cx="0" cy="0" r="20" fill="#fecdd3" opacity="0.24" /><text x="0" y="6" fill="#fff" font-size="14" text-anchor="middle" font-weight="800">${focus.value_display}</text></g>
      <rect x="544" y="212" width="284" height="108" rx="18" fill="rgba(2,6,23,0.84)" stroke="rgba(148,163,184,0.18)" />
      <text x="566" y="244" fill="#94a3b8" font-size="14" letter-spacing="2">UPHILL METRIC</text>
      <text x="566" y="276" fill="#e5eefb" font-size="28" font-weight="800">${caption}</text>
      <text x="566" y="304" fill="#94a3b8" font-size="16">${note}</text>
      <rect x="96" y="52" width="280" height="74" rx="16" fill="rgba(15,23,42,0.92)" stroke="rgba(34,211,238,0.28)" />
      <text x="116" y="84" fill="#22d3ee" font-size="16" font-weight="800">${scene.title}</text>
      <text x="116" y="108" fill="#94a3b8" font-size="12">${period.label} · ${overview?.connected ? 'live Graphite' : 'source unavailable'}</text>
      <rect x="82" y="340" width="380" height="42" rx="14" fill="rgba(15,23,42,0.88)" stroke="rgba(148,163,184,0.18)" />
      <text x="104" y="367" fill="#e5eefb" font-size="14" font-weight="700">${support.label}: ${support.value_display} — ${support.delta_display}</text>
      <path d="${trail}" fill="none" stroke="${accent}" stroke-width="3" opacity="0.38" />
    </svg>
  `;
}

function renderInspector(scene: SceneTemplate, cards: GraphiteCard[]): void {
  const selected = getSelectedCard(scene, cards);
  byId<HTMLSpanElement>('inspector-chip').textContent = selected.live ? 'live layer' : 'derived layer';
  byId<HTMLHeadingElement>('inspector-title').textContent = selected.label;
  byId<HTMLDivElement>('inspector-subtitle').textContent = scene.title;
  byId<HTMLSpanElement>('inspector-tag').textContent = selected.status;
  byId<HTMLDivElement>('inspector-metric').textContent = selected.metric;
  byId<HTMLDivElement>('inspector-value').textContent = `${selected.value_display} · ${selected.delta_display}`;
  byId<HTMLDivElement>('inspector-progress').textContent = `${selected.progress}% of canvas cap`;
  byId<HTMLDivElement>('inspector-hook').textContent = scene.hook;

  const stack: [string, string][] = [
    ['Scene', scene.title],
    ['Selected layer', selected.label],
    ['Period', selectedPeriodOption().label],
    ['Source', state.overview?.source_label || 'pending'],
    ['Samples', String(selected.points?.length || 0)],
  ];
  byId<HTMLDivElement>('layer-stack').innerHTML = stack.map(([label, value]) => `<div class="stack-item"><strong>${label}</strong><span>${value}</span></div>`).join('');

  byId<HTMLDivElement>('source-url').textContent = state.overview?.source_url || '—';
  byId<HTMLDivElement>('source-window').textContent = state.overview?.window || '—';
  byId<HTMLDivElement>('source-time').textContent = state.overview ? new Date(state.overview.captured_at).toLocaleString() : '—';
}

function renderCanvas(scene: SceneTemplate, cards: GraphiteCard[]): void {
  byId<HTMLDivElement>('scene-badge').textContent = scene.type;
  byId<HTMLHeadingElement>('scene-title').textContent = scene.title;
  byId<HTMLParagraphElement>('scene-summary').textContent = scene.summary;
  byId<HTMLDivElement>('scene-note-big').textContent = scene.noteBig;
  byId<HTMLDivElement>('scene-note-small').textContent = scene.noteSmall;
  byId<HTMLDivElement>('canvas-chip').textContent = `${scene.art} · ${selectedPeriodOption().label}`;
  byId<HTMLDivElement>('scene-art').innerHTML = renderSceneArt(scene, cards, state.overview);
  renderBeatList(scene);
}

function renderOverviewChips(): void {
  const current = state.overview;
  byId<HTMLDivElement>('source-pill').textContent = current ? `${current.source_label} · ${selectedPeriodOption().label}` : 'Connecting to Graphite…';
  byId<HTMLSpanElement>('scene-count-chip').textContent = String(scenes.length);
  byId<HTMLDivElement>('rail-source').textContent = current?.source_label || 'Graphite';
  byId<HTMLDivElement>('rail-source-meta').textContent = current ? `${current.connected ? 'Connected' : 'Fallback'} · ${current.window} · ${current.source_url}` : 'Live source pending…';
  byId<HTMLDivElement>('rail-period').textContent = selectedPeriodOption().label;
  byId<HTMLDivElement>('rail-period-meta').textContent = current ? 'Switch the period to change the shape of the joke and the trend.' : 'Choose a period to load the infographic.';
  byId<HTMLDivElement>('rail-notes').textContent = current?.notes?.[0] || 'The goal is levity and legibility, not a clone of Grafana.';
  byId<HTMLDivElement>('footer-left').textContent = state.overview ? `${state.overview.period_label} · Infograf scene editor` : 'Infograf scene editor';
  byId<HTMLDivElement>('footer-right').textContent = state.overview ? `${state.overview.cards.length} live metrics` : 'Live Graphite infographics with period controls';
}

function render(): void {
  const scene = selectedScene();
  const cards = state.overview?.cards || [];
  renderPeriodButtons();
  renderSceneList();
  renderOverviewChips();
  if (!cards.length) {
    byId<HTMLDivElement>('metric-grid').innerHTML = '<div class="metric-card neutral"><div class="metric-label">Loading</div><div class="metric-value">…</div><div class="metric-detail">Waiting for Graphite data.</div><div class="meter"><span style="width:25%"></span></div></div>';
    byId<HTMLDivElement>('scene-art').innerHTML = '<div class="scene-art" style="display:grid;place-items:center;color:var(--muted);min-height:340px">Loading infographic…</div>';
    renderInspector(scene, []);
    return;
  }
  renderCanvas(scene, cards);
  renderMetricGrid(scene, cards);
  renderInspector(scene, cards);
}

async function loadOverview(): Promise<void> {
  state.loading = true;
  byId<HTMLDivElement>('source-pill').textContent = 'Loading infographic…';
  try {
    const response = await fetch(`/api/v1/sources/graphite?period=${encodeURIComponent(state.period)}`);
    state.overview = await response.json() as GraphiteOverview;
    if (!state.selectedMetric) {
      state.selectedMetric = selectedScene().focusMetric;
    }
    if (state.overview?.cards?.length) {
      if (!state.overview.cards.find((card) => card.metric === state.selectedMetric)) {
        state.selectedMetric = selectedScene().focusMetric;
      }
    }
    render();
  } catch (error) {
    state.overview = {
      connected: false,
      source_url: 'n/a',
      source_label: 'source unavailable',
      period: state.period,
      period_label: periodLabels[state.period],
      window: state.period,
      captured_at: new Date().toISOString(),
      cards: [],
      notes: [String(error)],
      periods: (Object.entries(periodLabels) as Array<[PeriodKey, string]>).map(([key, label]) => ({ key, label, window: key })),
    };
    render();
  }
  state.loading = false;
}

function exportSvg(): void {
  const svg = byId<HTMLDivElement>('scene-art').querySelector('svg');
  if (!svg) return;
  const blob = new Blob([svg.outerHTML], { type: 'image/svg+xml;charset=utf-8' });
  const url = URL.createObjectURL(blob);
  const anchor = document.createElement('a');
  anchor.href = url;
  anchor.download = `infograf-${state.sceneId}-${state.period}.svg`;
  document.body.appendChild(anchor);
  anchor.click();
  anchor.remove();
  window.setTimeout(() => URL.revokeObjectURL(url), 2000);
}

function copyJson(): void {
  const payload = {
    scene: selectedScene(),
    period: state.period,
    overview: state.overview,
  };
  void navigator.clipboard?.writeText(JSON.stringify(payload, null, 2));
}

function surpriseMe(): void {
  state.surpriseIndex = (state.surpriseIndex + 1) % scenes.length;
  state.sceneId = scenes[state.surpriseIndex].id;
  state.selectedMetric = scenes[state.surpriseIndex].focusMetric;
  localStorage.setItem(storageKey('scene'), state.sceneId);
  localStorage.setItem(storageKey('metric'), state.selectedMetric);
  render();
}

byId<HTMLButtonElement>('refresh-btn').addEventListener('click', () => { void loadOverview(); });
byId<HTMLButtonElement>('export-svg-btn').addEventListener('click', exportSvg);
byId<HTMLButtonElement>('copy-json-btn').addEventListener('click', copyJson);
byId<HTMLButtonElement>('scene-joke-btn').addEventListener('click', surpriseMe);

if (!state.selectedMetric) {
  state.selectedMetric = selectedScene().focusMetric;
}

void loadOverview();
