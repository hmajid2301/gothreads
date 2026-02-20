# Architecture Research

**Domain:** Digital wardrobe management with AI integration
**Researched:** 2026-02-20
**Confidence:** HIGH

## Standard Architecture

### System Overview

```
┌─────────────────────────────────────────────────────────────┐
│                        Frontend Layer                       │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐        │
│  │ HTMX    │  │ Templ   │  │ Alpine  │  │ Svelte  │        │
│  │ Core    │  │Templates│  │ Simple  │  │ Canvas  │        │
│  │ UI      │  │         │  │ State   │  │ Complex │        │
│  └────┬────┘  └────┬────┘  └────┬────┘  └────┬────┘        │
│       │            │            │            │              │
├───────┴────────────┴────────────┴────────────┴──────────────┤
│                        API Gateway Layer                    │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────────────────────────────────────────────┐    │
│  │           HTTP Router + Middleware Stack            │    │
│  │         (Auth, Logging, Error Handling)             │    │
│  └─────────────────────────────────────────────────────┘    │
├─────────────────────────────────────────────────────────────┤
│                      Service Layer                          │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐     │
│  │ Wardrobe │  │ Outfit   │  │ AI Queue │  │ Analytics│     │
│  │ Service  │  │ Service  │  │ Manager  │  │ Service  │     │
│  └────┬─────┘  └────┬─────┘  └────┬─────┘  └────┬─────┘     │
│       │             │             │             │           │
├───────┴─────────────┴─────────────┴─────────────┴───────────┤
│                      Data Layer                             │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐                   │
│  │PostgreSQL│  │ S3 Store │  │AI Providers│                 │
│  │ Database │  │SeaweedFS │  │Local+Cloud │                 │
│  └──────────┘  └──────────┘  └──────────┘                   │
└─────────────────────────────────────────────────────────────┘
```

### Component Responsibilities

| Component | Responsibility | Typical Implementation |
|-----------|----------------|------------------------|
| HTMX Core UI | Server-driven HTML updates, forms, navigation | Progressive enhancement over standard HTML |
| Templ Templates | Type-safe server-side HTML generation | Go-native templating with compile-time safety |
| Alpine.js Simple State | Lightweight client-side reactivity (modals, dropdowns) | Minimal JavaScript framework for simple interactions |
| Svelte Canvas | Complex drag-drop outfit composition | Single-purpose component for visual outfit building |
| HTTP Router | Request routing, middleware orchestration | stdlib net/http with go-pkgz/routegroup |
| Service Layer | Business logic, external service coordination | Interface-based dependency injection pattern |
| PostgreSQL Database | Structured data persistence with ACID guarantees | pgx driver with sqlc-generated type-safe queries |
| S3 Store | Object storage for images (raw/processed) | SeaweedFS providing S3-compatible API |
| AI Providers | Background removal, tagging, virtual try-on | Pluggable provider pattern (local-first, cloud fallback) |

## Recommended Project Structure

```
gothreads/
├── cmd/                    # CLI utilities and main application
│   ├── gothreads/         # Main server application
│   ├── migrate/           # Database migration utility
│   └── scraper/           # Product page scraping tools
├── internal/               # Private application code
│   ├── service/           # Business logic layer
│   │   ├── wardrobe/      # Item and inventory management
│   │   ├── outfit/        # Outfit creation and management
│   │   ├── ai/            # AI processing coordination
│   │   └── analytics/     # Usage statistics and insights
│   ├── store/             # Data access layer
│   │   ├── db/            # PostgreSQL interactions
│   │   └── s3/            # Object storage interactions
│   ├── transport/         # HTTP transport layer
│   │   ├── handlers/      # HTTP handlers by feature
│   │   ├── middleware/    # Auth, logging, error handling
│   │   └── views/         # Templ templates
│   └── ai/                # AI provider implementations
│       ├── providers/     # Concrete AI service implementations
│       └── queue/         # Async job processing
├── pkg/                   # Public library code (minimal for self-hosted app)
├── web/                   # Frontend assets
│   ├── static/           # CSS, JS, icons (embedded with go:embed)
│   ├── templates/        # Templ source files
│   └── svelte/           # Svelte component source
├── configs/              # Configuration templates
├── deployments/          # Docker Compose, Kubernetes manifests
├── docs/                 # Documentation
├── scripts/              # Build and deployment scripts
└── tests/                # Test files (unit, integration, e2e)
```

### Structure Rationale

- **cmd/:** Multiple entry points for different utilities (server, migration, CLI tools)
- **internal/:** All application-specific code, preventing external imports
- **service/:** Clean separation of business logic from transport concerns
- **store/:** Data access abstraction enabling testing and potential backend swapping
- **transport/:** HTTP-specific concerns isolated from business logic
- **web/:** Frontend assets organized by type, with clear build pipeline

