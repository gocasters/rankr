# rankr
Rank repo contributors


---
## 🚀 Quick Start

For detailed setup instructions, see [SETUP.md](SETUP.md).

### Quick Setup Commands:
```bash
# 1. Install tools (ASDF)
asdf install

# 2. Setup protobuf (Buf)
make proto-setup

# 3. Setup Go dependencies
make mod-tidy

# 4. Start development environment
cd deploy/leaderboardscoring/development
docker-compose -f docker-compose.no-service.yml up -d
```

## 📋 Development Setup

This project uses ASDF to manage tool versions. For complete setup instructions, see [SETUP.md](SETUP.md).

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
git clone https://github.com/your-org/rankr.git
cd rankr
asdf install  # Install tools
make proto-setup  # Setup protobuf
make mod-tidy  # Setup Go dependencies
```

## 🔧 Protobuf Development

This project uses [Buf](https://buf.build/) for protobuf management. Here are the available commands:

### Setup Commands
```bash
# Basic setup (install Buf, initialize config, lint)
make proto-setup

# Complete setup with code generation
make proto-setup-full

# Install Buf CLI tool
make install-buf

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

---