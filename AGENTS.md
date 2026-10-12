# AGENTS.md — rankr

Read this before changing anything. It applies to every coding agent (Claude Code, Codex, Gemini CLI, Cursor, Copilot, Aider, …) and to humans.

## What rankr is

A contribution leaderboard for GitHub repos. GitHub webhooks are scored, and a public leaderboard and profile are shown per contributor. Go 1.25+, Echo, gRPC (buf), Postgres (one DB per service, `sql-migrate`), Redis, NATS JetStream (Watermill).

## Map

- `<name>app/` is one service, laid out as `delivery/{http,grpc,consumer,scheduler}`, `service/`, `repository/` and `adapter/`. Its binary is `cmd/<name>`.
- v1 services: `webhookapp` (ingest), `leaderboardscoringapp` (scoring, NATS consumer), `leaderboardstatapp` (public leaderboard and stats), `projectapp`, `contributorapp`, `authapp`, `userprofileapp`. Plus `cmd/web`, the HTML UI (html/template + htmx), once it lands.
- Event path: webhook → NATS `rankr_raw_events` → leaderboardscoring (Redis ZSET) → leaderboardstat cron (gRPC) → `GET /v1/leaderboard/public/:project_id`.
- `protobuf/` holds the .proto sources. `protobuf/golang/` is generated: never edit it by hand; run `make proto-gen`.
- **Frozen:** `realtimeapp/`, `notifapp/`, `hub/`, `argus`, `cmd/api`, Temporal and `deploy/development/*-compose.yml`. Do not change them.

## Commands

- `make check` is **the** gate. It runs gofmt, vet, `go test -race -short` and build on the packages your branch changes. It must be green before you open a PR, and CI runs it too.
- `make check-all` runs the same checks over the whole repo. It is red today because of a known baseline; see the issue labelled `baseline`.
- `make infra-up`, then `make start-<svc>-app-dev`: run locally.
- `make proto-gen` and `make proto-lint`: protobuf.

## Hard rules

1. Work from an issue. Its acceptance criteria are the spec. No issue, no PR.
2. Tests come first for behaviour changes. Never change an expected value just to make a test pass.
3. One issue per PR, at most 400 changed lines (generated code excluded), a branch that lives at most one day, and squash merge.
4. No new top-level directories, no moved packages, no new services. The only exception is `cmd/web`.
5. Proto and migrations are additive only: add proto fields with new numbers, and use expand/contract for migrations. Never rename or drop in the same PR.
6. A user-visible feature that isn't finished goes behind a feature flag, off by default.
7. Never commit secrets, `.env` files or real tokens, including in docs and examples.
8. Add no new dependency without an issue that says why.
9. No `math/rand` outside tests, and no HTTP client without a timeout.
10. Agents never merge. A human approves and merges, and the author must be able to explain every line in review.

## Roles and workflow

Four small roles live in `.claude/agents/`: `product`, `tester`, `developer` and `reviewer`. Claude Code loads them as subagents. With any other tool, tell it: "act as the role defined in `.claude/agents/<role>.md`". Who acts when is described in [`docs/agentic/WORKFLOW.md`](docs/agentic/WORKFLOW.md), which is one page.

## PRs

Use the PR template. The **Agent disclosure** section is required: which tool you used, what it did, what you verified yourself, and what you are unsure about.
