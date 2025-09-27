# --- Variables ---
BINARY_NAME ?= rankr
BUILD_DIR ?= bin
BUF_VERSION ?= v1.56.0
DEFAULT_BRANCH ?= main
PROTOC_GEN_GO_VERSION ?= v1.34.2
PROTOC_GEN_GO_GRPC_VERSION ?= v1.5.1
# Use a variable for the leaderboard script path for cleanliness
LB_SCRIPT := ./deploy/leaderboardscoring/development/service.sh

# ====================================================================================
# General Go Commands
# ====================================================================================
.PHONY: start test build clean mod-tidy lint install-linter help

start: build
	$(BUILD_DIR)/$(BINARY_NAME)

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

# ====================================================================================
# Protobuf Commands
# ====================================================================================
.PHONY: proto-setup proto-setup-full proto-gen proto-lint proto-breaking proto-clean proto-format proto-deps proto-validate
.PHONY: install-protoc-plugins install-buf install-buf-force
.PHONY: proto-bsr-push proto-bsr-push-create proto-bsr-info proto-bsr-login proto-bsr-whoami update-buf-version

proto-setup: install-buf
	@echo "Setting up Buf for Rankr project..."
	@if [ ! -f "protobuf/buf.yaml" ]; then \
	   echo "Initializing Buf configuration..."; \
	   cd protobuf && buf mod init; \
	fi
	@echo "Updating protobuf dependencies..."
	cd protobuf && buf dep update
	@echo "Linting protobuf files..."
	buf lint
	@echo "Buf setup complete!"

proto-setup-full: proto-setup install-protoc-plugins
	@echo "Generating Go code from protobuf..."
	buf generate
	@echo "Full Buf setup complete!"

proto-gen:
	@echo "Generating protobuf code..."
	buf generate

proto-lint:
	@echo "Linting protobuf files..."
	buf lint

proto-breaking:
	@echo "Checking for breaking changes..."
	@if git rev-parse --git-dir > /dev/null 2>&1; then \
	   echo "Git repository found, checking against $(DEFAULT_BRANCH) branch..."; \
	   buf breaking --against '.git#branch=$(DEFAULT_BRANCH)'; \
	else \
	   echo "No Git repository found. Skipping breaking change check."; \
	fi

proto-clean:
	@echo "Cleaning generated protobuf files..."
	find protobuf/golang -type f \( -name "*.pb.go" -o -name "*_grpc.pb.go" \) -print -delete || true

proto-format:
	@echo "Formatting protobuf files..."
	buf format -w

proto-deps:
	@echo "Updating protobuf dependencies..."
	cd protobuf && buf dep update

proto-validate:
	@echo "Validating protobuf files..."
	buf lint
	@if git rev-parse --git-dir > /dev/null 2>&1; then \
	   echo "Git repository found, checking for breaking changes..."; \
	   buf breaking --against '.git#branch=$(DEFAULT_BRANCH)'; \
	else \
	   echo "No Git repository found. Skipping breaking change check."; \
	fi

# BSR (Buf Schema Registry) targets
proto-bsr-push:
	@echo "Pushing protobuf module to BSR..."
	cd protobuf && buf push

proto-bsr-push-create:
	@echo "Pushing protobuf module to BSR (create if not exists)..."
	cd protobuf && buf push --create

# ... other proto commands remain the same ...

install-protoc-plugins:
	@echo "Installing protoc plugins..."
	@if ! command -v protoc-gen-go &> /dev/null; then \
	   echo "Installing protoc-gen-go..."; \
	   go install google.golang.org/protobuf/cmd/protoc-gen-go@$(PROTOC_GEN_GO_VERSION); \
	else \
	   echo "protoc-gen-go is already installed"; \
	fi
	@if ! command -v protoc-gen-go-grpc &> /dev/null; then \
	   echo "Installing protoc-gen-go-grpc..."; \
	   go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@$(PROTOC_GEN_GO_GRPC_VERSION); \
	else \
	   echo "protoc-gen-go-grpc is already installed"; \
	fi

