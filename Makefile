.PHONY: start test build mod-tidy lint docker-build docker-run help

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
	@echo "  start     - Build and run locally"
	@echo "  test      - Run tests"
	@echo "  mod-tidy  - Clean up dependencies"
	@echo "  lint      - Run linters"
	@echo "  help      - Show this help"
	@echo "  build     - Compile binary (BINARY_NAME=$(BINARY_NAME), BUILD_DIR=$(BUILD_DIR))"
	@echo "  clean        - Remove build artifacts"
