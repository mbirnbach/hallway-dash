# Hallway Dashboard

[![Docker Hub](https://img.shields.io/docker/v/mbirnbach/hallway-dash?label=docker%20hub&sort=semver)](https://hub.docker.com/r/mbirnbach/hallway-dash)
[![Build and publish Docker image](https://github.com/mbirnbach/hallway-dash/actions/workflows/docker-publish.yml/badge.svg)](https://github.com/mbirnbach/hallway-dash/actions/workflows/docker-publish.yml)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

Self-hosted, always-on wall-display dashboard for a hallway TV (portrait, 1080×1920). Shows clock, date, weather + 6-day forecast, sunrise/sunset, indoor temp/humidity (Home Assistant), a rolling 4-week calendar (personal + waste/bin collection), a NINA/BBK civil-protection alert banner, a monthly background photo, and a random joke.

Go backend (`net/http`, no framework) proxies every third-party call server-side — the browser only ever talks to this app, so there are no CORS issues and your Home Assistant token never reaches the client. Frontend is plain HTML/CSS/JS, no build step. Every data source degrades gracefully: if something's unset or unreachable, that section just hides itself (or shows `—`) instead of erroring.

## Run it

### Option A: Unraid (Community Applications)

1. Add this repo's template repository to Community Applications: **Apps → Settings → Template repositories** → add `https://github.com/mbirnbach/unraid-docker-templates`.
2. Search Community Applications for **hallway-dash** and install.
3. Fill in the fields you have (weather location, ICS URLs, AGS code, Home Assistant) — leave the rest blank, they degrade gracefully.
4. Set the `backgrounds` path mapping to a folder containing `01.jpg` … `12.jpg`, if you have them.
5. Point a browser/kiosk device at `http://<unraid-ip>:8080/`.

The template is also available directly: [`unraid/hallway-dash.xml`](unraid/hallway-dash.xml).

### Option B: Docker Compose (any Docker host, including Unraid via the terminal)

```bash
git clone https://github.com/mbirnbach/hallway-dash.git
cd hallway-dash
cp .env.example .env   # fill in your real values
mkdir -p backgrounds    # drop in 01.jpg ... 12.jpg (one per month)
docker compose up -d
```

This pulls the published `mbirnbach/hallway-dash:latest` image from Docker Hub. To build from source instead, edit `docker-compose.yml` (comment out `image:`, uncomment `build: .`).

Open `http://<host>:8080` fullscreen on the display (e.g. a Chromium kiosk on a fire-stick style device, or Unraid's browser widget).

## Configuration

All configuration is via environment variables — see [`.env.example`](.env.example) for the full list and defaults. Every data source degrades gracefully when unset or unreachable.

| Variable | Purpose |
|---|---|
| `LATITUDE`, `LONGITUDE`, `LOCATION_LABEL` | Open-Meteo weather location + display name |
| `CALENDAR_ICS_URL` | Personal calendar ICS feed |
| `WASTE_ICS_URL` | Bin/waste-collection ICS feed |
| `AGS_CODE` | 12-digit Amtlicher Gemeindeschlüssel for NINA/BBK alerts |
| `JOKE_API_URL` | Joke source (defaults to `https://nopunintended.xyz`, falls back to JokeAPI.dev) |
| `HA_URL`, `HA_TOKEN`, `HA_TEMP_ENTITY`, `HA_HUMIDITY_ENTITY` | Home Assistant indoor sensor |
| `CLOCK_FORMAT` | `24h` (default) or `12h` |
| `ACCENT_COLOR` | Hex accent used for bin-day pills / today's date / "Heute" |

## Background photos

Mount a `backgrounds/` directory containing `01.jpg` … `12.jpg` (one per month). Missing months fall back to a placeholder photo automatically.

## Local development

```bash
go run .              # serves on :8080, reads ./static and ./backgrounds, loads ./.env if present
```

## Building the image yourself

```bash
docker build -t hallway-dash .
```

The Dockerfile cross-compiles from the build platform to the target platform, so `docker buildx build --platform linux/amd64,linux/arm64` works without QEMU-emulating the Go toolchain itself. CI publishes both architectures to Docker Hub automatically on every push to `main` and on version tags (see [`.github/workflows/docker-publish.yml`](.github/workflows/docker-publish.yml)).

## Project layout

```
main.go                     # wiring: config, background pollers, HTTP routes
internal/config              # env var parsing + .env loading
internal/pollcache           # generic background-refresh cache used by every data source
internal/weather              # Open-Meteo proxy
internal/calendar             # ICS parsing (golang-ical) + RRULE expansion (rrule-go)
internal/alerts                # NINA/BBK proxy
internal/joke                    # joke source + JokeAPI.dev fallback
internal/homeassistant             # indoor sensor proxy
static/                              # frontend (index.html, style.css, app.js)
unraid/hallway-dash.xml              # Community Applications template
```

## License

[MIT](LICENSE)
