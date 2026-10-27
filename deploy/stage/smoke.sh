#!/usr/bin/env bash
# Smoke test of stage through the gateway (#227). deploy.sh runs it after the
# health check; it exits non-zero when any check fails, so deploy rolls back.
#
#   smoke.sh [base-url]   default http://127.0.0.1:8088
#
# SMOKE_PROJECT_ID picks the project for the public leaderboard (default 1).
# A project with no scores still answers 200 with an empty "rows" list.
set -uo pipefail

BASE_URL="${1:-http://127.0.0.1:8088}"
PROJECT_ID="${SMOKE_PROJECT_ID:-1}"
failed=0

BODY_FILE="$(mktemp)"
trap 'rm -f "$BODY_FILE"' EXIT

# check NAME STATUS BODY_PART CURL_ARGS...
# Passes when the response has STATUS and, unless BODY_PART is empty, its body
# contains BODY_PART. STATUS may list several codes separated by "|", such as
# "400|401". Each request is capped at 5 seconds.
check() {
    local name="$1" want="$2" want_body="$3" status
    shift 3
    : > "$BODY_FILE"
    status="$(curl -sS --max-time 5 -o "$BODY_FILE" -w '%{http_code}' "$@")" || status="000"
    if [[ "|$want|" == *"|$status|"* ]] && { [[ -z "$want_body" ]] || grep -qF -- "$want_body" "$BODY_FILE"; }; then
        echo "ok    $name ($status)"
    else
        echo "FAIL  $name (want $want${want_body:+ with $want_body}, got $status)"
        failed=1
    fi
}

check "public leaderboard" 200 '"rows"' \
    "$BASE_URL/v1/leaderboard/public/$PROJECT_ID"

# One protected route per auth block in gateway/nginx.conf, so a block that
# loses its auth include is caught.
check "projects without token" 401 '"unauthorized"' \
    "$BASE_URL/v1/projects"

check "profile without token" 401 '"unauthorized"' \
    "$BASE_URL/v1/profile/1"

check "contributor stats without token" 401 '"unauthorized"' \
    "$BASE_URL/v1/contributors/1/stats"

check "unknown route" 404 '"route_not_found"' \
    "$BASE_URL/rankr-smoke-no-such-route"

# Without the X-GitHub-* headers webhook answers 400 and publishes nothing,
# or 401 once it checks the delivery signature (TD-08 / #224). Either shows the
# request reached the route; the gateway itself would answer 404 or 502 here.
check "webhook without token" '400|401' '' \
    -X POST -H 'Content-Type: application/json' --data '{}' \
    "$BASE_URL/github-webhook/github/process"

exit "$failed"
