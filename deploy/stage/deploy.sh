#!/usr/bin/env bash
# Stage deploy helper for rankr. See docs/stage.md.
#
#   deploy.sh deploy     build the current checkout, start it, check health;
#                        rolls back to the previous release if health fails
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

if [[ ! -f "$ENV_FILE" ]]; then
    echo "missing $ENV_FILE (copy .env.stage.example and fill it)" >&2
    exit 1
fi

# `|| true`: GATEWAY_PORT is optional, and grep finding nothing must not stop
# the script under pipefail.
GATEWAY_PORT="$(grep -E '^GATEWAY_PORT=' "$ENV_FILE" | cut -d= -f2 || true)"
GATEWAY_PORT="${GATEWAY_PORT:-8088}"

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
health() {
    local deadline="${1:-}" failed=0 timeout remaining
    for svc in "${SERVICES[@]}"; do
        timeout=5
        if [[ -n "$deadline" ]]; then
            remaining=$(( deadline - SECONDS ))
            if (( remaining <= 0 )); then
                echo "SKIP  $svc (deadline reached)"
                failed=1
                continue
            fi
            if (( remaining < timeout )); then timeout=$remaining; fi
        fi
        if curl -fsS -o /dev/null --max-time "$timeout" "http://127.0.0.1:$GATEWAY_PORT/health/$svc"; then
            echo "ok    $svc"
        else
            echo "FAIL  $svc"
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

case "${1:-}" in
    deploy)
        tag="$(git -C "$STAGE_DIR" rev-parse --short HEAD)"
        prev="$(previous_tag)"
        echo "deploying $tag (previous: ${prev:-none})"
        TAG="$tag" compose build
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
