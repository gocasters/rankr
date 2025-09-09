
.PHONY: start test build clean mod-tidy lint install-linter help \
        proto-setup proto-setup-full proto-gen proto-lint proto-breaking \
        proto-clean proto-format proto-deps proto-validate \
        install-protoc-plugins install-buf \
        docker-build docker-run \
        proto-bsr-push proto-bsr-push-create proto-bsr-info proto-bsr-login proto-bsr-whoami \
        update-buf-version


BINARY_NAME ?= rankr
BUILD_DIR ?= bin
BUF_VERSION ?= v1.56.0
DEFAULT_BRANCH ?= main
PROTOC_GEN_GO_VERSION ?= v1.34.2
PROTOC_GEN_GO_GRPC_VERSION ?= v1.5.1

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

# Docker targets
docker-build:
	@echo "Building Docker image..."
	docker build -t rankr:latest -f deploy/leaderboardscoring/development/Dockerfile .

docker-run:
	@echo "Running Docker container..."
	docker run -p 8080:8080 rankr:latest

# Protobuf targets
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
	@echo ""
	@echo "Note: To generate Go code, run: make proto-gen"
	@echo "      (This requires protoc plugins - run: make install-protoc-plugins)"

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
		echo "To enable breaking change detection, initialize Git: git init"; \
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
		echo "To enable breaking change detection, initialize Git: git init"; \
	fi

# BSR (Buf Schema Registry) targets
proto-bsr-push:
	@echo "Pushing protobuf module to BSR..."
	cd protobuf && buf push

proto-bsr-push-create:
	@echo "Pushing protobuf module to BSR (create if not exists)..."
	cd protobuf && buf push --create

proto-bsr-info:
	@echo "Getting BSR module information..."
	buf registry module info buf.build/gocasters/rankr

proto-bsr-login:
	@echo "Logging in to BSR..."
	buf registry login

proto-bsr-whoami:
	@echo "Checking BSR login status..."
	buf registry whoami

update-buf-version:
	@echo "Current Buf version: $(BUF_VERSION)"
	@echo "To update Buf version, run:"
	@echo "  make install-buf-force BUF_VERSION=$(BUF_VERSION)"
	@echo "  # or edit the Makefile and change BUF_VERSION variable"

install-protoc-plugins:
	@echo "Installing protoc plugins..."
	@echo "Note: This may take a while due to Go version compatibility..."
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
			echo "Buf version mismatch. Expected $(BUF_VERSION), found $$CURRENT_VERSION"; \
			echo "Reinstalling Buf $(BUF_VERSION)..."; \
			$(MAKE) install-buf-force; \
		fi; \
	else \
		echo "Buf not found, installing $(BUF_VERSION)..."; \
		$(MAKE) install-buf-force; \
	fi

install-buf-force:
	@echo "Installing Buf $(BUF_VERSION)..."
	@if command -v curl &> /dev/null; then \
		echo "Using curl to install Buf $(BUF_VERSION)..."; \
		curl -sSL "https://github.com/bufbuild/buf/releases/$(BUF_VERSION)/download/buf-$$(uname -s)-$$(uname -m)" -o /tmp/buf; \
		chmod +x /tmp/buf; \
		if [ -w /usr/local/bin ]; then \
			mv /tmp/buf /usr/local/bin/buf; \
		else \
			sudo mv /tmp/buf /usr/local/bin/buf; \
		fi; \
		echo "Buf $(BUF_VERSION) installed successfully"; \
	elif command -v wget &> /dev/null; then \
		echo "Using wget to install Buf $(BUF_VERSION)..."; \
		wget -qO /tmp/buf "https://github.com/bufbuild/buf/releases/$(BUF_VERSION)/download/buf-$$(uname -s)-$$(uname -m)"; \
		chmod +x /tmp/buf; \
		if [ -w /usr/local/bin ]; then \
			mv /tmp/buf /usr/local/bin/buf; \
		else \
			sudo mv /tmp/buf /usr/local/bin/buf; \
		fi; \
		echo "Buf $(BUF_VERSION) installed successfully"; \
	else \
		echo "Neither curl nor wget found. Please install Buf manually:"; \
		echo "   Visit: https://docs.buf.build/installation"; \
		echo "   Or use: BUF_VERSION=$(BUF_VERSION) make install-buf"; \
		exit 1; \
	fi

help:
	@echo "Available targets:"
	@echo "  start          - Build and run locally"
	@echo "  test           - Run tests"
	@echo "  mod-tidy       - Clean up dependencies"
	@echo "  lint           - Run linters"
	@echo "  install-linter - Install golangci-lint"
	@echo "  build          - Compile binary (BINARY_NAME=$(BINARY_NAME), BUILD_DIR=$(BUILD_DIR))"
	@echo "  clean          - Remove build artifacts"
	@echo "  help      - Show this help"
	@echo "  clean     - Remove build artifacts"
	@echo ""
	@echo "Docker targets:"
	@echo "  docker-build - Build Docker image"
	@echo "  docker-run   - Run Docker container"
	@echo ""
	@echo "Protobuf targets:"
	@echo "  proto-setup    - Basic Buf setup (install, init, lint)"
	@echo "  proto-setup-full - Complete Buf setup with code generation"
	@echo "  proto-gen      - Generate Go code from protobuf files"
	@echo "  proto-lint     - Lint protobuf files"
	@echo "  proto-breaking - Check for breaking changes"
	@echo "  proto-clean    - Clean generated protobuf files"
	@echo "  proto-format   - Format protobuf files"
	@echo "  proto-deps     - Update protobuf dependencies"
	@echo "  proto-validate - Run all protobuf validations"
	@echo "  install-buf    - Install Buf CLI tool (version $(BUF_VERSION))"
	@echo "  install-buf-force - Force reinstall Buf CLI tool"
	@echo "  install-protoc-plugins - Install protoc plugins (may take time)"
	@echo ""
	@echo "BSR (Buf Schema Registry) targets:"
	@echo "  proto-bsr-push        - Push protobuf module to BSR"
	@echo "  proto-bsr-push-create - Push to BSR (create if not exists)"
	@echo "  proto-bsr-info        - Get BSR module information"
	@echo "  proto-bsr-login       - Login to BSR"
	@echo "  proto-bsr-whoami      - Check BSR login status"


start-contributor-app-dev:
	./deploy/docker-compose-dev.bash --profile contributor up -d

start-contributor-app-dev-log:
	./deploy/docker-compose-dev.bash --profile contributor up

start-contributor-debug: ## Start user in debug mode (local)
	./deploy/docker-compose-dev-user-local.bash up -d

start-contributor-debug-log: ## Start user in debug mode (local) with logs
	./deploy/docker-compose-dev-user-local.bash up

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