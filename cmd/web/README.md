# Web scaffold

Part of issue #218. Currently provides an HTTP server, health check,
configuration loading, and graceful shutdown.

An opt-in leaderboard preview renders sample, empty, and error states.
The API-backed page and shared feature flag integration are follow-up work.

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
