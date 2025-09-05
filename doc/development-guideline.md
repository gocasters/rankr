# Project Development Guidelines

## Philosophy & Core Principles

*   **Clarity over Cleverness:** Write code for the next developer. Avoid unnecessary complexity and "magic."
*   **Ownership:** You are responsible for your code's testing, documentation, and operational readiness (logging, metrics, etc.).
*   **Minimal Dependencies:** Evaluate the need for any external library critically. Dependencies are a long-term liability.

## Project Structure

We use a modular layout to enforce separation of concerns and improve navigability.

```
├── adapter/ # Tools which other services use
│   ├── redis
│   ├── nats
│   └── other-app-services
├── cli/  # Endpoints to test each app-service(presented in separeted folders)
├── cmd/  # External commands(same as app-service main.go, presented in separeted folders)
│   └── [app-service-name]/      # e.g., project, webhook, leaderboardscoring
│       ├── command/             # All command implementations
│       │    ├── root.go          # Root command definition
│       │    ├── serve.go         # Serve command (start service)
│       │    ├── migrate.go       # Migration commands (if DB required)
│       │    └── [other_commands].go
│       └── main.go              # Entry point (minimal, calls command package)
│ 
├── config/ # Project configuration management
├── deploy/ # Dockerfiles, Kubernetes manifests, Terraform
│   ├── [service-name1]/ # Deployment config of serveice 1 
│   └── [service-name2]/ # 
├── 
│
├── [service-name1]app/ # domain name concatinated with `app` postfix 
│       ├── delivery/ # Transport Layer
│       │   ├──  http (grps, cli, graphQl) # different presentation of app to the clients
│       │   │    ├── server.go
│       │   │    ├── health_check.go
│       │   │    ├── handler.go
│       │   │    └── 
│       │   └──  middleware/ #
│       │        ├── ...
│       │        └── ...
│       ├── repository # Data access layer (interface + implementation)
│       │   ├──  migrations #
│       │   │    ├── [migration-file-1].sql
│       │   │    └── [migration-file-2].sql
│       │   ├── dbconfig.yml
│       │   ├── [model1].go # Domain-specific structs with CRUD queries to db
│       │   └── [model2].go # 
│       ├── service # Business logic layer
│       │   ├── [related-entitiy1]/
│       │   │    ├── config.go
│       │   │    ├── entity.go
│       │   │    ├── param.go  # input parameters of usecase layer
│       │   │    ├── service.go
│       │   │    ├── validator.go
│       │   │    └── ...
│       │   └── [related-entitiy2]/
│       │   
│       ├── app.go # 
│       └── config.go # 
│
├── [service-name2]app/
│       └── ...
│
├── protobuf
├── pkg/ # Public, importable library code, helper functions
│   ├── logger #
│   │   └── logger.go # global logger (slog is suggested)
│   └── err_msg  # (pakage name)
│       ├── err_msg.go
│       └── ...
├── type/ # General types
├── doc/ # Project documentation (including this file)
├── go.mod
├── go.sum
└── Makefile # Common commands for building, testing, running
```

## Project Domains
### 1. Webhook (receives repository events)     2-3-6
### 2. Realtime     1-5
### 3. LeaderboardScoring (Update contributor score)        2-4
### 4. LeaderboardStat (Serves Leaderboard statistics to clients)      3-5

#### Communication with LeaderboardScoring

The **LeaderboardStat** service depends on **LeaderboardScoring** for refined contribution data (scores, events, and normalized repository activities).
There are two supported communication models depending on the stage of the system:
---

**Guideline:**
- For now, adopt **Option 3 (shared DB, read-only replica)** for speed and simplicity.
- Plan migration toward **Option 1 (event-driven, independent DB)** once LeaderboardStat grows in complexity, or when schema evolution between services becomes a bottleneck.

### 5. Contributor/User (Profile/ Project Scores)       9-4-3-8
### 6. TaskManagement (Managing Tasks based on repo issues/PRs)     1-
### 7. Notification (Managing Notifications)        7-5
### 8. Project/Repository (issue/ contribution(PR, Review, etc...)) 5
### 9- Auth     5

