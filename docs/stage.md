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
| `ci-deploy.sh` | Forced command for the CI deploy key (automatic deploy) |
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

Pushes to `main` deploy automatically, except docs-only pushes; intermediate pushes may be skipped (see [Automatic deploy](#automatic-deploy)).
To deploy by hand, use the same script as CI, so it waits for any running
deploy and checks out the tip of `origin/main` inside the same lock:

```bash
cd /opt/rankr-stage
deploy/stage/ci-deploy.sh deploy
```

Don't run `git checkout` or `deploy.sh deploy` directly while CI may be
deploying: that bypasses the lock.

The checkout in `/opt/rankr-stage` is on a detached HEAD and is managed by
`ci-deploy.sh`, which fetches and checks out `origin/main` itself. Don't
`git pull` or commit in it.

`deploy` builds images tagged with the short commit hash and starts them.

After starting, `deploy` checks `/health/<service>` for all six services
through the gateway, then the smoke test below. If a check does not pass
within `HEALTH_TIMEOUT` seconds, it starts the previous release again and exits
non-zero. Each healthy release is appended to `deploy/stage/.deploy-history`.

Smoke test: with `SMOKE_PROJECT_ID` set in `.env.stage`, the checks also call
`GET /v1/leaderboard/public/<SMOKE_PROJECT_ID>` and expect a 2xx. It is empty
by default: until TD-04 (#221) lands, leaderboardstat answers 401 on that path,
so turning it on would roll back every deploy. Once #221 is deployed, set it to
a project that exists on stage.

`HEALTH_TIMEOUT` is the whole wait, not a per-request limit: every request
(at most 5 seconds each) is cut short at the deadline, and services not
checked by then are reported as `SKIP`. It defaults to 150, must be a positive
whole number of seconds, and is read from the shell environment, not from
`.env.stage`. It applies to `deploy` and `rollback`:

```bash
HEALTH_TIMEOUT=60 deploy/stage/ci-deploy.sh deploy
```

## Automatic deploy

`.github/workflows/deploy-stage.yml` runs on every push to `main` (pushes that
only touch docs or `*.md` are skipped) and on demand from the Actions tab.
Deploys run one at a time. The runner only opens an SSH connection and sends
`deploy <commit sha>`; it never gets a shell.

Not every push gets its own deploy. GitHub keeps only the newest pending run,
and the script always deploys the current tip of `origin/main`, not the sha it
was sent. So when several pushes land close together, the intermediate commits
may be skipped, but stage always ends on the latest `main` and never moves
backward. A run whose tip is already deployed exits without restarting
anything.

On the host, the CI key is pinned to `deploy/stage/ci-deploy.sh`. It accepts
nothing but `deploy [full sha]`, takes a lock so two deploys cannot overlap,
refuses a sha that is not on `origin/main`, checks out the tip of
`origin/main` and runs `deploy.sh deploy`. A failed health check rolls back as above and
fails the workflow run.

The deploy runs detached from the SSH session, so cancelling the run, the job
timeout or a dropped connection does not stop it halfway: it still ends healthy
or rolled back, and keeps the lock until then. The run only shows its output;
the full log of the last deploy is in `deploy/stage/.deploy-log`.

One-time setup (needs host and repo admin):

1. Make a key pair just for this, without a passphrase:
   `ssh-keygen -t ed25519 -N "" -C rankr-stage-ci -f rankr-stage-ci`
2. On the host, in `~/.ssh/authorized_keys` of the user that owns
   `/opt/rankr-stage` and can run docker, add one line:

   ```
   restrict,command="/opt/rankr-stage/deploy/stage/ci-deploy.sh" ssh-ed25519 AAAA... rankr-stage-ci
   ```

3. In GitHub, Settings → Environments, create `stage`, limit it to the `main`
   branch, and add these environment secrets:

   | Secret | Value |
   |---|---|
   | `STAGE_SSH_KEY` | the private key file `rankr-stage-ci` |
   | `STAGE_SSH_TARGET` | `user@host` |
   | `STAGE_SSH_PORT` | only if SSH is not on 22 |
   | `STAGE_KNOWN_HOSTS` | output of `ssh-keyscan -p <port> -t ed25519 <host>` (keep `-p` even for 22), compared with `/etc/ssh/ssh_host_ed25519_key.pub` on the host. `<host>` must exactly match the host in `STAGE_SSH_TARGET` (same name or IP) |

4. Delete the local private key file once it is stored in GitHub.

To revoke, delete the line from `authorized_keys`.

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
deploy/stage/ci-deploy.sh rollback
```

This starts the previous release from `.deploy-history` without building.
`ci-deploy.sh` holds the deploy lock, so a rollback never runs during
a CI deploy. The next push to `main` deploys the new tip again; to keep stage
on the old release, pause the `deploy-stage` workflow first.
Migrations run on startup (`serve --migrate-up`) and are not rolled back: if a
release added a migration the previous code cannot handle, fix forward instead.
Old images are needed for rollback, so do not run `docker image prune -a` on
the stage host.

## Stop

```bash
deploy/stage/ci-deploy.sh down
```

`rollback` and `down` go through `ci-deploy.sh` so they wait for a running
deploy and hold the same lock. They are accepted only by hand on the host; the
CI key can only send `deploy`. The next push to `main` deploys the tip again.

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
  prefix skipper in `pkg/httpserver` (TD-04, #221). Until then, leave
  `SMOKE_PROJECT_ID` empty.
- NATS has no authentication yet. It is only reachable on the private network.
- The webhook `recovery_config` in the production config holds placeholder
  hook credentials, so lost-delivery recovery logs errors until real values
  are set.
