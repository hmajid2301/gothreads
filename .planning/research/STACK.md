# Stack Research

**Domain:** Production-ready digital wardrobe management with AI integration
**Researched:** 2026-02-20
**Confidence:** MEDIUM

## Recommended Stack

### Core Technologies

| Technology | Version | Purpose | Why Recommended |
|------------|---------|---------|-----------------|
| Go | 1.26.0 | Backend API, web server | Latest stable with security updates, excellent concurrency for AI processing, mature ecosystem |
| PostgreSQL | 18.x | Primary database | ACID compliance for transaction integrity, excellent JSON support for flexible schemas, proven scalability |
| HTMX | 2.0.x | Frontend interactivity | Server-side rendering performance, reduces JS complexity, excellent for form-heavy applications |
| Templ | 0.3.x | Type-safe HTML templates | Compile-time safety, Go-native templating, perfect integration with HTMX workflows |
| Docker | Latest | Containerization | Industry standard for self-hosted applications, reproducible deployments |

### Supporting Libraries

| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| github.com/jackc/pgx/v5 | v5.8.0+ | PostgreSQL driver | Always - best performance and feature support for PostgreSQL |
| github.com/pressly/goose/v3 | v3.22.0+ | Database migrations | Production deployments requiring schema versioning |
| github.com/golang-migrate/migrate/v4 | v4.18.0+ | Alternative migration tool | When advanced migration features needed |
| github.com/aws/aws-sdk-go | v1.55.8+ | S3-compatible storage | Object storage for images and files |
| github.com/go-co-op/gocron/v2 | v2.15.0+ | Background job scheduling | AI processing queues, maintenance tasks |
| github.com/golang-jwt/jwt/v5 | v5.3.1+ | JWT authentication | Stateless auth for API endpoints |
| go.uber.org/zap | v1.28.0+ | Structured logging | Production logging with performance |
| github.com/kelseyhightower/envconfig | v1.4.0+ | Configuration management | 12-factor app configuration |
| github.com/testcontainers/testcontainers-go | v0.38.0+ | Integration testing | Database testing with real PostgreSQL |

### Development Tools

| Tool | Purpose | Notes |
|------|---------|-------|
| Air | Live reload in development | Faster development cycles, watches Go and templ files |
| SQLC | Type-safe SQL code generation | Eliminates ORM complexity, compile-time query validation |
| Taskfile | Build automation | Better than Make for Go projects, cross-platform |
| golangci-lint | Code quality | Essential for production code, enforces best practices |
| Nix Flakes | Reproducible dev environments | Ensures consistent builds across environments |

## Installation

```bash
# Core dependencies (via go.mod)
go mod init your-project
go get github.com/jackc/pgx/v5@v5.8.0
go get github.com/pressly/goose/v3@v3.22.0
go get github.com/a-h/templ@v0.3.977
go get github.com/go-co-op/gocron/v2@v2.15.0
go get github.com/aws/aws-sdk-go@v1.55.8
go get github.com/golang-jwt/jwt/v5@v5.3.1
go get go.uber.org/zap@v1.28.0

# Development dependencies
go install github.com/air-verse/air@latest
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
go install github.com/a-h/templ/cmd/templ@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

## Alternatives Considered

| Recommended | Alternative | When to Use Alternative |
|-------------|-------------|-------------------------|
| PostgreSQL 18 | MySQL 8.4 | When team has existing MySQL expertise or strict compatibility requirements |
| HTMX + Templ | React/Vue SPA | When team prioritizes client-side state management over server performance |
| Gin Framework | Standard net/http | Consider Gin if you need middleware ecosystem, but net/http is sufficient for most cases |
| SQLC | GORM ORM | Use GORM only if rapid prototyping is priority over performance and type safety |
| Zap logging | Standard log package | Standard log is acceptable for simple applications, Zap for production |

## What NOT to Use

| Avoid | Why | Use Instead |
|-------|-----|-------------|
| MinIO (newer versions) | Licensing restrictions for self-hosted | SeaweedFS or AWS SDK with S3-compatible storage |
| Echo Framework | Less active maintenance than Gin, over-engineered for simple cases | Standard net/http or Gin if middleware needed |
| GORM (for this use case) | Runtime reflection overhead impacts AI processing performance | SQLC for type-safe, performant queries |
| Gorilla Mux | Maintenance mode, slower than modern alternatives | Standard net/http ServeMux (Go 1.22+) |
| go-redis v8 | Older version, use v9+ | go-redis v9 or rueidis for better performance |

## Stack Patterns by Variant

**If prioritizing development speed over performance:**
- Use Gin framework for middleware ecosystem
- Consider GORM for rapid prototyping
- Accept some performance trade-offs for faster iteration

**If optimizing for self-hosted resource efficiency:**
- Use standard net/http (lightweight)
- Implement custom middleware only where needed
- Focus on SQLC + PostgreSQL for minimal overhead

**If requiring high-availability deployment:**
- Add Redis for session storage and caching
- Implement graceful shutdown patterns
- Use connection pooling and health checks

## Version Compatibility

| Package A | Compatible With | Notes |
|-----------|-----------------|-------|
| Go 1.26.0 | PostgreSQL 18.x | Excellent driver support via pgx/v5 |
| HTMX 2.0.x | Templ 0.3.x | Perfect integration for server-side rendering |
| SQLC latest | PostgreSQL 18.x | Full JSON and modern PostgreSQL feature support |
| Docker latest | Nix Flakes | Can co-exist, Nix for dev, Docker for production |

## Production Hardening Considerations

**Security:**
- Use pgx connection pooling with proper limits
- Implement proper JWT token rotation
- Configure CORS appropriately for self-hosted environments
- Use Templ's built-in XSS protection

**Performance:**
- Connection pooling: max 25 connections for typical self-hosted
- Use SQLC prepared statements for repeated queries
- Implement proper context timeouts for AI operations
- Consider read replicas for high-traffic read operations

**Reliability:**
- Graceful shutdown handling
- Health check endpoints for load balancers
- Proper error logging with structured fields
- Circuit breaker pattern for external AI services

**Self-hosted Specific:**
- Single binary deployment preferred over microservices
- Built-in admin interface for user management
- Configuration via environment variables
- Support for SQLite as alternative for smaller deployments

## Sources

- [Go 1.26 Release Notes](https://go.dev/doc/devel/release) — Latest version verification
- [PostgreSQL 18 Release Notes](https://www.postgresql.org/docs/18/release.html) — Database compatibility
- [HTMX Documentation](https://htmx.org/docs/) — Frontend architecture patterns
- [Templ Documentation](https://templ.guide) — Template safety and performance
- Production Go applications survey 2025 — Library adoption patterns

---
*Stack research for: Digital wardrobe management with AI integration*
*Researched: 2026-02-20*