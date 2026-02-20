# Technology Stack

**Analysis Date:** 2026-02-20

## Languages

**Primary:**
- Go 1.24.0 - Backend API, web scraping, database operations

**Secondary:**
- SQL - Database schemas and queries via SQLC
- Python - ML/AI services (CATVTon virtual try-on)

## Runtime

**Environment:**
- Go 1.24.0
- Python 3.12 (for AI services)

**Package Manager:**
- Go modules
- Lockfile: `mockups/api/go.sum` present

## Frameworks

**Core:**
- Standard Go net/http - HTTP server
- github.com/go-pkgz/routegroup v1.6.0 - HTTP routing
- github.com/jackc/pgx/v5 v5.8.0 - PostgreSQL driver

**Testing:**
- Standard Go testing (referenced in Taskfile)
- gotestsum - Test runner

**Build/Dev:**
- Air - Live reload development server (config: `mockups/api/.air.toml`)
- Task/Taskfile.yml v3 - Build automation
- SQLC v2 - SQL code generation from queries
- Nix Flakes - Declarative builds and development environment

## Key Dependencies

**Critical:**
- github.com/jackc/pgx/v5 v5.8.0 - Primary database driver
- github.com/aws/aws-sdk-go v1.55.8 - S3-compatible storage client
- github.com/golang-jwt/jwt/v5 v5.3.1 - JWT authentication tokens

**Infrastructure:**
- github.com/gocolly/colly/v2 v2.3.0 - Web scraping framework
- github.com/joho/godotenv v1.5.1 - Environment variable loading
- gopkg.in/yaml.v3 v3.0.1 - Configuration parsing

## Configuration

**Environment:**
- direnv with Nix flakes (`use flake` in `.envrc`)
- YAML configuration (`config.yaml`)
- Environment variables override config values

**Build:**
- `flake.nix` - Nix development environment and builds
- `Taskfile.yml` - Task automation and workflows
- `sqlc.yaml` - Database code generation config

## Platform Requirements

**Development:**
- Nix with flakes support
- Docker and Docker Compose
- Go 1.24+

**Production:**
- Docker containers
- PostgreSQL 18
- S3-compatible storage (SeaweedFS in development)

---

*Stack analysis: 2026-02-20*