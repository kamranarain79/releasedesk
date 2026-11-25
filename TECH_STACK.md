# ReleaseDesk Technology Stack

## Frontend

- **Lit** — web components for interactive console pieces (login, drawers, upload forms)
- **Tailwind CSS** — utility styles; scanned from `web/` and `*.gohtml`
- **esbuild** — bundles `web/src/entrypoint.ts` → `cmd/releasedesk-server/assets/bundle.js`
- **Go `html/template`** — server-rendered pages (`internal/views/templates/console/`)

## Backend

- **Chi** — HTTP routing; console on `:8080`, portal on `:9090`
- **XORM** — SQLite ORM; migrations in `cmd/releasedesk-server/migrations.go`
- **SQLite** — `_data/releasedesk.db` (WAL)
- **uber/dig** — dependency injection
- **go-chi/jwtauth** — session JWT in HttpOnly cookie

## Runtime layout

| Port | Role |
|------|------|
| 8080 | Admin console (login, dashboard, releases, builds) |
| 9090 | Download portal (public app list; branded downloads later) |

Static files under `/assets/*` are served from **`go:embed`** in the server binary.

## DevOps & Infrastructure

- Docker & Docker Compose (`Dockerfile`, `docker-compose.yml`; Go 1.23 + Node 22)
- GitHub Actions (CI/CD)
- Local: [Air](https://github.com/air-verse/air) optional (`.air.toml`, dev proxy :8090)

## Languages

- Go
- TypeScript (web/)
- Shell/Bash scripts

## Documentation

- [docs/DEVELOPMENT.md](docs/DEVELOPMENT.md) — build, run, troubleshooting
- [CONTRIBUTING.md](CONTRIBUTING.md) — contributor workflow
- `.cursor/rules/` — Cursor agent rules for this repo