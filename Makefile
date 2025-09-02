.PHONY: start test build mod-tidy lint docker-build docker-run help clean install-linter

BINARY_NAME ?= rankr
BUILD_DIR ?= bin

start: build
	./bin/rankr

test:
	go test -v ./...

build:
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) cmd/rankr/main.go

clean:
	rm -rf $(BUILD_DIR)/

mod-tidy:
	go mod tidy

lint:
	golangci-lint run

install-linter:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

help:
	@echo "Available targets:"
	@echo "  start          - Build and run locally"
	@echo "  test           - Run tests"
	@echo "  mod-tidy       - Clean up dependencies"
	@echo "  lint           - Run linters"
	@echo "  install-linter - Install golangci-lint"
	@echo "  build          - Compile binary (BINARY_NAME=$(BINARY_NAME), BUILD_DIR=$(BUILD_DIR))"
	@echo "  clean          - Remove build artifacts"
	@echo "Available targets:"
	@echo "  start     - Build and run locally"
	@echo "  test      - Run tests"
	@echo "  mod-tidy  - Clean up dependencies"
	@echo "  lint      - Run linters"
	@echo "  help      - Show this help"
	@echo "  build     - Compile binary (BINARY_NAME=$(BINARY_NAME), BUILD_DIR=$(BUILD_DIR))"
	@echo "  clean        - Remove build artifacts"



start-contributor-app-dev:
	./deploy/docker-compose-dev.bash --profile contributor up -d

start-contributor-app-dev-log:
	./deploy/docker-compose-dev.bash --profile contributor up

start-contributor-debug: ## Start user in debug mode (local)
	./deploy/docker-compose-dev-user-local.bash up -d

start-contributor-debug-log: ## Start user in debug mode (local) with logs
	./deploy/docker-compose-dev-user-local.bash up

start-task-app-dev:
	./deploy/docker-compose-dev.bash --profile task up

stop-contributor-debug: ## Stop user debug mode (keep volumes)
	docker compose \
	--env-file ./deploy/.env \
	--project-directory . \
	--profile user-local \
	-f ./deploy/rankr/development/traefik-compose.yml \
	-f ./deploy/user/development/docker-compose.no-service.yaml \
	down --remove-orphans
