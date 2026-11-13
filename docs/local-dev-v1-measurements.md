# Local dev stack — simple measurement snapshot

Use this file as raw material for an issue, PR description, or docs. Numbers below are **informal** (single snapshot, no warm-up or scripted workload).

## When and what

| Field | Value |
|-------|--------|
| **Date** | 2026-11-06 (local) |
| **Command / stack** | `make up-v1`–style components (six v1 app containers + Postgres, Redis, NATS, OpenResty; Redis Insights present with `make infra-up` redis compose) |

---

## Host

| Field | Value |
|-------|--------|
| **OS** | macOS 26.5.1 (build 25F80) |
| **Arch** | arm64 |
| **CPU** | Apple M1 |

---

## Docker

| Field | Value |
|-------|--------|
| **Docker Desktop / Engine** | 28.1.1 |
| **Compose** | v2.35.1-desktop.1 (CLI plugin) |
| **Context** | `desktop-linux` |
| **`docker info` — VM** | Name: `docker-desktop`; OSType: `linux`; Architecture: `aarch64` |
| **`docker info` — Total Memory** | **3.828 GiB** (Docker Desktop VM allocation as reported by the engine) |
| **`docker info` — CPUs** | 8 |

> **Note:** “Total Memory” from `docker info` is the **memory allocated to the Docker VM**, not your Mac’s physical RAM. For the limit set in Docker Desktop → Settings → Resources, compare with this value and update the table if they differ.

---

## Shared development image

| Image | Size (`docker images`) |
|-------|-------------------------|
| `rankr-go-dev:1.25.3-alpine` | **1.41 GB** |

Built from `deploy/development/Dockerfile` (shared by dev compose files).

---

## Container memory — one snapshot (`docker stats --no-stream`)

Captured while the listed containers were running. Format: **usage / limit** (per container limits come from compose `mem_limit` where set).

| Container | Mem usage / limit | % of limit |
|-----------|-------------------|------------|
| `auth-app` | 34.84 MiB / 3.828 GiB | 0.89% |
| `contributor-app` | 37.12 MiB / 3.828 GiB | 0.95% |
| `leaderboardscoring-app` | 40.63 MiB / 3.828 GiB | 1.04% |
| `leaderboardstat-app` | 38.02 MiB / 3.828 GiB | 0.97% |
| `project-app` | 40.25 MiB / 3.828 GiB | 1.03% |
| `webhook-app` | 36.47 MiB / 3.828 GiB | 0.93% |
| `rankr-shared-postgres` | 58.69 MiB / **2 GiB** | 2.87% |
| `rankr-shared-redis` | 32.61 MiB / **512 MiB** | 6.37% |
| `rankr-redisinsight` | 212.8 MiB / **512 MiB** | 41.56% |
| `rankr-shared-nats` | 21.86 MiB / 3.828 GiB | 0.56% |
| `rankr-infra-nginx-1` (OpenResty) | 14.64 MiB / 3.828 GiB | 0.37% |

**Rough sum of “usage” above (MiB):** ~567 MiB (apps + infra rows in the table only; excludes any other containers such as `rankr-base-placeholder`).

---

## Optimizations referenced in discussion (for PR text)

Copy or trim as needed:

1. **Single Go/Air image** (`rankr-go-dev`) for all dev services — one build surface instead of per-service Dockerfiles.
2. **Sequential bring-up** (`start-and-wait-*` in `make up-v1`) — services start **in order** so workloads are not all compiling at once; Compose still waits at most **`DEV_WAIT_TIMEOUT`** (default 20s) per healthcheck, so a slow cold compile can exceed the window—see Makefile / README (`DEV_WAIT_TIMEOUT`, `FORCE=1 make docker-build-dev`).
3. **Shared Go build cache volume** (`rankr-go-build-cache`) — less repeated compilation after restarts.
4. **Narrower Air `include_dir`** — smaller watch surface, less reload churn on file changes.

### Makefile notes (intended trade-offs)

| Topic | Behavior |
|-------|----------|
| **Health wait** | `start-and-wait-<svc>-app-dev` uses `docker compose up --wait` with `--wait-timeout $(DEV_WAIT_TIMEOUT)`. On timeout it **fails** that step; `up-v1` runs **all** v1 services anyway, then exits **non-zero** and **does not** print “✅ All v1 services are up!” unless every wait succeeded. |
| **Long cold compile** | Override e.g. `DEV_WAIT_TIMEOUT=120 make up-v1`. |
| **`docker-build-dev`** | Skips `docker build` if `rankr-go-dev:$(DEFAULT_GO_IMAGE_VERSION)` already exists. After **Dockerfile / Air / go.mod** changes, run **`FORCE=1 make docker-build-dev`**. |

---

## How to refresh this file

```bash
# Host / Docker
sw_vers
uname -m
sysctl -n machdep.cpu.brand_string
docker version
docker info | grep -E "Server Version|Total Memory|CPUs|Name:"

# Image
docker images rankr-go-dev --format "table {{.Repository}}:{{.Tag}}\t{{.Size}}"

# Commit
git rev-parse HEAD
git rev-parse --short HEAD

# Containers (after make up-v1 or equivalent)
docker stats --no-stream --format "{{.Name}}\t{{.MemUsage}}\t{{.MemPerc}}"
```

Then paste updated tables and bump the date.
