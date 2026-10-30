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

# 4. Start shared infrastructure (PostgreSQL, Redis, NATS, ...)
make infra-up

# 5. Start the services you need, one by one
make start-auth-app-dev
make start-project-app-dev
```

To start every service at once, use `make services-up` (heavy on a laptop).

New here? Read [AGENTS.md](AGENTS.md) and [docs/agentic/WORKFLOW.md](docs/agentic/WORKFLOW.md) before your first PR.

### Stop and inspect

```bash
make stop-auth-app-dev    # stop one service
make infra-logs           # follow infrastructure logs
make infra-down           # stop shared infrastructure
```

### Webhook development

The webhook service requires an HMAC secret at startup. Generate a local-only
secret and save it in the Compose environment file:

```bash
openssl rand -hex 32

# Create this ignored file and paste the generated value after `=`.
printf 'GITHUB_WEBHOOK_SECRET=replace-with-the-generated-value\n' \
  > deploy/webhook/development/.env
```

Never commit the generated secret or any `.env` file. Docker Compose loads
`deploy/webhook/development/.env` when starting the webhook service.

Start the shared infrastructure and webhook service with:

```bash
make infra-up
make start-webhook-app-dev-log
```

`make start-webhook-app-dev`, `make services-up`, and `make up` also require
`GITHUB_WEBHOOK_SECRET`. Any GitHub webhook sender or simulator must sign its
payloads with the same value.

The repository does not currently provide a supported webhook simulator or a
host-exposed development webhook endpoint. Use the package tests for webhook
verification changes until that development workflow is added:

```bash
go test -short ./webhookapp/...
make check
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
