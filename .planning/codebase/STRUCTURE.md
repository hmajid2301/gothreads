# Codebase Structure

**Analysis Date:** 2026-02-20

## Directory Layout

```
gothreads/
├── backend/                # Database layer and future main app
│   └── store/             # Data persistence layer
├── mockups/               # Prototyping and current implementation
│   ├── api/               # Go API server (current main implementation)
│   ├── css/               # Stylesheet assets
│   ├── images/            # Static image assets
│   ├── js/                # Frontend JavaScript
│   ├── catvton/           # AI try-on service Docker setup
│   └── ladivton-shoes/    # Shoe try-on service setup
├── static/                # User-uploaded content
├── docs/                  # Documentation
├── nix/                   # Nix package management
└── .planning/             # Planning and architecture docs
```

## Directory Purposes

**backend/:**
- Purpose: Future main application structure (currently minimal)
- Contains: Database models, migrations, generated code
- Key files: `store/db/models.go` (sqlc-generated database types)

**mockups/api/:**
- Purpose: Current working HTTP API server implementation
- Contains: All business logic, handlers, AI integration
- Key files: `main.go` (entry point), `config.go` (configuration), `auth.go`, `db.go`, `s3.go`

**mockups/:**
- Purpose: Frontend prototype HTML pages and assets
- Contains: Static HTML files, CSS, JavaScript, Docker configurations
- Key files: `wardrobe.html` (main app), `outfit-builder.html`, `admin.html`

**static/:**
- Purpose: User-generated content storage
- Contains: Uploaded images, processed results
- Generated: Yes (user uploads)
- Committed: No

## Key File Locations

**Entry Points:**
- `mockups/api/main.go`: Main HTTP server and routing
- `Taskfile.yml`: Development task automation

**Configuration:**
- `config.yaml`: Application configuration
- `docker-compose.yml`: Service orchestration
- `flake.nix`: Nix package definition

**Core Logic:**
- `mockups/api/main.go`: AI job processing, HTTP handlers
- `mockups/api/auth.go`: Authentication and user management
- `mockups/api/db.go`: Database operations
- `backend/store/db/models.go`: Type-safe database models

**Testing:**
- Not detected: No test files found in current structure

## Naming Conventions

**Files:**
- Go files: `snake_case.go` (e.g., `main.go`, `config.go`)
- HTML files: `kebab-case.html` (e.g., `outfit-builder.html`)
- Directory names: `lowercase` or `kebab-case`

**Directories:**
- Module directories: `lowercase` (e.g., `backend`, `mockups`)
- Service directories: `kebab-case` (e.g., `ladivton-shoes`)

## Where to Add New Code

**New API Endpoint:**
- Primary code: `mockups/api/main.go` (add handler and route)
- Tests: No established testing pattern detected

**New Database Entity:**
- Implementation: `backend/store/db/migrations/` (new migration)
- Regenerate: Run `sqlc generate` to update `backend/store/db/models.go`

**New Frontend Page:**
- Implementation: `mockups/` (new HTML file)
- Assets: `mockups/css/` and `mockups/js/`

**New AI Feature:**
- Primary code: `mockups/api/main.go` (add job worker function)
- Configuration: Update `aiJobWorkers` map

**Utilities:**
- Shared helpers: `mockups/api/` (add to existing files or create new module files)

## Special Directories

**backend/store/db/:**
- Purpose: Generated database code and migrations
- Generated: Yes (sqlc generates models and query code)
- Committed: Yes (generated code is committed)

**mockups/api/vendor/:**
- Purpose: Go module dependencies (vendored)
- Generated: Yes (go mod vendor)
- Committed: Yes (all dependencies vendored)

**nix/:**
- Purpose: Nix package management and reproducible builds
- Generated: Partially (lock files are generated)
- Committed: Yes

**.planning/:**
- Purpose: Architecture documentation and planning artifacts
- Generated: Yes (by GSD tools)
- Committed: Yes

---

*Structure analysis: 2026-02-20*