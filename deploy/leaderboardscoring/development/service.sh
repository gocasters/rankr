#!/bin/bash

# A unified script to manage the leaderboardscoring service lifecycle.
# It handles different running modes for development, including running the full
# stack with Docker or running only the dependencies for local Go development.

set -e

# --- Configuration ---
# Path to the full docker-compose file (app + dependencies)
COMPOSE_FILE_FULL="deploy/leaderboardscoring/development/docker-compose.yml"
# Path to the docker-compose file for running only the dependencies
COMPOSE_FILE_DEPS_ONLY="deploy/leaderboardscoring/development/docker-compose.no-service.yml"
# Go service configuration
SERVICE_NAME="leaderboardscoring"
SERVICE_PATH="./cmd/${SERVICE_NAME}/main.go"

COMMAND=$1

# --- Help Function ---
show_help() {
    echo "Usage: ./service.sh <command>"
    echo ""
    echo "This script is the single entry point for managing the ${SERVICE_NAME} service."
    echo ""
    echo "Available commands:"
    echo "  --- Full Dockerized Environment ---"
    echo "  up      Builds and starts the full application stack (app + dependencies)."
    echo "  stop    Stops the full application stack without deleting data."
    echo "  down    Stops and removes the full application stack and its data volumes."
    echo "  logs    Follows the logs for the full application stack."
    echo ""
    echo "  --- Local Development Helpers ---"
    echo "  up-deps   Starts only the dependency services (Postgres, Redis, NATS)."
    echo "  run       Starts the Go service locally (requires dependencies to be running)."
    echo "  stop-deps Stops the standalone dependency services."
    echo "  down-deps Stops and removes the standalone dependency services and their data."
    echo "  logs-deps Follows the logs for the standalone dependency services."
    echo ""
    echo "  --- General ---"
    echo "  help      Shows this help message."
}

# --- Command Logic ---
case "$COMMAND" in
    # --- Full Dockerized Environment Commands ---
    up)
        echo "--> Starting the full application stack from ${COMPOSE_FILE_FULL}..."
        docker compose -f "${COMPOSE_FILE_FULL}" up --build -d
        echo "--> Full stack is up and running."
        ;;
    stop)
        echo "--> Stopping the full application stack from ${COMPOSE_FILE_FULL}..."
        docker compose -f "${COMPOSE_FILE_FULL}" stop
        echo "--> Full stack stopped."
        ;;
    down)
        echo "--> Tearing down the full application stack from ${COMPOSE_FILE_FULL}..."
        docker compose -f "${COMPOSE_FILE_FULL}" down -v
        echo "--> Full stack and all associated data have been removed."
        ;;
    logs)
        echo "--> Following logs for the full application stack from ${COMPOSE_FILE_FULL}..."
        docker compose -f "${COMPOSE_FILE_FULL}" logs -f
        ;;

    # --- Local Development Helper Commands ---
    up-deps)
        echo "--> Starting only the dependency services from ${COMPOSE_FILE_DEPS_ONLY}..."
        docker compose -f "${COMPOSE_FILE_DEPS_ONLY}" up -d
        echo "--> Dependency services are up and running."
        ;;
    run)
        echo "--> Starting the Go service locally: ${SERVICE_NAME}..."
        echo "--> Ensure dependencies are running (e.g., via './service.sh up-deps')."
        go run "${SERVICE_PATH}" serve --migrate-up
        ;;
    stop-deps)
        echo "--> Stopping the standalone dependencies from ${COMPOSE_FILE_DEPS_ONLY}..."
        docker compose -f "${COMPOSE_FILE_DEPS_ONLY}" stop
        echo "--> Standalone dependencies stopped."
        ;;
    down-deps)
        echo "--> Tearing down the standalone dependencies from ${COMPOSE_FILE_DEPS_ONLY}..."
        docker compose -f "${COMPOSE_FILE_DEPS_ONLY}" down -v
        echo "--> Standalone dependencies and their data have been removed."
        ;;
    logs-deps)
        echo "--> Following logs for the standalone dependencies from ${COMPOSE_FILE_DEPS_ONLY}..."
        docker compose -f "${COMPOSE_FILE_DEPS_ONLY}" logs -f
        ;;

    # --- Help Command ---
    help|--help|-h|*)
        show_help
        exit 1
        ;;
esac

exit 0

