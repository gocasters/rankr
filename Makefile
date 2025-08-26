.PHONY: start test build mod-tidy lint docker-build docker-run help proto-gen proto-lint proto-breaking proto-clean proto-format proto-setup install-buf

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
		echo "Git repository found, checking against main branch..."; \
		buf breaking --against '.git#branch=main'; \
	else \
		echo "No Git repository found. Skipping breaking change check."; \
		echo "To enable breaking change detection, initialize Git: git init"; \
	fi

proto-clean:
	@echo "Cleaning generated protobuf files..."
	rm -rf protobuf/golang/eventpb/*.pb.go

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
		buf breaking --against '.git#branch=main'; \
	else \
		echo "No Git repository found. Skipping breaking change check."; \
		echo "To enable breaking change detection, initialize Git: git init"; \
	fi

install-protoc-plugins:
	@echo "Installing protoc plugins..."
	@echo "Note: This may take a while due to Go version compatibility..."
	@if ! command -v protoc-gen-go &> /dev/null; then \
		echo "Installing protoc-gen-go..."; \
		go install google.golang.org/protobuf/cmd/protoc-gen-go@latest; \
	else \
		echo "protoc-gen-go is already installed"; \
	fi
	@if ! command -v protoc-gen-go-grpc &> /dev/null; then \
		echo "Installing protoc-gen-go-grpc..."; \
		go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest; \
	else \
		echo "protoc-gen-go-grpc is already installed"; \
	fi

install-buf:
	@echo "Installing Buf..."
	@if command -v buf &> /dev/null && buf --version &> /dev/null; then \
		echo "Buf is already installed and working"; \
	else \
		echo "Buf not found or not working, installing..."; \
		if command -v curl &> /dev/null; then \
			echo "Using curl to install Buf..."; \
			curl -sSL "https://github.com/bufbuild/buf/releases/latest/download/buf-$$(uname -s)-$$(uname -m)" -o /tmp/buf; \
			chmod +x /tmp/buf; \
			sudo mv /tmp/buf /usr/local/bin/buf; \
			echo "Buf installed successfully"; \
		elif command -v wget &> /dev/null; then \
			echo "Using wget to install Buf..."; \
			wget -qO /tmp/buf "https://github.com/bufbuild/buf/releases/latest/download/buf-$$(uname -s)-$$(uname -m)"; \
			chmod +x /tmp/buf; \
			sudo mv /tmp/buf /usr/local/bin/buf; \
			echo "Buf installed successfully"; \
		else \
			echo "Neither curl nor wget found. Please install Buf manually:"; \
			echo "   Visit: https://docs.buf.build/installation"; \
			exit 1; \
		fi; \
	fi

help:
	@echo "Available targets:"
	@echo "  start     - Build and run locally"
	@echo "  test      - Run tests"
	@echo "  mod-tidy  - Clean up dependencies"
	@echo "  lint      - Run linters"
	@echo "  help      - Show this help"
	@echo "  build     - Compile binary (BINARY_NAME=$(BINARY_NAME), BUILD_DIR=$(BUILD_DIR))"
	@echo "  clean     - Remove build artifacts"
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
	@echo "  install-buf    - Install Buf CLI tool"
	@echo "  install-protoc-plugins - Install protoc plugins (may take time)"
