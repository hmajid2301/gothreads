# Architecture

**Analysis Date:** 2026-02-20

## Pattern Overview

**Overall:** Microservices + Frontend-Backend Separation with AI Integration

**Key Characteristics:**
- Go-based HTTP API server with mockup frontend prototypes
- External AI service integration (Ollama for local LLMs)
- Database abstraction through generated code (sqlc)
- Configuration-driven architecture with environment overrides
- Async job processing for AI operations

## Layers

**API Layer:**
- Purpose: HTTP endpoints for all application functionality
- Location: `mockups/api/main.go`
- Contains: REST handlers, middleware, routing logic
- Depends on: Database layer, AI services, storage services
- Used by: Frontend mockup pages, external clients

**Service Layer:**
- Purpose: Business logic and external service integration
- Location: `mockups/api/*.go` (distributed across multiple files)
- Contains: AI job processing, authentication, scraping, image processing
- Depends on: Database models, external APIs (Ollama, S3, rembg)
- Used by: API layer handlers

**Database Layer:**
- Purpose: Data persistence and query generation
- Location: `backend/store/db/`
- Contains: Generated sqlc code, models, migrations
- Depends on: PostgreSQL database
- Used by: Service layer

**External Services Layer:**
- Purpose: Integration with AI, storage, and image processing
- Location: Integration code spread across `mockups/api/` files
- Contains: Ollama AI calls, S3 storage, background removal, virtual try-on
- Depends on: External service availability
- Used by: Service layer for AI operations

## Data Flow

**AI Analysis Flow:**

1. Frontend submits image via `/api/ai/jobs` (POST)
2. Job queued in memory with async worker
3. Worker calls Ollama API with image and prompt
4. Response parsed and returned via Server-Sent Events
5. Frontend receives real-time updates

**User Management Flow:**

1. OAuth authentication via `/api/auth/{provider}`
2. User approval workflow for new accounts
3. Session management through middleware
4. Role-based access control (admin features)

**State Management:**
- In-memory job storage with mutex-based concurrency
- Database state through PostgreSQL with sqlc-generated queries
- Configuration state through YAML file with runtime updates

## Key Abstractions

**AIJob:**
- Purpose: Represents any asynchronous AI operation
- Examples: `mockups/api/main.go:244-255`
- Pattern: State machine with status transitions, event listeners

**Config:**
- Purpose: Application configuration management
- Examples: `mockups/api/config.go:21-29`
- Pattern: Hierarchical configuration with environment overrides

**Database Models:**
- Purpose: Type-safe database interactions
- Examples: `backend/store/db/models.go`
- Pattern: Generated Go structs from SQL schema

## Entry Points

**Main Application:**
- Location: `mockups/api/main.go:511`
- Triggers: Direct execution
- Responsibilities: Service initialization, routing setup, middleware configuration

**Task Runner:**
- Location: `Taskfile.yml`
- Triggers: Development commands
- Responsibilities: Build automation, database operations, service orchestration

## Error Handling

**Strategy:** Layered error handling with graceful degradation

**Patterns:**
- HTTP middleware for authentication/authorization errors
- Fallback model selection for AI operations (primary → fallback)
- Structured error responses with JSON format

## Cross-Cutting Concerns

**Logging:** Structured logging with contextual information (job IDs, operation types)
**Validation:** Input validation at API boundary with typed request/response structures  
**Authentication:** OAuth-based with session middleware and role-based access

---

*Architecture analysis: 2026-02-20*