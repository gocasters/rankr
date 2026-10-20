# Rankr architecture

## Purpose and scope

This page is an onboarding map of Rankr's running Stage v1 architecture. It
describes the six services in the checked-in Stage stack and separates them
from service code and components that Stage does not run. Runtime connections
below are based on application wiring, not on configuration fields alone.

## Stage topology

GitHub webhooks and public API clients enter through host nginx, which
terminates TLS and forwards traffic to the Stage gateway at `127.0.0.1:8088`.
The gateway listens on port 80 inside the private `rankr-stage` network and
routes requests to service HTTP endpoints. Postgres, Redis, NATS, service HTTP
ports, and gRPC ports are not published outside the host.

Postgres is one Stage server with separate databases and users for services.
Redis and NATS JetStream are shared infrastructure; each service connects only
to the stores shown in the matrix below.

## Service map

| Stage service | Responsibility |
|---|---|
| `webhook` | Receives GitHub activity and publishes contribution events. |
| `leaderboardscoring` | Consumes raw events, calculates scores, updates live leaderboards, and emits processed score events. |
| `leaderboardstat` | Serves leaderboard/stat endpoints and refreshes public leaderboard data. |
| `project` | Stores project/repository data and serves project gRPC calls. |
| `contributor` | Stores contributor data and verifies credentials for auth. |
| `auth` | Issues and refreshes JWTs and validates tokens and permissions. |

The repository also contains `userprofileapp/` and `taskapp/`; they are not in
the Stage compose stack. `realtimeapp/` and `notifapp/` are also outside Stage
and are frozen. `hub/`, `argus`, `cmd/api`, and Temporal are frozen components,
not services in the Stage v1 stack. `deploy/stage/docker-compose.yml` is the
source for which services actually run in Stage.

## Ports

These are configured service ports. The runtime gRPC column lists only
listeners started by the corresponding app; some configs contain unused RPC
port fields.

| Service | HTTP | Runtime gRPC | Stage status |
|---|---:|---:|---|
| Auth | 8081 | — | Stage |
| Contributor | 8082 | 8093 | Stage |
| Leaderboard scoring | 8081 | 8070 | Stage |
| Leaderboard stats | 8081 | 8091 | Stage |
| Project | 8084 | 8094 | Stage |
| Webhook | 8090 | — | Stage |
| User profile | 8081 | — | Not in Stage |
| Task | 8084 | — | Not in Stage |
| Notification | 8081 | — | Not in Stage; frozen |
| Realtime | 8084 | — | Not in Stage; frozen |
| Argus (auxiliary) | — | — | Not in Stage; frozen; development config sets server port 5015 (protocol unspecified) |

Stage infrastructure and ingress ports:

| Endpoint | Port | Exposure |
|---|---:|---|
| Host nginx to gateway | `127.0.0.1:8088` | Host-local; host nginx terminates TLS. |
| Gateway container | `:80` | Private Stage network. |
| PostgreSQL | `5432` | Private Stage network. |
| Redis | `6379` | Private Stage network. |
| NATS client | `4222` | Private Stage network. |
| NATS monitoring/health | `8222` | Internal; not published. |

Service values come from `deploy/<service>/production/config.yml`; running
listeners are confirmed by the service app startup code. Stage exposure is
defined by `deploy/stage/docker-compose.yml` and `deploy/stage/gateway/`.

## Datastore connection matrix

The Stage stack provides shared Postgres, Redis, and NATS JetStream. Shared
means a common infrastructure server, not that every service uses every
store. This matrix describes connections confirmed in the Stage service
application wiring.

| Stage service | Postgres | Redis | NATS JetStream |
|---|---|---|---|
| Auth | Yes | — | — |
| Contributor | Yes | Yes | — |
| Leaderboard scoring | Yes | Yes | Yes |
| Leaderboard stats | Yes | Yes | — |
| Project | Yes | — | — |
| Webhook | Yes | Yes | Yes |

Postgres databases are service-specific. Config entries without corresponding
runtime wiring are not shown as connections.

## Event flow

1. GitHub sends a delivery to the webhook HTTP handler.
2. Webhook validates required delivery headers, handles supported event types,
   persists event data, and publishes raw events to NATS JetStream on
   `rankr_raw_events`.