# Idiomatic Patterns & Must-Dos
## CLI Architecture with Cobra

All services **must** implement a consistent Cobra-based CLI structure. This ensures uniformity in service management, configuration, and deployment across all domains. ```cmd``` contains the application entry points.
### Standard Structure
```
cmd/
└── [service-name]/          # e.g., project, webhook, leadrboardscoring
    ├── main.go              # Entry point (minimal, calls command package)
    └── command/             # All command implementations
        ├── root.go          # Root command definition
        ├── serve.go         # Serve command (start service)
        ├── migrate.go       # Migration commands (if DB required)
        └── [other_commands].go
```
### Required Commands

Every service should implement these standard commands:

**1 — ```serve```**
- Purpose: Start the service
- Flags:
  - ```--migrate-up``` - Run migrations before starting (if applicable)
  - ```--migrate-down``` - Rollback migrations before starting (if applicable)

**2 — ```migrate``` (for services with databases)**
- Purpose: Database migration management
- Flags:
  - ```--up``` - Run migrations up
  - ```--down``` - Run migrations down

### configuration Loading

All services must support multiple configuration sources in this order of precedence:

**1 — CLI Flags** - Highest priority

**2 — Environment Variables** - Prefixed with ```[SERVICE]_``` (e.g., ```LEADERBOARDSTAT_```)

**3 — YAML Config File** - Default fallback

**Usage Example**
```
# Start any service
go run cmd/[service-name]/main.go serve

# Start with migrations
go run cmd/[app-service-name]/main.go serve --migrate-up

# Run migrations separately
go run cmd/[app-service-name]/main.go migrate --up

# Show help
go run cmd/[app-service-name]/main.go --help
```
## Project Setup & main.go
## Logging

Rankr uses a structured logging system with Go's slog package and Lumberjack for log rotation. The logger is implemented as a singleton pattern for consistent application-wide logging, with automatic file rotation based on size and age constraints.

```
// Initialize logger with config
logger.Init(cfg.Logger)

// Get logger instance and use it
log, _ := logger.L()
log.Info("Service starting", "port", 8080, "environment", "development")
log.Error("Database connection failed", "error", err, "attempt", 3)
```

**Log rotation** is an automated process of managing log files to prevent them from growing indefinitely and consuming all available disk space. It involves:

**1 —Archiving** the current log file.

**2 —Creating** a new, empty log file for new entries.

**3 —Deleting** old log files after a certain period or number.

Without log rotation, a single log file (```service.log```) would just keep getting bigger and bigger. This leads to several major problems.
This is the logger configurations in ```deploy/[appservice]/development/config.yml```

```
logger:
  file_path: "log/[appservice]/service.log"   # The active log file
  file_max_size_in_mb: 10         # Rotate when the file reaches 10 MB
  file_max_age_in_days: 7         # Delete archived logs older than 7 days
```

Here’s what your ```log/``` directory might look like over time:
```
log/[appservice]/
├── service.log          # <-- Active file (0.5 MB). Currently being written to.
├── service-2025-01-30T15-04-05.log  # <-- Archived, 10 MB, 1 day old (kept)
├── service-2025-01-29T10-22-17.log  # <-- Archived, 10 MB, 2 days old (kept)
└── service-2025-01-22T09-51-01.log  # <-- Archived, 10 MB, 9 days old (DELETED by lumberjack!)
```
## Observability: OpenTelemetry (OTel)
## Error Handling


# Domain-Specific Best Practices
## System-Level Best Practices
(How services in this project should communicate & integrate across domains)

**1 — Protobuf for Event Messages**
- All domain events (e.g., GitHub webhook → Scoring, Task created → Notification) must be serialized with Protobuf before publishing.
- Reduces payload size, enforces explicit contracts between services.

**2 — Reliability Patterns**
- **Retry with backoff** for transient errors.
- **Transactional Outbox** for DB+event consistency.
- **Idempotency** keys to avoid duplicates.
- **Eventual Consistency** is the default assumption.

**3 — Domain Communications**

