# Local development

This document reflects **this repository** (v1.0.0 early console). The hosted product docs at [releasedesk.dev/docs](https://releasedesk.dev/docs/) describe additional features (full portal, API v1, org settings) that are not all implemented here yet.

## Prerequisites

| Tool | Version / notes |
|------|------------------|
| Go | 1.23+ (`go.mod`) |
| Node.js | 22.x for `web/` (esbuild + Tailwind) |
| Docker | Optional; **preferred** so Go, Node, and embedded UI stay on one version |
| Optional | [Air](https://github.com/air-verse/air) for Go live reload (`.air.toml`) |

On hosts without matching Go/Node versions, use Docker instead of mixing toolchains.

## Docker (recommended)

The image compiles the Lit/Tailwind bundle, then embeds it into the Go binary. That avoids a stale `bundle.js` in a binary built on a different host.

```bash
docker compose up --build
```

| URL | Purpose |
|-----|---------|
| http://localhost:8080 | Console |
| http://localhost:9090 | Public download portal |

`docker compose up --build` starts ReleaseDesk and PostgreSQL together. Metadata goes to Postgres. Firmware files stay on the `releasedesk-data` volume. See [INFRASTRUCTURE.md](INFRASTRUCTURE.md).

Environment:

| Variable | Default | Purpose |
|----------|---------|---------|
| `RELEASEDESK_DATA_DIR` | `_data` | Artifact files (firmware, branding). Metadata is not stored here |
| `RELEASEDESK_DATABASE_URL` | `postgres://releasedesk:releasedesk@127.0.0.1:5432/releasedesk?sslmode=disable` | PostgreSQL connection. Compose sets this to the `postgres` service |
| `RELEASEDESK_PORTAL_URL` | `http://localhost:9090` | URL shown in the console Download Portal page |
| `RELEASEDESK_API_TOKEN` | unset | Optional extra global token. The instance token is generated on first start and shown under Administration |
| `RELEASEDESK_JWT_SECRET` | `secret` outside production | Required non-default value when `RELEASEDESK_ENV=production` |
| `RELEASEDESK_COOKIE_SECURE` | unset | Set to `1` behind HTTPS |
| `RELEASEDESK_MAX_UPLOAD_MB` | `512` | Multipart upload cap |

## First-time setup (host toolchain)

PostgreSQL must be listening before the binary starts. `make run` starts the Compose Postgres service, then the server.

```bash
make build
make run
```

Or by hand:

```bash
cd web
npm install
npm run build

cd ..
docker compose up -d postgres
go build -o ./_tmp/releasedesk-server ./cmd/releasedesk-server
mkdir -p _data
./_tmp/releasedesk-server
```

## URLs and auth

| URL | Purpose |
|-----|---------|
| http://localhost:8080 | **Console** — admin UI |
| http://localhost:8080/homepage | Applications |
| http://localhost:8080/portal | Download Portal (console) |
| http://localhost:8080/people | People |
| http://localhost:8080/administration | Administration |
| http://localhost:8080/account/settings | Account settings |
| http://localhost:9090 | **Portal** — public app list |
| http://localhost:8080/login | Sign-in |
| http://localhost:8080/health | Health check |

Seeded console user: **admin** / **admin**. The first login is required to change that password; do not use it on a public host.

API reference: [openapi.yaml](openapi.yaml), also served at `/api/docs` after sign-in.

Unauthenticated access to `/` redirects: `/` → `/homepage` → `/login`.

Use **HTTP**, not HTTPS, on 8080/9090 unless you terminate TLS elsewhere.

## Frontend build and `go:embed`

Console interactivity uses Lit components (`<login-form>`, drawers, etc.) in `web/src/`. Built assets land in:

- `cmd/releasedesk-server/assets/bundle.js`
- `cmd/releasedesk-server/assets/tailwind.css`

Static files (logo, Font Awesome) also live under `cmd/releasedesk-server/assets/` and are embedded with `//go:embed assets`.

They are embedded at **Go compile time**. You must run **`npm run build` then `go build`** and **restart** the server after UI changes. Docker does this in one image.

**Symptom:** Login page shows heading “Sign in to your account” but an empty `<login-form>` — check that http://localhost:8080/assets/bundle.js returns **200**.

### UI watch mode

```bash
cd web && npm run dev   # watches esbuild + Tailwind into assets/
```

Still re-run `go build` and restart the server when you want those assets inside the binary (Air does not rebuild embed automatically unless you add a hook).

### Air (optional)

From repo root:

```bash
air
```

Browser proxy: **http://localhost:8090** → app on **8080** (see `.air.toml`).

## Architecture (short)

```
cmd/releasedesk-server/     main, migrations, Chi routers, embed FS
internal/domains/           app, build, platform, release, user
internal/services/          business logic
internal/controllers/       HTTP handlers + view data
internal/persistence/sql/   XORM repositories
internal/views/             gohtml templates (embedded)
web/                        Lit + TypeScript + Tailwind
_data/                      artifact files (runtime, gitignored)
```

Two HTTP servers in one process: console `:8080`, portal `:9090`.

## Implemented console features

| Area | Path |
|------|------|
| App list | `/homepage` |
| Download Portal | `/portal` (console) and `:9090` (public list) |
| People | `/people` (lists users; inviting is later) |
| Administration | `/administration` (placeholder for later org settings) |
| Account settings | `/account/settings` (read-only profile) |
| Overview / setup guide | `/dashboard/{slug}` |
| Releases | `/dashboard/{slug}/releases`, version summary, release notes, QA tests |
| Builds | `/dashboard/{slug}/platforms/{OS}/builds` — Android/iOS/desktop plus **ESP32**, **STM32**, Linux **RAUC** |
| Upload (API) | `POST /api/apps/{slug}/platforms/{platform}/builds` (API token) |
| Python / Filebrowser API | `GET /api/v1/apps/...` — files, notes, test methods (see [PYTHON_TOOL.md](PYTHON_TOOL.md)) |

## Later-version placeholders

These routes render a page (not 404) until implemented: inviting people, org administration, branded portal downloads, distribution-groups, announcements, feeds, analytics, settings, platform configurations.

Releases with no channels (setup guide not finished) show a setup prompt instead of crashing.

## Troubleshooting

| Issue | Fix |
|-------|-----|
| Port 8080 in use | Stop other processes (`ss -tlnp \| grep 8080`); only one `releasedesk-server` or Compose stack at a time |
| Go/Node version mismatch | Use `docker compose up --build` |
| 404 on navbar/sidebar links (old builds) | Rebuild so routes in `console.go` are in the binary |
| Releases page errors | Complete app **setup guide** on overview (channels/platforms) |
| Empty login form | Rebuild web assets, then the Go binary, then restart |
| Docs vs local behavior | This repo is early stage; README roadmap vs [releasedesk.dev](https://releasedesk.dev/docs/) |

## Related

- [TECH_STACK.md](../TECH_STACK.md) — stack summary
- [CONTRIBUTING.md](../CONTRIBUTING.md) — contributor workflow
- [PYTHON_TOOL.md](PYTHON_TOOL.md) — Filebrowser-style API for the production Python tool
- `.cursor/rules/` — Cursor agent context for this repo
