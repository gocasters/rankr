#!/usr/bin/env bash
# Stage deploy helper for rankr. See docs/stage.md.
#
#   deploy.sh deploy     rebuild the services changed since the last release,
#                        restart every service, check health; rolls back to
#                        the previous release if health fails
#   deploy.sh rollback   start the previous release again (no build)
#   deploy.sh status     container status
#   deploy.sh logs [svc] follow logs
#   deploy.sh health     health of every v1 service through the gateway
#   deploy.sh down       stop the stack (volumes are kept)
set -euo pipefail

STAGE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ENV_FILE="${ENV_FILE:-$STAGE_DIR/.env.stage}"
HISTORY="$STAGE_DIR/.deploy-history"
SERVICES=(auth contributor leaderboardscoring leaderboardstat project webhook)

# Paths only one service (or a few) is built from. Keep in sync with
# `go list -deps ./cmd/<service>`: leaderboardstat imports leaderboardscoringapp
# and project imports webhookapp. Any changed path not listed here or in
# NO_SERVICE_PATHS rebuilds every service.
declare -A SERVICE_PATHS=(
    [auth]="authapp/ cmd/auth/ deploy/auth/"
    [contributor]="contributorapp/ cmd/contributor/ deploy/contributor/"
    [leaderboardscoring]="leaderboardscoringapp/ cmd/leaderboardscoring/ deploy/leaderboardscoring/"
    [leaderboardstat]="leaderboardstatapp/ leaderboardscoringapp/ cmd/leaderboardstat/ deploy/leaderboardstat/"
    [project]="projectapp/ webhookapp/ cmd/project/ deploy/project/"
    [webhook]="webhookapp/ cmd/webhook/ deploy/webhook/"
)
NO_SERVICE_PATHS="docs/ .github/ .claude/"

if [[ ! -f "$ENV_FILE" ]]; then
    echo "missing $ENV_FILE (copy .env.stage.example and fill it)" >&2
    exit 1
fi

# `|| true`: GATEWAY_PORT is optional, and grep finding nothing must not stop
# the script under pipefail.
GATEWAY_PORT="$(grep -E '^GATEWAY_PORT=' "$ENV_FILE" | cut -d= -f2 || true)"
GATEWAY_PORT="${GATEWAY_PORT:-8088}"

# Smoke test of the public leaderboard (#227). Off while SMOKE_PROJECT_ID is
# empty: leaderboardstat still answers 401 on that path until TD-04 (#221).
SMOKE_PROJECT_ID="$(grep -E '^SMOKE_PROJECT_ID=' "$ENV_FILE" | cut -d= -f2 || true)"
if [[ -n "$SMOKE_PROJECT_ID" && ! "$SMOKE_PROJECT_ID" =~ ^[0-9]+$ ]]; then
    echo "SMOKE_PROJECT_ID must be a project id (got '$SMOKE_PROJECT_ID')" >&2
    exit 1
fi

# Overall time budget for wait_healthy. Checked up front so a typo fails before
# anything is built, and so only a plain number ever reaches $(( )).
HEALTH_TIMEOUT="${HEALTH_TIMEOUT:-150}"
if [[ ! "$HEALTH_TIMEOUT" =~ ^[1-9][0-9]*$ ]]; then
    echo "HEALTH_TIMEOUT must be a positive whole number of seconds (got '$HEALTH_TIMEOUT')" >&2
    exit 1
fi

compose() {
    docker compose --env-file "$ENV_FILE" -f "$STAGE_DIR/docker-compose.yml" "$@"
}

# health [deadline]: without an argument (`deploy.sh health`) each service gets
# up to 5 seconds. With a deadline (a $SECONDS value), each request is bounded
# by the time left and no request is sent once the deadline has passed.
# With SMOKE_PROJECT_ID set, the public leaderboard is checked last.
health() {
    local deadline="${1:-}" failed=0 timeout remaining name path check
    local checks=()
    for svc in "${SERVICES[@]}"; do checks+=("$svc /health/$svc"); done
    if [[ -n "$SMOKE_PROJECT_ID" ]]; then
        checks+=("leaderboard /v1/leaderboard/public/$SMOKE_PROJECT_ID")
    fi
    for check in "${checks[@]}"; do
        read -r name path <<< "$check"
        timeout=5
        if [[ -n "$deadline" ]]; then
            remaining=$(( deadline - SECONDS ))
            if (( remaining <= 0 )); then
                echo "SKIP  $name (deadline reached)"
                failed=1
                continue
            fi
            if (( remaining < timeout )); then timeout=$remaining; fi
        fi
        if curl -fsS -o /dev/null --max-time "$timeout" "http://127.0.0.1:$GATEWAY_PORT$path"; then
            echo "ok    $name"
        else
            echo "FAIL  $name"
            failed=1
        fi
    done
    return $failed
}

