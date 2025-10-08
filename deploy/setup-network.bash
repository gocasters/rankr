#!/usr/bin/env bash
set -e

NETWORK_NAME="rankr-development-network"

# Check if network exists, create if not
if ! docker network inspect "$NETWORK_NAME" >/dev/null 2>&1; then
    echo "Creating network: $NETWORK_NAME"
    docker network create "$NETWORK_NAME"
    # No need for the else part - silent if already exists
fi