## Architectural Patterns

### Pattern 1: Hybrid Frontend (HTMX + Targeted JavaScript)

**What:** Server-driven HTML with minimal client-side JavaScript only where needed
**When to use:** When you want fast development, SEO-friendly content, and reduced complexity
**Trade-offs:** Less interactive than SPA, but much simpler to develop and maintain

**Example:**
```go
// Templ template with HTMX
templ ItemList(items []Item) {
    <div hx-get="/items?cursor=next" 
         hx-trigger="intersect" 
         hx-swap="afterend">
        for _, item := range items {
            @ItemCard(item)
        }
    </div>
}

// Alpine.js for simple interactivity
<div x-data="{ uploading: false }">
    <form hx-post="/items/upload" 
          @htmx:before-request="uploading = true"
          @htmx:after-request="uploading = false">
        <button :disabled="uploading">
            <span x-show="!uploading">Upload</span>
            <span x-show="uploading">Processing...</span>
        </button>
    </form>
</div>
```

### Pattern 2: AI Processing Queue with SSE Updates

**What:** Asynchronous AI job processing with real-time progress updates
**When to use:** For long-running AI operations that shouldn't block the user interface
**Trade-offs:** More complex than synchronous processing, but provides better user experience

**Example:**
```go
// Queue job for background processing
func (s *AIService) ProcessBackgroundRemoval(ctx context.Context, itemID uuid.UUID) error {
    job := &AIJob{
        ItemID: itemID,
        Type: "background_removal",
        Status: "queued",
    }
    return s.queue.Enqueue(job)
}

// Server-sent events for progress updates
func (h *AIHandler) JobStatus(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "text/event-stream")
    w.Header().Set("Cache-Control", "no-cache")
    
    for update := range h.aiService.Subscribe(jobID) {
        fmt.Fprintf(w, "data: %s\n\n", update.JSON())
        if f, ok := w.(http.Flusher); ok {
            f.Flush()
        }
    }
}
```

### Pattern 3: Provider Strategy for AI Services

**What:** Pluggable AI providers with fallback capabilities
**When to use:** When you want local-first AI with cloud fallbacks and easy provider switching
**Trade-offs:** More complex than single provider, but provides flexibility and reliability

**Example:**
```go
type BackgroundRemovalProvider interface {
    RemoveBackground(ctx context.Context, imageData []byte) ([]byte, error)
    IsHealthy() bool
    Priority() int
}

type AIProviderManager struct {
    providers map[string]BackgroundRemovalProvider
}

func (m *AIProviderManager) RemoveBackground(ctx context.Context, imageData []byte) ([]byte, error) {
    // Try providers in priority order
    for _, provider := range m.getSortedProviders() {
        if !provider.IsHealthy() {
            continue
        }
        
        result, err := provider.RemoveBackground(ctx, imageData)
        if err == nil {
            return result, nil
        }
        // Log error and try next provider
    }
    return nil, errors.New("all providers failed")
}
```

## Data Flow

### Request Flow

```
[User Action (Upload Item)]
    ↓
[HTMX Request] → [HTTP Handler] → [Service Layer] → [AI Queue]
    ↓                ↓                ↓               ↓
[HTML Response] ← [Template] ← [Business Logic] ← [Database]
    ↓
[SSE Connection] ← [AI Worker] ← [AI Provider] ← [Background Processing]
```

### State Management

```
[PostgreSQL Database]
    ↓ (queries)
[Service Layer] ←→ [AI Processing Queue] → [Background Workers]
    ↓ (templates)           ↓ (updates)
[Templ Views] ←─────── [Server-Sent Events]
    ↓
[HTMX Frontend] + [Alpine.js State] + [Svelte Canvas State]
```

### Key Data Flows

1. **Item Upload Flow:** User uploads → Immediate storage → Background AI processing → Real-time progress updates → Completion notification
2. **Outfit Creation Flow:** Load items via HTMX → Compose in Svelte canvas → Save positions via form submission → Persist to database
3. **Analytics Flow:** User actions → Event collection → Real-time aggregation → Chart rendering with ApexCharts

## Scaling Considerations

| Scale | Architecture Adjustments |
|-------|--------------------------|
| 1-10 users | Single server with Docker Compose, local AI processing only |
| 10-100 users | Add Redis for session storage, separate AI workers, cloud AI fallbacks |
| 100-1000 users | Separate database server, load balancer, CDN for images, horizontal AI workers |

### Scaling Priorities

1. **First bottleneck:** AI processing becomes slow - solution: horizontal AI workers, cloud provider integration
2. **Second bottleneck:** Database connections - solution: connection pooling optimization, read replicas for analytics

## Anti-Patterns

### Anti-Pattern 1: Heavy Client-Side Framework