3. Leaderboard scoring consumes raw events, calculates scores, and updates
   Redis leaderboard sorted sets.
4. Scoring publishes processed score events to NATS for batch persistence and
   persists processed events/snapshots to Postgres.
5. Leaderboard stats calls scoring over gRPC for leaderboard data, uses Redis
   for public leaderboard reads/cache, and calls project over gRPC for project
   details. The public route `GET /v1/leaderboard/public/:project_id` currently
   returns HTTP 401; see the known limitation in `docs/stage.md`.

The stats service also has scheduled refresh work. The flow above reflects
service responsibilities; it does not imply that stats consumes the raw event
stream.

## Auth flow

1. A client sends username/password credentials to the public auth login route
   through the gateway.
2. Auth verifies credentials through contributor's gRPC service and issues
   access and refresh JWTs.
3. For protected routes, the gateway makes an internal `auth_request` to auth's
   `/v1/me`, passing the bearer token and original request details.
4. Auth validates the JWT and required permission. The gateway forwards the
   request to its service with identity/access headers.

The gateway exposes auth login/refresh routes and rewrites `/auth/*` to auth's
internal routes. The inspected flow is username/password plus JWT; it does not
implement GitHub OAuth.

## Frozen areas

Do not change the following areas, as listed in `AGENTS.md`:

- `realtimeapp/`, `notifapp/`, `hub/`, `argus`, and `cmd/api`
- Temporal
- `deploy/development/*-compose.yml`

## Sources of truth

- Stage service list, network, and infrastructure: `deploy/stage/docker-compose.yml`.
- Stage gateway routes and upstreams: `deploy/stage/gateway/nginx.conf` and
  `deploy/stage/gateway/auth_headers.conf`.
- Stage topology and ingress explanation: `docs/stage.md`.
- Configured service ports: each service's `deploy/<service>/production/config.yml`.
- Active listeners and datastore wiring: service `app.go` startup/setup and
  `cmd/<service>/command/serve.go`.
- Event topic names: `pkg/topicsname/topicname.go`; event handling/publishing:
  `webhookapp/` and `leaderboardscoringapp/`.
- Auth behavior: `authapp/delivery/http/`, `authapp/app.go`, and Stage gateway
  configuration.
- Frozen areas: `AGENTS.md`.

## Architecture diagram

```mermaid
flowchart LR
  GH[GitHub] -->|webhook delivery| HOST[Host nginx<br/>TLS termination]
  CLIENT[Client] -->|HTTPS| HOST[Host nginx<br/>TLS termination]
  HOST -->|127.0.0.1:8088| GW[Stage gateway<br/>container :80]

  subgraph Stage["Stage private network"]
    WH -->|raw event: rankr_raw_events| NATS[(NATS JetStream<br/>4222)]
    NATS --> LS[Leaderboard scoring<br/>HTTP 8081 · gRPC 8070]
    LS -->|live leaderboard scores| REDIS[(Redis<br/>6379)]
    LS -->|processed score events| NATS
    LS -->|processed events / snapshots| PG[(Postgres<br/>5432)]
    WH -->|event records| PG

    STAT[Leaderboard stats<br/>HTTP 8081 · gRPC 8091]
    STAT -->|leaderboard gRPC| LS
    STAT -->|project lookup gRPC| PROJECT[Project<br/>HTTP 8084 · gRPC 8094]
    STAT -->|public leaderboard read/cache (route currently returns 401)| REDIS
    STAT --> PG

    AUTH[Auth<br/>HTTP 8081]
    CONTRIB[Contributor<br/>HTTP 8082 · gRPC 8093]
    AUTH -->|credential verification gRPC| CONTRIB
    AUTH --> PG
    CONTRIB --> PG
    CONTRIB --> REDIS

    GW -->|auth_request: /v1/me| AUTH
    GW --> WH
    GW --> STAT
    GW --> PROJECT
    GW --> CONTRIB
  end

  subgraph OutsideStage["Repository service code; not in Stage"]
    USERPROFILE[User profile]
    TASK[Task]
    NOTIF[Notification · frozen]
    REALTIME[Realtime · frozen]
  end

  subgraph OtherFrozen["Other frozen components"]
    HUB[hub]
    ARGUS[argus]
    API[cmd/api]
    TEMPORAL[Temporal]
  end
```
