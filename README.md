# rankr
Rank repo contributors

See [the architecture overview](docs/architecture.md) for the Stage v1 service
map, ports, data stores, event flow, and authentication flow.


---
## 🚀 Quick Start

**Prerequisites:** Go 1.25.3+, Docker (running), Make, Git.

```bash
# 1. Clone and enter the repo
git clone https://github.com/gocasters/rankr.git
cd rankr

# 2. Download Go dependencies
go mod download

# 3. Prepare environment files
cp deploy/infrastructure/postgresql/development/.env-example deploy/infrastructure/postgresql/development/.env
cp deploy/task/development/.env.example deploy/task/development/.env

# 4a. Stage v1 stack (recommended): shared infra + six v1 apps + dev gateway
make up-v1

# — or —

# 4b. Full dev stack: all shared infra (including EMQX) plus every application service
make infra-up
make services-up

# 4c. Start only the services you need, one by one (after make infra-up)
make start-auth-app-dev
make start-project-app-dev
```

`make up-v1` brings up the base Docker network, PostgreSQL, Redis, NATS, OpenResty (dev gateway on port 80), and the six Stage v1 services: auth, contributor, leaderboardscoring, leaderboardstat, project, and webhook. It skips EMQX and non–v1 apps (notification, realtime, task, userprofile).

**Behavior:** each v1 service is started in order and Compose waits up to **`DEV_WAIT_TIMEOUT` seconds** (default **20**) for its **healthcheck** to pass. If any service fails that window, `up-v1` still starts the rest, then **exits with status 1** and prints a warning instead of “all services are up”. On a cold machine or first compile, raise the budget, e.g. `DEV_WAIT_TIMEOUT=120 make up-v1`. After you change **`deploy/development/Dockerfile`**, **`DEFAULT_AIR_VERSION` / Air args in the Makefile**, or **`go.mod`**, run **`FORCE=1 make docker-build-dev`** once so the shared `rankr-go-dev` image is rebuilt (by default `docker-build-dev` skips the build when the image already exists). GNU Make reserves **`-f`** for the makefile path, so we use **`FORCE=1`** instead of a `-f` flag.

To start every application service at once without using `up-v1`, use `make services-up` (heavy on a laptop).

New here? Read [AGENTS.md](AGENTS.md) and [docs/agentic/WORKFLOW.md](docs/agentic/WORKFLOW.md) before your first PR.

### Stop and inspect

```bash
make stop-auth-app-dev    # stop one service
make down                 # stop all dev application services and shared infrastructure
make logs                 # follow shared infrastructure logs (via infra-logs)
make infra-logs           # same as above
```

## 📋 Development Setup

ASDF is **optional**. Any way of installing Go 1.25.3+ works. Use ASDF if you want the exact versions in `.tool-versions`.

### Prerequisites
- **Go 1.25.3+** (via ASDF, Homebrew, or go.dev)
- **Git**
- **Docker** (for local development)

### 1. Install ASDF

Run these commands in your terminal:

```bash 
# Install ASDF
git clone https://github.com/asdf-vm/asdf.git ~/.asdf --branch v0.14.0

# Add to shell (choose one based on your shell)
echo -e '\n. "$HOME/.asdf/asdf.sh"' >> ~/.bashrc  # Bash
echo -e '\n. "$HOME/.asdf/asdf.sh"' >> ~/.zshrc   # Zsh
source ~/.bashrc  # or restart terminal

# Verify installation
asdf --version  # Should print "v0.14.0" or higher
```

### 2. Install Plugins & Tools

This project requires:
```bash 
# Add plugins
asdf plugin add golang https://github.com/asdf-community/asdf-golang.git

# Install versions from .tool-versions
asdf install
```

Verify installations:
```bash
go version
```

### 3. Project Setup

After installing the tools, follow the [Quick Start](#-quick-start) steps above.

### 4. BSR Setup (For Contributors)

If you want to contribute to protobuf schemas or use BSR features:

```bash
# Login to BSR (first time only)
make proto-bsr-login

# Verify you can access the module
make proto-bsr-info

# Pull latest schemas from BSR
make proto-deps
```

## 🔧 Protobuf Development

This project uses [Buf](https://buf.build/) for protobuf management. Here are the available commands:

### Setup Commands
```bash
# Basic setup (install Buf, initialize config, lint)
make proto-setup

# Complete setup with code generation
make proto-setup-full

# Install Buf CLI tool (pinned version v1.56.0)
make install-buf

# Force reinstall Buf CLI tool
make install-buf-force

# Install protoc plugins (required for code generation)
make install-protoc-plugins
```

### Development Commands
```bash
# Generate Go code from protobuf files
make proto-gen

# Lint protobuf files for style and best practices
make proto-lint

# Check for breaking changes against main branch
make proto-breaking

# Format protobuf files
make proto-format

# Update protobuf dependencies
make proto-deps

# Run all protobuf validations (lint + breaking changes)
make proto-validate

# Clean generated protobuf files
make proto-clean
```

### Workflow
1. **Initial Setup**: Run `make proto-setup-full` once to set up everything
2. **Development**: Use `make proto-gen` to generate code after changing `.proto` files
3. **Validation**: Use `make proto-validate` before committing to check for issues
4. **Formatting**: Use `make proto-format` to ensure consistent formatting

### BSR (Buf Schema Registry) Integration

This project is integrated with BSR for centralized protobuf schema management.

#### For Different User Types:

**🔍 Just Using the Project (No BSR needed):**
- Nothing to do: generated Go code is already committed. Follow the Quick Start.
- All protobuf files are included in the repository
- You can generate code locally without BSR

**👨‍💻 Contributing to Protobuf Schemas:**
```bash
# Login to BSR (first time only)
make proto-bsr-login

# Push schemas to BSR
make proto-bsr-push

# Check module information
make proto-bsr-info

# Verify login status
make proto-bsr-whoami
```

**🔄 Pulling Latest Schemas:**
```bash
# Update dependencies from BSR
make proto-deps

# Generate code from latest schemas
make proto-gen
```

**BSR Benefits:**
- Centralized schema storage and versioning
- Multi-language code generation
- Team collaboration on API design
- Automatic breaking change detection
- API documentation generation

Visit this module at: https://buf.build/gocasters/rankr

---