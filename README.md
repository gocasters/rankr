# rankr
Rank repo contributors


---
## 🚀 Quick Start

### Quick Setup Commands
```bash
# 1. Install tools (ASDF)
asdf install

# 2. Setup protobuf (Buf)
make proto-setup

# 3. Setup Go dependencies
make mod-tidy

# 4. Start development environment
cd deploy/leaderboardscoring/development
docker compose -f docker-compose.no-service.yml up -d
```

## 📋 Development Setup

This project uses ASDF to manage tool versions

### Prerequisites
- **Go 1.21+** (managed via ASDF)
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

After cloning the repository:

```bash
git clone https://github.com/gocasters/rankr/
cd rankr
asdf install  # Install tools
make proto-setup  # Setup protobuf
make mod-tidy  # Setup Go dependencies
```

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
- Run `make proto-setup-full` and you're done!
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