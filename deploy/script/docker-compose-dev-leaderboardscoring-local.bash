#!/bin/bash

set -e

COMPOSE_FILE_FULL="../leaderboardscoring/development/docker-compose.yml"
SERVICE_NAME="leaderboardscoring"
COMMAND=$1

# --- Help Function ---
show_help() {
    echo "Usage: ./service.sh <command>"
    echo ""
    echo "This script is the single entry point for managing the ${SERVICE_NAME} service."
    echo ""
    echo "Available commands:"
    echo "  --- Dockerized leaderboardscoring application ---"
    echo "  up      Builds and starts the leaderboardscoring application"
    echo "  stop    Stops the leaderboardscoring application without deleting data."
    echo "  down    Stops and removes the leaderboardscoring application."
    echo "  logs    Follows the logs for the leaderboardscoring application."
    echo ""
    echo "  --- General ---"
    echo "  help      Shows this help message."
}

# --- Command Logic ---
case "$COMMAND" in
    up)
        echo "--> Starting the leaderboardscoring application from ${COMPOSE_FILE_FULL}..."
        docker compose -f "${COMPOSE_FILE_FULL}" up --build -d
        echo "--> Full stack is up and running."
        ;;
    stop)
        echo "--> Stopping the leaderboardscoring application from ${COMPOSE_FILE_FULL}..."
        docker compose -f "${COMPOSE_FILE_FULL}" stop
        echo "--> Full stack stopped."
        ;;
    down)
        echo "--> Tearing down the leaderboardscoring application from ${COMPOSE_FILE_FULL}..."
        docker compose -f "${COMPOSE_FILE_FULL}" down -v
        echo "--> Full stack and all associated data have been removed."
        ;;
    logs)
        echo "--> Following logs for the leaderboardscoring application from ${COMPOSE_FILE_FULL}..."
        docker compose -f "${COMPOSE_FILE_FULL}" logs -f
        ;;

    # --- Help Command ---
    help|--help|-h|*)
        show_help
        exit 1
        ;;
esac

exit 0

