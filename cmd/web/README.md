# Web scaffold

Part of issue #218. Currently provides an HTTP server, health check,
configuration loading, and graceful shutdown.

An opt-in leaderboard preview renders sample, empty, and error states.
An opt-in leaderboard page fetches data from the leaderboardstat HTTP API.

## Run locally

From the repository root:

```bash
WEB__WEB__ENABLED=true go run ./cmd/web serve
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

Start leaderboardstat and its dependencies, then enable the web service and leaderboard:

```bash
WEB__WEB__ENABLED=true WEB__WEB__LEADERBOARD=true go run ./cmd/web serve
```

Open http://localhost:8085/leaderboard?project_id=1028435569.

`project_id` is the numeric GitHub repository ID.
The page displays up to 10 rows, or an empty or error state.
The page is disabled by default.

Configure `leaderboard_api.base_url` and `leaderboard_api.timeout`
in YAML. Environment overrides are also supported:

```bash
WEB__WEB__ENABLED=true WEB__WEB__LEADERBOARD=true \
  WEB__LEADERBOARD_API__BASE_URL=http://127.0.0.1:8081 go run ./cmd/web serve
```

The refresh button updates the leaderboard using htmx.
htmx 2.0.11 is loaded from jsDelivr with an integrity check.

## Configuration

The default file is deploy/web/development/config.yml.
Set CONFIG_PATH to use another YAML file.

Set `web.enabled` to start the HTTP server and `web.leaderboard` to expose
`/leaderboard`. Both default to off. To enable both in YAML, add:

```yaml
web:
  enabled: true
  leaderboard: true
```

For local use, `--preview` and `--leaderboard` also start the server when
`web.enabled` is unset. An explicit `web.enabled: false` disables the server
even with either switch. `--leaderboard=false` disables the API-backed route
when the server is enabled.

Environment variables override YAML values. For example, in Git Bash:

```bash
WEB__WEB__ENABLED=true WEB__HTTP_SERVER__PORT=8086 go run ./cmd/web serve
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
