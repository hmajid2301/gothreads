# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Project scaffolding with `backend/` and `frontend/` directory structure
- Nix flake environment with gomod2nix and **Bun** support
- Docker Compose stack (PostgreSQL, SeaweedFS, mock-oauth2, rembg)
- Task automation with Taskfile.yml (using Bun for frontend)
- direnv configuration for automatic environment loading
- Comprehensive README with quickstart guide
- Documentation structure in `docs/` and `docs/ai/`
- Development workflow documentation in `docs/GETTING_STARTED.md`
- **Consolidated task file**: `docs/ai/IMPLEMENTATION_TASKS.md` (replaces TODO.md, TASKS.md, .ai-tasks.json)
- CHANGELOG.md for tracking changes
- Hybrid frontend approach documented in docs/ARCHITECTURE.md
- **Complete mockup suite**: 8 throwaway HTML pages demonstrating UX
  - index.html (dashboard)
  - upload.html (file upload)
  - wardrobe.html (item grid)
  - item-detail.html (edit item)
  - outfit-builder.html ⭐ (drag-drop canvas)
  - outfits.html (outfit gallery)
  - analytics.html (charts and stats)
  - settings.html (preferences and data management)

### Planning
- Comprehensive project planning and architecture decisions documented in `docs/ARCHITECTURE.md`
- Product requirements and features in `docs/SPEC.md`
- Decided on **hybrid frontend stack**: HTMX + Alpine.js for most UI, minimal Svelte for outfit canvas only
- Decided on tech stack: Go, PostgreSQL, HTMX, Templ, Alpine.js, Svelte (canvas only), SeaweedFS
- Designed database schema for MVP (items, outfits, wear history, ratings, AI providers)
- Planned PWA features (offline mode, push notifications, background sync)
- Planned analytics dashboard with ApexCharts
- Planned AI integration strategy (background removal, auto-tagging, outfit scoring)
- Decided on trunk-based development with conventional commits
- Organized documentation into `docs/` directory

### Changed
- Moved from flat structure to `backend/` and `frontend/` separation
- Moved `PLANNING.md` to `docs/ARCHITECTURE.md`
- Moved `SPEC.md` to `docs/SPEC.md`
- Configured SeaweedFS instead of MinIO for S3 storage
- Set application port to 8080 (SeaweedFS uses 8080 for volume server)
- **Changed from full Svelte to hybrid HTMX + Alpine + minimal Svelte**
  - HTMX for 95% of UI (CRUD, navigation, forms)
  - Alpine.js for simple interactivity (dropdowns, modals)
  - Svelte only for outfit canvas component
- **Switched from pnpm to Bun** for frontend package management (faster, simpler)
- **Consolidated task management** - 3 separate files → 1 unified IMPLEMENTATION_TASKS.md

### Removed
- ClamAV from MVP (using file validation instead of virus scanning)
- TODO.md, TASKS.md, .ai-tasks.json (consolidated into docs/ai/IMPLEMENTATION_TASKS.md)

---

## Template for Future Entries

When implementing features, copy this template for each release:

```markdown
## [X.Y.Z] - YYYY-MM-DD

### Added
- New features

### Changed
- Changes in existing functionality

### Deprecated
- Soon-to-be removed features

### Removed
- Removed features

### Fixed
- Bug fixes

### Security
- Security improvements
```

---

## Example: What a real changelog entry will look like

```markdown
## [0.1.0] - 2026-02-15

### Added
- Initial project scaffolding with backend and frontend directories
- PostgreSQL database with initial schema (users, items, outfits)
- OAuth/OIDC authentication with Authelia/Authentik support
- Item CRUD operations (create, read, update, delete)
- Basic item listing with cursor-based pagination
- SeaweedFS S3 storage integration
- Image upload with EXIF stripping and virus scanning
- Background removal using rembg sidecar
- AI processing queue with retry logic
- Templ templates for item management UI
- Docker Compose setup for local development
- Nix flake for reproducible development environment
- Basic health check endpoint
- Structured logging with slog
- GitLab CI pipeline with linting and testing

### Changed
- N/A (initial release)

### Fixed
- N/A (initial release)

### Security
- Added CSP headers to prevent XSS
- Implemented virus scanning for uploads (ClamAV)
- EXIF stripping to protect user privacy
```
