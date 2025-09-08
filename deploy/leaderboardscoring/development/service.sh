#!/bin/bash

# A unified script to manage the leaderboardscoring service lifecycle.
# It contains all the core logic for running, stopping, and managing the service
# and its dependencies via Docker Compose.

set -e

COMPOSE_FILE="deploy/leaderboardscoring/development/docker-compose.no-service.yml"
SERVICE_NAME="leaderboardscoring"
SERVICE_PATH="./cmd/${SERVICE_NAME}/main.go"

COMMAND=$1

show_help() {
    echo "Usage: ./service.sh <command>"
    echo ""
    echo "This script is the single entry point for managing the leaderboardscoring service."
    echo ""
    echo "Available commands:"
    echo "  up      Starts all required background services (Postgres, Redis, NATS) using Docker Compose."
    echo "  run     Starts the leaderboardscoring Go service."
    echo "  logs    Follows the logs of the background services."
    echo "  stop    Stops the background services without deleting data."
    echo "  down    Stops and removes all background services and their data volumes."
    echo "  help    Shows this help message."
}

case "$COMMAND" in
    up)
        echo "--> Starting background services from ${COMPOSE_FILE}..."
        docker compose -f "${COMPOSE_FILE}" up -d
        echo "--> Background services are up and running."
        ;;
    run)
        echo "--> Starting the Go service: ${SERVICE_NAME}..."
        go run "${SERVICE_PATH}" serve
        ;;
    logs)
        echo "--> Following logs for services in ${COMPOSE_FILE}..."
        docker compose -f "${COMPOSE_FILE}" logs -f
        ;;
    stop)
        echo "--> Stopping background services from ${COMPOSE_FILE}..."
        docker compose -f "${COMPOSE_FILE}" stop
        echo "--> Background services stopped."
        ;;
    down)
        echo "--> Tearing down all services and data from ${COMPOSE_FILE}..."
        docker compose -f "${COMPOSE_FILE}" down -v
        echo "--> All services and data have been removed."
        ;;
    help|--help|-h|*)
        show_help
        exit 1
        ;;
esac

exit 0

