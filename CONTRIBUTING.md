# Contributing to ReleaseDesk

Hey there! 👋 Thanks for your interest in contributing to my project. It means a lot!
If you need any help or have questions, don’t hesitate to ask.

To understand the tech stack, please refer to the [Tech Stack](TECH_STACK.md) document.

## Quick Start

1. Build web assets (required before the Go binary embeds them):

   ```bash
   cd web && npm install && npm run build
   ```

2. Build and run the server:

   ```bash
   go build -o ./_tmp/releasedesk-server ./cmd/releasedesk-server
   ./_tmp/releasedesk-server
   ```

3. Open **http://localhost:8080/login** (console). Optional: [Air](https://github.com/air-verse/air) in the repo root proxies **http://localhost:8090** → **8080**.

To avoid Go/Node version drift, run **`docker compose up --build`** instead.

See [docs/DEVELOPMENT.md](docs/DEVELOPMENT.md) for full setup and troubleshooting.

### UI Development

- **Bundler**: Uses `esbuild` for fast build times. Note that it doesn’t validate TypeScript; that’s left to your IDE or
  the TypeScript CLI if needed.
- **CSS**: Tailwind CLI generates the CSS file, pulling classes from the web components and Go template files.

To run the UI development environment with live reload, use:

```bash
npm run dev
```

## Project Architecture

This project is a multi-page application with a hybrid setup. It uses Go’s `html/template` for server-side rendering,
sending
mostly plain HTML over the wire. Web Components step in only where extra interactivity is needed, keeping things simple
and sticking to web standards.

### Dependency Injection

Uses `uber-dig` for managing dependencies. All components (services, repositories, controllers) are added to a container
that resolves the dependency graph.

### Database

SQLite database at `_data/releasedesk.db`. Schema changes are **XORM migrations** registered in
`cmd/releasedesk-server/migrations.go` (`xorm.io/xorm/migrate`), applied automatically on server start.

### HTTP Server

Built using `go-chi` to manage both API and page requests.

### Code

This project follows Domain-Driven Design (DDD) concepts and draws inspiration from both the MVC and Clean Architecture
patterns and composability is highly recommended whenever possible.
The code style used is the default formatting provided by the IntelliJ IDE.

## References

- https://go-chi.io
- https://jmoiron.github.io/sqlx/
- https://lit.dev/docs/
- https://go.uber.org/dig
- https://github.com/golang-migrate/migrate