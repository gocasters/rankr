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
│   ├── [app-service-name]/ # Application entry point(s)
│   └── main.go # main package and function
├── config/ # Project configuration management
├── deploy/ # Dockerfiles, Kubernetes manifests, Terraform
│   ├── [service-name1]/ # Deployment config of serveice 1 
│   └── [service-name2]/ # 
├── 
│
├── [service-name1]app/ # domain name concatinated with `app` postfix 
│       ├── delivery/ # Transport Layer
│       │   ├──  http (grps, cli, graphQl) # different presentation of app to the clients
│       │   │    ├── server.go # 
│       │   │    ├── health_check.go # 
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
│   │   └── logger.go # glbal logger (slog is suggested)
│   └── err_msg  # (pakage name)
│       ├── err_msg.go
│       └── ...
├── types/ # General types
├── docs/ # Project documentation (including this file)
├── go.mod
├── go.sum
└── Makefile # Common commands for building, testing, running
```

**Rationale:**

*   **cmd/** contains the application entry points.
*   **internal/** is Go's mechanism for preventing unwanted imports, enforcing clean boundaries.
*   **Domain-specific packages** (auth/, user/) within internal/ group all related code (DDD-style).

## Project Domains
### 1. Webhook (receives repository events)     2-3-6
### 2. Realtime     1-5
### 3. LeaderboardScoring (Update contributor score)        2-4
### 4. LeaderboardStats (Serves Leaderboard statistics to clients)      3-5

### Communication with LeaderboardScoring

The **LeaderboardStats** service depends on **LeaderboardScoring** for refined contribution data (scores, events, and normalized repository activities).  
There are two supported communication models depending on the stage of the system:

---

**Option 1 — Shared Database (Default in early stage / modular monolith)**
- **Scoring service is the sole writer** to the Leaderboard DB.
- **Stats service is a read-only consumer** of that same data.
- A **read replica** (primary–replica setup) is recommended to avoid load contention:
    - Primary DB → writes from Scoring only.
    - Replica DB → read-only connection for Stats.
- Stat is provisioned with a **read-only DB user** (SELECT privileges only).
- Any schema changes are owned by Scoring. Stats must adapt accordingly.

✅ Pros: simple, low infra cost, fast to implement.  
⚠️ Cons: couples Stats to Scoring’s schema; changes must be coordinated.

---

**Option 2 — Event-driven / Stats maintains its own copy (Target state in mature microservices)**
- Scoring service **publishes events** (e.g., via NATS/Kafka/Redis streams) whenever contributor scores or repository events are updated.
- Stats service **subscribes to these events** and maintains its own **read-optimized schema**.
- This allows Stats to serve leaderboard queries without depending on Scoring’s DB schema.

✅ Pros: loose coupling, independent evolution of services, scales well.  
⚠️ Cons: requires event infrastructure and introduces eventual consistency.

---

**Guideline:**
- For now, adopt **Option 1 (shared DB, read-only replica)** for speed and simplicity.
- Plan migration toward **Option 2 (event-driven, independent DB)** once Stats grows in complexity, or when schema evolution between services becomes a bottleneck.



### 5. Contributor/User (Profile/ Project Scores)       9-4-3-8
### 6. TaskManagement (Managing Tasks based on repo issues/PRs)     1-
### 7. Notification (Managing Notifications)        7-5
### 8. Project/Repository (issue/ contribution(PR, Review, etc...)) 5
### 9- Auth     5

# Idiomatic Patterns & Must-Dos
## Project Setup & main.go
## Logging
## Observability: OpenTelemetry (OTel)
## Error Handling

# Domain-Specific Best Practices
Each domain package (e.g., internal/user/) should follow this pattern:
### Layer Responsibilities:
**1 — [appService]/service/entity.go**
- Defines pure domain structs. No logic.

**2 — [appService]/repository/** (Data Layer):
- Defines the persistence interface (type Repository interface {...}).
- Provides the implementation (e.g., postgresRepository).

**3 — [appService]/service/** (Business Layer):
- Contains all business logic and rules.
- Depends on the Repository interface.
- Must remain free of HTTP/transport concerns.

**4 — [appService]/delivery/** (Transport Layer):
- Handles HTTP-specific tasks (JSON marshaling/unmarshaling, parameter parsing).
- Calls the Service layer.
- Instruments with logging and tracing.

### Dependency Injection:
Use constructor functions to explicitly require dependencies.
# Testing
# CI/CD & Git Practices
