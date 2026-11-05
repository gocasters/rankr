# Web scaffold

Part of issue #218. Currently provides an HTTP server, health check,
configuration loading, and graceful shutdown.

An opt-in leaderboard preview renders sample, empty, and error states.
An opt-in leaderboard page fetches data from the leaderboardstat HTTP API.
Shared feature flag integration and htmx remain follow-up work.

## Run locally

From the repository root:

```bash
go run ./cmd/web serve
```

Health endpoint:

http://localhost:8085/v1/health-check

Expected response:

```json
{ "status": "ok" }
```

Press Ctrl+C to stop the server.

## Leaderboard preview

Run from the repository root:

```bash
go run ./cmd/web serve --preview
```

Open http://localhost:8085/preview.

Use the page links to inspect the sample table, empty state, and error state.
The preview uses sample data and is disabled by default.

Templates are embedded in the binary. Restart the web process after editing them.

## API-backed leaderboard

Start leaderboardstat and its dependencies, then run:

```bash
go run ./cmd/web serve --leaderboard
```

Open http://localhost:8085/leaderboard?project_id=1028435569.

`project_id` is the numeric GitHub repository ID.
The page displays up to 10 rows, or an empty or error state.
The page is disabled by default.

Configure `leaderboard_api.base_url` and `leaderboard_api.timeout`
in YAML. Environment overrides are also supported:

```bash
WEB__LEADERBOARD_API__BASE_URL=http://127.0.0.1:8081 go run ./cmd/web serve --leaderboard
```

## Configuration

The default file is deploy/web/development/config.yml.
Set CONFIG_PATH to use another YAML file.

Environment variables override YAML values. For example, in Git Bash:

```bash
WEB__HTTP_SERVER__PORT=8086 go run ./cmd/web serve
```

The health endpoint then uses port 8086.

## Checks

```bash
make check
```

Without Make, in Git Bash:

```bash
CGO_ENABLED=1 bash scripts/check.sh
```

Race tests require a working C compiler on PATH.
