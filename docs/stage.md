# Stage environment

Stage runs the six v1 services (auth, contributor, leaderboardscoring,
leaderboardstat, project, webhook) with Postgres, Redis and NATS on one host.
Every Friday demo runs from stage.

Everything lives in `deploy/stage/`:

| File | Purpose |
|---|---|
| `docker-compose.yml` | The whole stack on a private `rankr-stage` network |
| `gateway/nginx.conf` | OpenResty gateway: public routes, health checks, `auth_request` for the rest |
| `nats.conf` | NATS with JetStream memory and disk caps |
| `.env.stage.example` | Every variable stage needs, without values |
| `deploy.sh` | Deploy, rollback, status, logs, health |
| `host-nginx.conf.example` | vhost for the host nginx that terminates TLS |

## Layout

```
internet ──443──▶ host nginx (TLS) ──▶ 127.0.0.1:8088 gateway ──▶ services
                                                    (rankr-stage network)
```

- Only the gateway is published, and only on `127.0.0.1`. No service, gRPC,
  Postgres, Redis or NATS port is reachable from outside the host.
- Service configs are the production `config.yml` files baked into the images.
  Stage-specific values (passwords, JWT key, log level, stream size) are
  environment overrides in `docker-compose.yml`, read from `.env.stage`.
- Network aliases (`shared-postgres`, `project-app-prod`, ...) match the host
  names in those configs, so no config file is copied for stage.

## Gateway routes

| Route | Auth | Service |
|---|---|---|
| `/health/<service>` | public | health check of each v1 service |
| `/github-webhook/*` | public | webhook |
| `GET /v1/leaderboard/public/:project_id` | public | leaderboardstat |
| `/auth/*` (login, refresh-token) | public | auth |
| `/v1/projects`, `/v1/vcs-repos` | bearer | project |
| `/v1/profile`, `/create`, `/update`, `/jobs`, `/password` | bearer | contributor |
| `/v1/contributors/*` | bearer | leaderboardstat |

## First-time setup on the host

Requirements: Docker with the compose plugin, git, curl, nginx, certbot.

```bash
git clone https://github.com/gocasters/rankr /opt/rankr-stage
cd /opt/rankr-stage/deploy/stage

cp .env.stage.example .env.stage
chmod 600 .env.stage
# Fill every empty value; generate each with: openssl rand -hex 24
nano .env.stage

./deploy.sh deploy
```

Then put the host nginx in front of the gateway:

```bash
cp host-nginx.conf.example /etc/nginx/sites-available/rankr-stage
# edit server_name if the stage domain differs
ln -s /etc/nginx/sites-available/rankr-stage /etc/nginx/sites-enabled/
nginx -t && systemctl reload nginx
certbot --nginx -d stage.rankr.gocasts.ir
```

## Deploy

```bash
cd /opt/rankr-stage
git pull
deploy/stage/deploy.sh deploy
```

`deploy` builds images tagged with the short commit hash, starts them, and
checks `/health/<service>` for all six services through the gateway. If a
service is not healthy within `HEALTH_TIMEOUT` seconds, it starts the previous
release again and exits non-zero. Each healthy release is appended to
`deploy/stage/.deploy-history`.

`HEALTH_TIMEOUT` is the whole wait, not a per-request limit: every request
(at most 5 seconds each) is cut short at the deadline, and services not
checked by then are reported as `SKIP`. It defaults to 150, must be a positive
whole number of seconds, and is read from the shell environment, not from
`.env.stage`. It applies to `deploy` and `rollback`:

```bash
HEALTH_TIMEOUT=60 deploy/stage/deploy.sh deploy
```

## Health

```bash
deploy/stage/deploy.sh health
curl https://stage.rankr.gocasts.ir/health/leaderboardstat
```

## Logs

```bash
deploy/stage/deploy.sh status
deploy/stage/deploy.sh logs              # all services
deploy/stage/deploy.sh logs webhook      # one service
```

Container logs are capped at 3 × 10 MB per container.

## Roll back

```bash
deploy/stage/deploy.sh rollback
```

This starts the previous release from `.deploy-history` without building.
Migrations run on startup (`serve --migrate-up`) and are not rolled back: if a
release added a migration the previous code cannot handle, fix forward instead.
Old images are needed for rollback, so do not run `docker image prune -a` on
the stage host.

## Stop

```bash
deploy/stage/deploy.sh down
```

Volumes (`rankr-stage_postgres-data`, `rankr-stage_redis-data`,
`rankr-stage_nats-data`) are kept. Never run `docker system prune --volumes`
on a shared host: it deletes other projects' data too.

## Resource limits

Stage shares a small host with other projects, so the stack is capped:

| Container | Memory |
|---|---|
| postgres | 512 MB (`max_connections=200`) |
| redis | 192 MB (`maxmemory 128mb`, `noeviction`) |
| nats | 384 MB (JetStream: 256 MB memory, 2 GB disk) |
| each service | 256 MB |
| gateway | 64 MB |

## Known gaps

- `GET /v1/leaderboard/public/:project_id` passes the gateway without a token
  but leaderboardstat itself still answers 401: `pkg/httpserver` only skips
  exact public paths, and this path has a parameter. Fixing it needs a
  prefix skipper in `pkg/httpserver` (TD-04, #221).
- NATS has no authentication yet. It is only reachable on the private network.
- The webhook `recovery_config` in the production config holds placeholder
  hook credentials, so lost-delivery recovery logs errors until real values
  are set.
- Images build from the production Dockerfiles, which need TD-16 (#220) to
  build from a clean clone.