**Option 1 — Event-Driven Communication (preferred)** “Something happened — whoever cares, react!”
- Services publish events (e.g., UserCreated, ScoreUpdated) without knowing who consumes them.
- Other services subscribe to relevant events.
- Follows Choreography Saga pattern with retry policies instead of rollbacks

✅ Pros: loose coupling, independent evolution of services, scales well.
⚠️ Cons: requires event infrastructure and introduces eventual consistency.

**Option 2 — Message-Driven Communication** “Send this to Service B”
- A service sends a message to a specific recipient after updating its state.
- The sender must know the exact address of the receiver.
- Usually one receiver per message (point-to-point).
- If the receiver gets overloaded, it can use a queue for buffering

✅ Pros: Easy to trace message flow and debug issues.
⚠️ Cons: Adding new consumers requires code changes in producers.

**Option 3 — Shared Database**
- Normally, each service owns its DB.
- If read duplication is too expensive, a service may provide read-only access (ideally through a replica) to another service.

✅ Pros: simple, low infra cost, fast to implement.
⚠️ Cons: tight coupling between two services.

## Service-Level Best Practices
Each domain package should follow this pattern:
### Layer Responsibilities:
1 — ```[appService]/service/entity.go```
- Defines pure domain structs. No logic.

2 — ```[appService]/repository/``` (Data Layer):
- Defines the persistence interface (type Repository interface {...}).
- Provides the implementation (e.g., postgresRepository).

3 — ```[appService]/service/``` (Business Layer):
- Contains all business logic and rules.
- Depends on the Repository interface.
- Must remain free of HTTP/transport concerns.

4 — ```[appService]/delivery/``` (Transport Layer):
- Handles HTTP-specific tasks (JSON marshaling/unmarshaling, parameter parsing).
- Calls the Service layer.
- Instruments with logging and tracing.

### Additional Guidelines
- Validation and data sanitizing must happen before data enters the service layer. We uses ```ozzo-validation``` library as Golang is a ***strong type*** language and this library use this property perfectly.
- Keep entities and use cases free of external dependencies.
- Outer layers (delivery, repository) can depend on frameworks and libraries.
- Dependency Injection: Use constructor functions to explicitly require dependencies.

# Testing
# CI/CD & Git Practices
**Continuous Integration (CI)** means every code change is automatically tested, linted, and built before merging.
In Go, CI usually covers:

- **Dependency management** (download modules).
- **Linting** (```golangci-lint```) to enforce code quality & style.
- **Unit tests** with coverage reports (```go test ./... -cover```).
- **Build check** to ensure the app compiles.
  This guarantees that the ```main``` branch is always stable.

## GitHub Actions for CI
We use GitHub Actions as our CI/CD platform.
- Workflows are defined in ```.github/workflows/*.yml```.
- They are triggered automatically on:
  - Pushes to ```main```
  - Pull Requests targeting ```main```
## CI Workflow (```ci.yml```)
The CI pipeline runs on every PR and push to ```main``` and enforces the following checks:

**1 — Checkout & Setup**
- Checks out the repository code.
- Sets up Go (```1.25.x```) on the runner.
- Caches dependencies to speed up builds.

**2 — Dependencies**
- Runs ```go mod download``` to ensure all modules are installed.

**3 — Linting**
- Runs ```golangci-lint``` (```v2.4.0```) with a ```5m``` timeout.
- Ensures code follows project linting and formatting rules.

**4 — Testing**
- Runs ```go test ./... -v -coverprofile=coverage.out```.
- All tests must pass, and coverage is tracked.

**5 — Build**
- Runs ```go build -v ./...```.
- Ensures the project compiles without errors.
## Git Practices
- **All changes must go through Pull Requests (PRs)**.
- **CI must pass before merging** into ```main```.
- Developers should run ```go test ./...``` and ```golangci-lint run``` locally before pushing.
- ```main``` branch is always deployable.
---

### Styleguide
Make the code clear for readers by effective naming, helpful commentary, and efficient code organization.
- Don't describe what is the code actually doing.
- Describe why is the code doing what it does.

### TODO
- use [-TODO - the must be done part] in the case of Technical Dept or Logics which must be implemented but not yet.