install-buf:
	@echo "Installing Buf $(BUF_VERSION)..."
	@if command -v buf &> /dev/null && buf --version &> /dev/null; then \
	   CURRENT_VERSION=$$(buf --version | sed 's/^v//'); \
	   EXPECTED_VERSION=$$(echo "$(BUF_VERSION)" | sed 's/^v//'); \
	   if [ "$$CURRENT_VERSION" = "$$EXPECTED_VERSION" ]; then \
	      echo "Buf $(BUF_VERSION) is already installed"; \
	   else \
	      echo "Buf version mismatch. Expected $(BUF_VERSION), found $$CURRENT_VERSION. Reinstalling..."; \
	      $(MAKE) install-buf-force; \
	   fi; \
	else \
	   echo "Buf not found, installing $(BUF_VERSION)..."; \
	   $(MAKE) install-buf-force; \
	fi

install-buf-force:
	@echo "Force installing Buf $(BUF_VERSION)..."
	@if command -v curl &> /dev/null; then \
	   curl -sSL "https://github.com/bufbuild/buf/releases/$(BUF_VERSION)/download/buf-$$(uname -s)-$$(uname -m)" -o /tmp/buf && \
	   chmod +x /tmp/buf && \
	   sudo mv /tmp/buf /usr/local/bin/buf && \
	   echo "Buf $(BUF_VERSION) installed successfully"; \
	else \
	   echo "curl not found. Please install Buf manually."; \
	   exit 1; \
	fi

# ====================================================================================
# Other Service Commands (Unchanged)
# ====================================================================================
.PHONY: start-contributor-app-dev start-contributor-app-dev-log start-contributor-debug start-contributor-debug-log
.PHONY: start-task-app-dev stop-contributor-debug start-project-app-dev start-project-app-dev-log

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

start-project-app-dev:
	./deploy/docker-compose-dev.bash --profile project up -d

start-project-app-dev-log:
	./deploy/docker-compose-dev.bash --profile project up

# ====================================================================================
# Leaderboard Service Lifecycle Commands (via service.sh)
# ====================================================================================
.PHONY: lb-up lb-down lb-stop lb-logs lb-up-deps lb-run lb-down-deps lb-stop-deps lb-logs-deps lb-help

# --- Full Dockerized Environment ---
lb-up: ## Build and start the full leaderboard stack (app + dependencies)
	@$(LB_SCRIPT) up

lb-down: ## Stop and remove the full leaderboard stack and its data
	@$(LB_SCRIPT) down

lb-stop: ## Stop the full leaderboard stack without deleting data
	@$(LB_SCRIPT) stop

lb-logs: ## Follow the logs for the full leaderboard stack
	@$(LB_SCRIPT) logs

# --- Local Development Helpers ---
lb-up-deps: ## Start only the dependency services (Postgres, Redis, NATS)
	@$(LB_SCRIPT) up-deps

lb-run: ## Start the Go service locally (requires dependencies to be running)
	@$(LB_SCRIPT) run

lb-down-deps: ## Stop and remove the standalone dependency services and their data
	@$(LB_SCRIPT) down-deps

lb-stop-deps: ## Stop the standalone dependency services
	@$(LB_SCRIPT) stop-deps

lb-logs-deps: ## Follow the logs for the standalone dependency services
	@$(LB_SCRIPT) logs-deps

lb-help: ## Show the help message for the leaderboard service script
	@$(LB_SCRIPT) help

# ====================================================================================
# Help Target
# ====================================================================================
help:
	@echo "Available targets:"
	@echo ""
	@echo "General Go Commands:"
	@echo "  start          - Build and run locally"
	@echo "  test           - Run tests"
	@echo "  mod-tidy       - Clean up dependencies"
	@echo "  lint           - Run linters"
	@echo "  install-linter - Install golangci-lint"
	@echo "  build          - Compile binary"
	@echo "  clean          - Remove build artifacts"
	@echo ""
	@echo "Leaderboard Service Lifecycle Commands (run 'make lb-help' for details):"
	@echo "  --- Full Dockerized Environment ---"
	@echo "  lb-up          - Build and start the full leaderboard stack (app + dependencies)"
	@echo "  lb-down        - Stop and remove the full leaderboard stack"
	@echo "  lb-stop        - Stop the full leaderboard stack"
	@echo "  lb-logs        - Follow the logs for the full stack"
	@echo "  --- Local Development Helpers ---"
	@echo "  lb-up-deps     - Start only the dependency services"
	@echo "  lb-run         - Start the Go service locally"
	@echo "  lb-down-deps   - Stop and remove the dependency services"
	@echo ""
	@echo "Protobuf targets (run 'make' to see all):"
	@echo "  proto-gen      - Generate Go code from protobuf files"
	@echo "  proto-lint     - Lint protobuf files"
	@echo "  ..."