# Waits up to HEALTH_TIMEOUT seconds overall. Every request and sleep is
# bounded by the deadline, and the final report reuses the last pass instead of
# sending new requests, so a hanging service cannot delay the rollback.
wait_healthy() {
    local deadline=$(( SECONDS + HEALTH_TIMEOUT ))
    local report="" remaining
    while (( SECONDS < deadline )); do
        if report="$(health "$deadline" 2>/dev/null)"; then
            echo "$report"
            return 0
        fi
        remaining=$(( deadline - SECONDS ))
        (( remaining > 0 )) || break
        sleep $(( remaining < 5 ? remaining : 5 ))
    done
    echo "${report:-no health check ran}"
    return 1
}

previous_tag() {
    [[ -f "$HISTORY" ]] && tail -n 1 "$HISTORY" || true
}

start() {
    TAG="$1" compose up -d --remove-orphans "${@:2}"
}

# changed_services <prev tag>: the services whose code changed between the
# previous release and HEAD, one per line. Prints every service when there is
# no previous release, the commit is unknown, or a shared path changed.
changed_services() {
    local prev="$1" file svc dir hit
    local -A changed=()
    if [[ -z "$prev" ]] || ! git -C "$STAGE_DIR" rev-parse -q --verify "$prev^{commit}" > /dev/null; then
        printf '%s\n' "${SERVICES[@]}"
        return
    fi
    while IFS= read -r file; do
        [[ "$file" == *.md ]] && continue
        hit=""
        for dir in $NO_SERVICE_PATHS; do
            [[ "$file" == "$dir"* ]] && hit=none
        done
        for svc in "${SERVICES[@]}"; do
            for dir in ${SERVICE_PATHS[$svc]}; do
                [[ "$file" == "$dir"* ]] && { changed[$svc]=1; hit=svc; }
            done
        done
        if [[ -z "$hit" ]]; then
            printf '%s\n' "${SERVICES[@]}"
            return
        fi
    done < <(git -C "$STAGE_DIR" diff --name-only "$prev" HEAD)
    for svc in "${SERVICES[@]}"; do
        [[ -n "${changed[$svc]:-}" ]] && echo "$svc"
    done
    return 0
}

case "${1:-}" in
    deploy)
        tag="$(git -C "$STAGE_DIR" rev-parse --short HEAD)"
        prev="$(previous_tag)"
        echo "deploying $tag (previous: ${prev:-none})"
        mapfile -t build < <(changed_services "$prev")
        if (( ${#build[@]} == ${#SERVICES[@]} )); then
            echo "rebuilding every service"
            TAG="$tag" compose build
        else
            # Unchanged services keep their image under the new tag, so
            # rollback and the next deploy find every service at every tag.
            # One whose previous image is gone is rebuilt instead.
            for svc in "${SERVICES[@]}"; do
                [[ " ${build[*]} " == *" $svc "* ]] && continue
                if docker image inspect "rankr-stage/$svc:$prev" > /dev/null 2>&1; then
                    docker tag "rankr-stage/$svc:$prev" "rankr-stage/$svc:$tag"
                else
                    build+=("$svc")
                fi
            done
            echo "rebuilding: ${build[*]:-none}"
            if (( ${#build[@]} > 0 )); then
                TAG="$tag" compose build "${build[@]}"
            fi
        fi
        # A new tag changes every service's image, so `up` recreates all of
        # them in depends_on order and .env.stage edits reach every service.
        # Deploying the last release again changes nothing compose can see,
        # so recreate the services by force: that replaces a hung container.
        if [[ "$tag" == "$prev" ]]; then
            start "$tag" --force-recreate "${SERVICES[@]}"
        fi
        start "$tag"
        if wait_healthy; then
            echo "$tag" >> "$HISTORY"
            echo "deployed $tag"
        else
            echo "health check failed for $tag" >&2
            if [[ -n "$prev" ]]; then
                echo "rolling back to $prev" >&2
                start "$prev" --no-build
                wait_healthy || echo "rollback is unhealthy too" >&2
            fi
            exit 1
        fi
        ;;
    rollback)
        if [[ ! -f "$HISTORY" ]] || [[ "$(wc -l < "$HISTORY")" -lt 2 ]]; then
            echo "no previous release to roll back to" >&2
            exit 1
        fi
        prev="$(tail -n 2 "$HISTORY" | head -n 1)"
        echo "rolling back to $prev"
        start "$prev" --no-build
        wait_healthy
        echo "$prev" >> "$HISTORY"
        ;;
    status)
        compose ps
        ;;
    logs)
        compose logs -f --tail=200 "${@:2}"
        ;;
    health)
        health
        ;;
    down)
        compose down
        ;;
    *)
        sed -n '2,11p' "$0"
        exit 1
        ;;
esac
