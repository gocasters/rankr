# Buf Protobuf Setup Documentation

## Overview

This document provides comprehensive documentation for the Buf setup in the Rankr project. The project uses Buf v2 for Protocol Buffer management, code generation, and API validation.

## Table of Contents

- [Project Structure](#project-structure)
- [Configuration Files](#configuration-files)
- [Rules and Validation](#rules-and-validation)
- [Plugin Configuration](#plugin-configuration)
- [Code Generation](#code-generation)
- [Buf v2 Features](#buf-v2-features)
- [Usage Examples](#usage-examples)
- [Best Practices](#best-practices)
- [Troubleshooting](#troubleshooting)

## Project Structure

```
rankr/
├── doc/                      # Documentation
│   └── BUFBUILD.md           # This documentation
├── protobuf/                 # Protocol Buffer definitions
│   ├── buf.yaml              # Module configuration
│   ├── event/
│   │   ├── v1/
│   │   └── event.proto       # Event definitions
│   └── golang/               # Generated Go code
│       └── event/
│           └── event.pb.go
├── buf.gen.yaml              # Code generation configuration
├── buf.yaml                  # Workspace configuration
└── Makefile                  # Build automation
```

## Configuration Files

### Workspace Configuration (`buf.yaml`)

```yaml
version: v2
modules:
  - path: protobuf
```

This file defines the workspace and its modules. In v2, workspaces replace the old `buf.work.yaml` files and provide a unified configuration for multi-module projects.

### Module Configuration (`protobuf/buf.yaml`)

```yaml
version: v2
modules:
  - path: .
lint:
  use:
    - STANDARD
  except:
    - PACKAGE_VERSION_SUFFIX
breaking:
  use:
    - FILE
  except:
    - WIRE_JSON
```

### Code Generation Configuration (`buf.gen.yaml`)

```yaml
version: v2
inputs:
  - directory: protobuf
managed:
  enabled: true
  override:
    - file_option: go_package_prefix
      value: github.com/gocasters/rankr/protobuf/golang
plugins:
  - remote: buf.build/protocolbuffers/go:v1.34.2
    out: protobuf/golang
    opt: paths=source_relative
  - remote: buf.build/grpc/go:v1.5.1
    out: protobuf/golang
    opt: paths=source_relative
```

## Rules and Validation

### Linting Rules

Buf provides comprehensive linting rules to ensure protobuf files follow best practices. Rules are categorized and can be customized per module.

#### Available Rule Categories

- **STANDARD**: Default rules for most projects
- **MINIMAL**: Basic rules for compatibility
- **ALL**: All available rules

#### Common Linting Rules

| Rule | Description | Category |
|------|-------------|----------|
| `PACKAGE_VERSION_SUFFIX` | Package names should have version suffixes | STANDARD |
| `PACKAGE_NO_IMPORT_CYCLE` | No circular imports between packages | STANDARD |
| `ENUM_VALUE_UPPER_SNAKE_CASE` | Enum values should be UPPER_SNAKE_CASE | STANDARD |
| `MESSAGE_PASCAL_CASE` | Message names should be PascalCase | STANDARD |
| `FIELD_LOWER_SNAKE_CASE` | Field names should be lower_snake_case | STANDARD |

#### Customizing Rules

```yaml
# In protobuf/buf.yaml
lint:
  use:
    - STANDARD          # Use all standard rules
  except:
    - PACKAGE_VERSION_SUFFIX  # Except this specific rule
  rules:
    - ENUM_ZERO_VALUE_SUFFIX:  # Override specific rule
        suffix: "_UNSPECIFIED"
```

### Breaking Change Rules

Breaking change detection ensures API compatibility when updating protobuf definitions.

#### Available Breaking Rules

- **FILE**: File-level breaking changes
- **PACKAGE**: Package-level breaking changes
- **WIRE**: Wire-level breaking changes
- **WIRE_JSON**: Wire + JSON compatibility

#### Current Configuration

```yaml
breaking:
  use:
    - FILE
  except:
    - WIRE_JSON
```

## Plugin Configuration

### Plugin Types

Buf v2 supports three types of plugins:

1. **Remote Plugins** (`remote:`): Hosted on Buf Schema Registry
2. **Local Plugins** (`local:`): Installed locally on your system
3. **Built-in Plugins** (`protoc_builtin:`): Built into protoc

### Current Plugin Setup

#### Protocol Buffers Go Plugin

```yaml
- remote: buf.build/protocolbuffers/go:v1.34.2
  out: protobuf/golang
  opt: paths=source_relative
```

#### gRPC Go Plugin

```yaml
- remote: buf.build/grpc/go:v1.5.1
  out: protobuf/golang
  opt: paths=source_relative
```

### Plugin Options

- `out`: Output directory for generated files
- `opt`: Plugin-specific options
- `path`: Custom plugin path (for local plugins)

### Using Local Plugins

If you prefer local plugins instead of remote ones:

```yaml
plugins:
  - local: protoc-gen-go
    out: protobuf/golang
    opt: paths=source_relative
  - local: protoc-gen-go-grpc
    out: protobuf/golang
    opt: paths=source_relative
```

## Code Generation

### Managed Mode

Managed mode automatically handles common protobuf options:

```yaml
managed:
  enabled: true
  override:
    - file_option: go_package_prefix
      value: github.com/gocasters/rankr/protobuf/golang
```

### Generating Code

```bash
# Generate code for all configured inputs
buf generate

# Generate code with verbose output
buf generate --verbose

# Generate code for specific files
buf generate --path protobuf/event/event.proto
```

### Input Specification

Specify which protobuf files to generate code for:

```yaml
inputs:
  - directory: protobuf          # All .proto files in directory
  - file: protobuf/event/event.proto  # Specific file
  - module: buf.build/googleapis/googleapis  # External module
```

## Buf v2 Features

### Enhanced Module Management

- **Unified Configuration**: Single `buf.yaml` for workspace and modules
- **Multi-Module Support**: Handle complex projects with multiple modules
- **Dependency Management**: Shared dependencies across modules

### Improved Code Generation

- **Input Flexibility**: Specify inputs directly in configuration
- **Plugin Types**: Explicit remote/local/built-in plugin support
- **Advanced Managed Mode**: More control with `override` and `disable`

### Better Performance

- **Parallel Processing**: Faster builds with better resource utilization
- **Smart Caching**: Improved caching of plugins and dependencies
- **Optimized Generation**: Faster code generation process

### Developer Experience

- **Clear Error Messages**: More detailed and actionable errors
- **Enhanced CLI**: Better command-line interface
- **Migration Tools**: Automated migration from v1 to v2

### Enterprise Features

- **Multi-Module Push**: Push modules in dependency order
- **Registry Integration**: Better Buf Schema Registry support
- **Team Collaboration**: Improved workflows for teams

## Usage Examples

### Basic Workflow

```bash
# Lint protobuf files
buf lint

# Check for breaking changes
buf breaking --against '.git#branch=main'

# Generate code
buf generate

# Build the module
buf build
```

### Advanced Usage

```bash
# Lint with specific configuration
buf lint --config protobuf/buf.yaml

# Generate code for specific plugin
buf generate --path protobuf/event/event.proto

# Push to Buf Schema Registry
buf push
```

## Best Practices

### File Organization

- Keep protobuf files in dedicated directories
- Use version suffixes for packages (`event.v1`, `user.v1`)
- Group related messages in the same file
- Use clear, descriptive names

### Rule Configuration

- Start with `STANDARD` rules
- Gradually add stricter rules as your project matures
- Document exceptions with comments
- Regularly review and update rules

### Plugin Management

- Use remote plugins for consistency
- Pin plugin versions for reproducible builds
- Keep plugin configurations in version control
- Test plugin updates in development before production

### Version Control

- Commit generated code to version control
- Use `.gitignore` for temporary files
- Tag releases with semantic versioning
- Document breaking changes

### Getting Help

- [Buf Documentation](https://buf.build/docs)
- [Buf GitHub Issues](https://github.com/bufbuild/buf/issues)
- [Protocol Buffers Style Guide](https://developers.google.com/protocol-buffers/docs/style)

---

## Quick Reference

### Essential Commands

```bash
# Development workflow
buf lint                    # Lint protobuf files
buf generate               # Generate code
buf build                  # Build the module

# Validation
buf breaking --against '.git#branch=main'  # Check breaking changes

# Registry operations
buf push                   # Push to Buf Schema Registry
buf dep update            # Update dependencies
```

### Configuration Structure

```yaml
buf.yaml (workspace)
├── buf.gen.yaml (code generation)
└── protobuf/
    ├── buf.yaml (module)
    └── *.proto (definitions)
```

This setup provides a robust, scalable foundation for Protocol Buffer development in the Rankr project, leveraging Buf v2's advanced features for better developer experience and maintainability.
