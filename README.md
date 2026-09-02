# Infograf

Infograf is a greenfield monitoring and visualization mockup that turns telemetry into a story clients can buy.

This repo currently ships a working demo with:

- a SQLite-backed control plane
- a live Graphite-backed demo surface
- a monetization-oriented landing page
- a TypeScript frontend shell that mirrors the same data

## Run locally

```bash
tools/bootstrap-toolchain.sh
PATH="$(pwd)/.tools/bin:$PATH"
go run ./cmd/infograf
```

Open the app at:

- http://127.0.0.1:8080/

## Graphite connection

By default the demo tries to read live data from:

- `http://172.16.0.11:8080`

You can override that with:

```bash
set INFOGRAF_GRAPHITE_URL=http://your-graphite-host:8080
```

If the Graphite host is unavailable, the demo falls back to seeded values so the mockup still renders.

## Environment variables

- `INFOGRAF_ADDR` — HTTP listen address, default `:8080`
- `INFOGRAF_DATA_DIR` — directory for the SQLite DB, default `data`
- `INFOGRAF_DB_PATH` — explicit SQLite file path override
- `INFOGRAF_GRAPHITE_URL` — Graphite web host for live demo data

## Build checks

```bash
tools/bootstrap-toolchain.sh
PATH="$(pwd)/.tools/bin:$PATH"
go test ./...
cd frontend
npm run build
```

## What the demo shows

- live Graphite metrics cards
- workspace/revision status from SQLite
- a pricing mockup for a future monetized product
- a layout that can be trimmed down before source control if desired
