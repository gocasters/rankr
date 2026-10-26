#!/usr/bin/env bash
# Deploys the tip of origin/main to stage, one deploy at a time (#227).
#
# CI: the deploy key's authorized_keys line forces this script for every
# login, whatever command the client sent:
#
#   ssh <stage> "deploy <full commit sha>"
#
# By hand, on the host (takes the same lock as CI):
#
#   deploy/stage/ci-deploy.sh deploy
#
# The sha only has to be on origin/main; what gets deployed is always the
# current tip. So stage never moves backward when runs finish out of order,
# and a run GitHub cancelled while pending is covered by the next one.
#
# The body is a function called on the last line: checkout below may replace
# this file, and bash has read the whole function before that happens.
set -euo pipefail

main() {
    local repo_dir action sha extra tip
    repo_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"

    # Arguments only exist when run on the host; sshd passes none.
    if [[ $# -gt 0 ]]; then
        read -r action sha extra <<< "$*"
    else
        read -r action sha extra <<< "${SSH_ORIGINAL_COMMAND:-}"
    fi
    if [[ "$action" != deploy || -n "$extra" ]] ||
        [[ -n "$sha" && ! "$sha" =~ ^[0-9a-f]{40}$ ]]; then
        echo "usage: deploy [full commit sha]" >&2
        exit 2
    fi

    # One deploy at a time, CI or by hand. Checkout happens inside the lock.
    exec 9> "$repo_dir/deploy/stage/.deploy-lock"
    if ! flock -w 600 9; then
        echo "another deploy is still running" >&2
        exit 1
    fi

    git -C "$repo_dir" fetch --quiet origin main
    tip="$(git -C "$repo_dir" rev-parse FETCH_HEAD)"
    if [[ -n "$sha" ]] && ! { git -C "$repo_dir" cat-file -e "$sha^{commit}" 2>/dev/null &&
        git -C "$repo_dir" merge-base --is-ancestor "$sha" "$tip"; }; then
        echo "$sha is not on origin/main" >&2
        exit 1
    fi

    local history="$repo_dir/deploy/stage/.deploy-history"
    # History alone is not proof: `deploy.sh down` keeps it, so also require
    # the running stack to pass the health check before skipping.
    if [[ -f "$history" && "$(tail -n 1 "$history")" == "$(git -C "$repo_dir" rev-parse --short "$tip")" ]] &&
        "$repo_dir/deploy/stage/deploy.sh" health > /dev/null 2>&1; then
        echo "origin/main ($tip) is already deployed and healthy"
        exit 0
    fi

    [[ -n "$sha" && "$sha" != "$tip" ]] && echo "$sha is behind origin/main; deploying $tip"
    git -C "$repo_dir" checkout --quiet --detach "$tip"

    "$repo_dir/deploy/stage/deploy.sh" deploy
}

main "$@"
exit
