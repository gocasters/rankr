#!/usr/bin/env bash
# Forced command for the CI deploy key (#227). The key's authorized_keys line
# runs this script for every login, whatever command the client sent, so the
# key can only deploy a commit that is already on origin/main:
#
#   ssh <stage> "deploy <full commit sha>"
#
# The body is a function called on the last line: checkout below may replace
# this file, and bash has read the whole function before that happens.
set -euo pipefail

main() {
    local repo_dir action sha extra
    repo_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"

    read -r action sha extra <<< "${SSH_ORIGINAL_COMMAND:-}"
    if [[ "$action" != deploy || ! "$sha" =~ ^[0-9a-f]{40}$ || -n "$extra" ]]; then
        echo "usage: deploy <full commit sha>" >&2
        exit 2
    fi

    # One deploy at a time, CI or by hand through this script.
    exec 9> "$repo_dir/deploy/stage/.deploy-lock"
    if ! flock -w 600 9; then
        echo "another deploy is still running" >&2
        exit 1
    fi

    git -C "$repo_dir" fetch --quiet origin main
    if ! git -C "$repo_dir" merge-base --is-ancestor "$sha" FETCH_HEAD; then
        echo "$sha is not on origin/main" >&2
        exit 1
    fi
    git -C "$repo_dir" checkout --quiet --detach "$sha"

    "$repo_dir/deploy/stage/deploy.sh" deploy
}

main "$@"
exit
