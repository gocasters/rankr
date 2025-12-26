
# --- Variables ---
SHELL := /bin/bash
BINARY_NAME ?= rankr
BUILD_DIR ?= bin
BUF_VERSION ?= v1.56.0
DEFAULT_BRANCH ?= main
PROTOC_GEN_GO_VERSION ?= v1.34.2
PROTOC_GEN_GO_GRPC_VERSION ?= v1.5.1
DOCKER_COMPOSE ?= docker compose
INFRA_SCRIPT := ./deploy/script/start_infrastructure.sh
SERVICES := auth contributor leaderboardscoring leaderboardstat notification project realtime task userprofile webhook

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

proto-bsr-info:
	@echo "Getting BSR module info..."
	cd protobuf && buf push --info

proto-bsr-login:
	@echo "Logging into BSR..."
	buf registry login

proto-bsr-whoami:
	@echo "Checking BSR login status..."
	buf registry whoami

update-buf-version:
	@echo "Updating Buf version to $(BUF_VERSION)..."
	sed -i 's/version: .*/version: $(BUF_VERSION)/' protobuf/buf.yaml

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
	   CURRENT_VERSION=$(buf --version | sed 's/^v//'); \
	   EXPECTED_VERSION=$(echo "$(BUF_VERSION)" | sed 's/^v//'); \
	   if [ "$CURRENT_VERSION" = "$EXPECTED_VERSION" ]; then \
	      echo "Buf $(BUF_VERSION) is already installed"; \
	   else \
	      echo "Buf version mismatch. Expected $(BUF_VERSION), found $CURRENT_VERSION. Reinstalling..."; \
	      $(MAKE) install-buf-force; \
	   fi; \
	else \
	   echo "Buf not found, installing $(BUF_VERSION)..."; \
	   $(MAKE) install-buf-force; \
	fi

install-buf-force:
	@echo "Force installing Buf $(BUF_VERSION)..."
	@if command -v curl &> /dev/null; then \
	   ARCHIVE="/tmp/buf-$(BUF_VERSION).tar.gz"; \
	   curl -sSL "https://github.com/bufbuild/buf/releases/download/v$(BUF_VERSION)/buf-$$(uname -s)-$$(uname -m).tar.gz" -o $$ARCHIVE && \
	   tar -xzf $$ARCHIVE -C /tmp && \
	   chmod +x /tmp/buf/bin/buf && \
	   sudo mv /tmp/buf/bin/buf /usr/local/bin/buf && \
	   rm -rf $$ARCHIVE /tmp/buf && \
	   echo "Buf $(BUF_VERSION) installed successfully"; \
else \
	   echo "curl not found. Please install Buf manually."; \
	   exit 1; \
	fi

# ====================================================================================
# Infrastructure Commands
# ====================================================================================
.PHONY: infra-up infra-down infra-logs
.PHONY: infra-up-postgres infra-up-redis infra-up-nats infra-up-emqx

infra-up:
	@echo "Starting shared infrastructure..."
	bash $(INFRA_SCRIPT) up-all

infra-down:
	@echo "Stopping shared infrastructure..."
	bash $(INFRA_SCRIPT) down-all

infra-logs:
	@echo "Showing logs for infrastructure stack..."
	bash $(INFRA_SCRIPT) logs-all

infra-up-postgres:
	@echo "Starting infrastructure PostgreSQL..."
	bash $(INFRA_SCRIPT) up-postgres

infra-up-redis:
	@echo "Starting infrastructure Redis..."
	bash $(INFRA_SCRIPT) up-redis

infra-up-nats:
	@echo "Starting infrastructure NATS..."
	bash $(INFRA_SCRIPT) up-nats

infra-up-emqx:
	@echo "Starting infrastructure EMQX..."
	bash $(INFRA_SCRIPT) up-emqx

# ====================================================================================
# Service Commands
# ====================================================================================
.PHONY: $(SERVICES:%=start-%-app-dev) $(SERVICES:%=start-%-app-dev-log) $(SERVICES:%=stop-%-app-dev)

define SERVICE_template
start-$(1)-app-dev:
	@echo "Starting $(1) service..."
	cd deploy/$(1)/development && PROJECT_ROOT=$(CURDIR) $(DOCKER_COMPOSE) up -d --build

start-$(1)-app-dev-log:
	@echo "Starting $(1) service with logs..."
	cd deploy/$(1)/development && PROJECT_ROOT=$(CURDIR) $(DOCKER_COMPOSE) up --build

