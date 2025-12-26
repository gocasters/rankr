#!/bin/sh
set -eu


MAX_TRIES=${MAX_TRIES:-60}
SLEEP_SECS=${SLEEP_SECS:-2}

wait_for() {
  host="$1"
  port="$2"
  i=1
  echo "Waiting for $host:$port ..."
  while [ "$i" -le "$MAX_TRIES" ]; do
    if nc -z "$host" "$port" >/dev/null 2>&1; then
      echo "Ready: $host:$port"
      return 0
    fi
    echo "Still waiting $host:$port ($i/$MAX_TRIES)"
    i=$((i+1))
    sleep "$SLEEP_SECS"
  done
  echo "Timeout waiting for $host:$port after $MAX_TRIES attempts"
  return 1
}

# Required dependencies for leaderboardstat
wait_for leaderboardscoring-app 8091
wait_for project-app 8094

if [ "$#" -gt 0 ]; then
  exec "$@"
fi
exit 0