**What people do:** Use React/Vue for everything including simple CRUD operations
**Why it's wrong:** Adds complexity, build steps, and maintenance overhead for minimal benefit in content-heavy applications
**Do this instead:** Use HTMX for 95% of interactions, JavaScript only for complex components like drag-drop canvas

### Anti-Pattern 2: Synchronous AI Processing

**What people do:** Block user interface while waiting for AI operations to complete
**Why it's wrong:** Creates poor user experience, timeouts, and resource contention
**Do this instead:** Queue AI operations with progress updates via Server-Sent Events

### Anti-Pattern 3: Single AI Provider Dependency

**What people do:** Hardcode integration with one AI service (e.g., only OpenAI)
**Why it's wrong:** Creates vendor lock-in, single point of failure, and limits deployment flexibility
**Do this instead:** Use provider interface with local-first approach and cloud fallbacks

## Integration Points

### External Services

| Service | Integration Pattern | Notes |
|---------|---------------------|-------|
| Ollama (Local AI) | REST API with health checks | Primary for privacy-focused deployments |
| OpenAI/Hugging Face | REST API with rate limiting | Fallback for better quality/speed |
| SeaweedFS | S3-compatible API | Local object storage with familiar interface |
| PostgreSQL | Connection pool with pgx | ACID transactions for consistency |

### Internal Boundaries

| Boundary | Communication | Notes |
|----------|---------------|-------|
| Frontend ↔ API | HTMX requests + SSE | Server-driven updates minimize state sync issues |
| Service ↔ Store | Interface-based calls | Enables testing with mocks, potential backend swaps |
| AI Queue ↔ Workers | Channel-based messaging | Go channels for type-safe, efficient job distribution |
| Workers ↔ Providers | Strategy pattern | Runtime provider selection based on availability/priority |

## Production Hardening

### Error Handling & Recovery

- **Graceful Degradation:** AI features continue working even if advanced providers fail
- **Circuit Breakers:** Prevent cascade failures when external services are down
- **Retry Logic:** Exponential backoff for transient failures (1min → 2min → 4min → 8min)
- **Error Monitoring:** Structured logging with correlation IDs for request tracing

### Security Considerations

- **Input Validation:** Image format/size validation before processing
- **CSRF Protection:** HTMX includes CSRF tokens automatically
- **Content Security Policy:** Strict CSP headers to prevent XSS
- **File Upload Security:** Virus scanning with ClamAV, EXIF data stripping

### Observability

- **Structured Logging:** JSON logs with request context using slog
- **Metrics Collection:** Prometheus metrics for API latency, AI processing times
- **Health Checks:** `/health` endpoint for container orchestration
- **Tracing:** OpenTelemetry for distributed request tracing

### Deployment Patterns

- **Docker Compose:** Primary deployment method for self-hosted environments
- **NixOS Module:** Declarative deployment option for Nix users
- **Rolling Updates:** Zero-downtime deployments with health checks
- **Configuration Management:** Environment-based config with validation

## Migration Path: Prototype → Production

### Phase 1: Foundation (Current State → Proper Architecture)
1. **Restructure codebase** according to standard Go project layout
2. **Implement service layer** interfaces and dependency injection
3. **Add comprehensive error handling** with structured logging
4. **Set up proper database migrations** with goose

### Phase 2: Production Readiness
1. **Add comprehensive testing** (unit, integration, E2E with Playwright)
2. **Implement AI provider abstraction** with local/cloud fallbacks
3. **Set up monitoring and observability** (metrics, tracing, alerting)
4. **Create deployment automation** (CI/CD pipeline, Docker images)

### Phase 3: Enhanced Features
1. **Implement async AI processing** with queue and progress updates
2. **Add PWA capabilities** (offline mode, background sync, notifications)
3. **Build analytics dashboard** with real-time charts
4. **Create sharing functionality** for outfit links

### Build Order Dependencies

1. **Core Infrastructure:** Database, migrations, service interfaces
2. **Business Logic:** Service implementations, AI provider framework
3. **Transport Layer:** HTTP handlers, middleware, templates
4. **AI Processing:** Queue implementation, worker processes
5. **Frontend Enhancement:** HTMX templates, Svelte components
6. **Production Features:** Monitoring, deployment automation

## Sources

- [Go Standard Project Layout](https://github.com/golang-standards/project-layout) - Industry standard Go project structure
- [Go Module Organization](https://go.dev/doc/modules/layout) - Official Go module layout guidance  
- [Docker Compose Production](https://docs.docker.com/compose/production/) - Production deployment patterns
- [HTMX Architecture Patterns](https://htmx.org/essays/) - Server-driven UI architecture
- Existing codebase analysis (gothreads prototype) - Current implementation patterns and pain points

---
*Architecture research for: Digital wardrobe management with AI integration*
*Researched: 2026-02-20*