stop-$(1)-app-dev:
	@echo "Stopping $(1) service..."
	cd deploy/$(1)/development && PROJECT_ROOT=$(CURDIR) $(DOCKER_COMPOSE) down
endef

$(foreach svc,$(SERVICES),$(eval $(call SERVICE_template,$(svc))))

.PHONY: services-up services-down services-logs

# All services commands
services-up:
	@echo "Starting all application services..."
	@for svc in $(SERVICES); do \
		$(MAKE) start-$$svc-app-dev; \
	done
	@echo "✅ All services are up!"

services-down:
	@echo "Stopping all application services..."
	@for svc in $(SERVICES); do \
		$(MAKE) stop-$$svc-app-dev; \
	done
	@echo "✅ All services stopped!"

services-logs:
	@echo "Showing all services logs..."
	@trap 'for job in $$(jobs -p); do kill $$job 2>/dev/null || true; done' SIGINT SIGTERM EXIT; \
	for svc in $(SERVICES); do \
		( \
			echo ""; \
			echo ">>> $$svc logs"; \
			cd deploy/$$svc/development && $(DOCKER_COMPOSE) logs -f \
		) & \
	done; \
	wait

# Complete startup (infrastructure + services)
up:
	$(MAKE) infra-up
	@echo "Waiting for infrastructure to be ready..."
	@sleep 10
	@echo "Starting all services..."
	$(MAKE) services-up
	@echo "✅ All services are up!"

# Complete shutdown
down:
	$(MAKE) services-down
	$(MAKE) infra-down
	@echo "✅ All services stopped!"

# Show logs for everything
logs:
	@echo "Showing logs..."
	$(MAKE) infra-logs

# Restart everything
restart: down
	@sleep 2
	$(MAKE) up

# Show status of all containers
status:
	@echo "Container status:"
	@docker ps -a --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}" | grep -E "rankr|NAMES"

# ====================================================================================
# Help
# ====================================================================================
.PHONY: help

help:
	@echo "Available targets:"
	@echo "  Infrastructure:"
	@echo "    infra-up              - Start all infrastructure services"
	@echo "    infra-down            - Stop all infrastructure services"
	@echo "    infra-up-postgres     - Start PostgreSQL only"
	@echo "    infra-up-redis        - Start Redis only"
	@echo "    infra-up-nats         - Start NATS only"
	@echo "    infra-up-emqx         - Start EMQX only"
	@echo "    infra-logs            - Tail infrastructure logs"
	@echo ""
	@echo "  Services ($(SERVICES)):"
	@echo "    start-<service>-app-dev      - Start a single service (e.g. make start-auth-app-dev)"
	@echo "    start-<service>-app-dev-log  - Start a service and attach logs"
	@echo "    stop-<service>-app-dev       - Stop a single service"
	@echo "    services-up                  - Start every service listed above"
	@echo "    services-down                - Stop every service listed above"
	@echo "    services-logs                - Tail logs for every service (Ctrl+C to stop)"
	@echo ""
	@echo "  Protobuf:"
	@echo "    proto-setup           - Setup Buf for project"
	@echo "    proto-setup-full      - Setup Buf and generate code"
	@echo "    proto-gen             - Generate protobuf code"
	@echo "    proto-lint            - Lint protobuf files"
	@echo "    proto-breaking        - Check for breaking changes"
	@echo "    proto-clean           - Clean generated protobuf files"
	@echo "    proto-format          - Format protobuf files"
	@echo "    proto-deps            - Update protobuf dependencies"
	@echo "    proto-validate        - Validate protobuf files"
	@echo ""
	@echo "  Build & Development:"
	@echo "    build                 - Build the binary"
	@echo "    start                 - Run the binary"
	@echo "    test                  - Run tests"
	@echo "    clean                 - Clean build artifacts"
	@echo "    mod-tidy              - Tidy go modules"
	@echo "    lint                  - Run linter"
	@echo "    install-linter        - Install linter"
	@echo "    install-protoc-plugins - Install protoc plugins"
	@echo "    install-buf           - Install Buf"
	@echo ""
	@echo "  Docker:"
	@echo "    services-up           - Start all services"
	@echo "    services-down         - Stop all services"
	@echo "    services-logs         - Show logs for all services"
	@echo ""
	@echo "  Utilities:"
	@echo "    up                    - Start infrastructure and services"
	@echo "    down                  - Stop infrastructure and services"
	@echo "    logs                  - Show logs"
	@echo "    restart               - Restart everything"
	@echo "    status                - Show container status"
