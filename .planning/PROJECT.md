# go-threads: Privacy-First Digital Wardrobe Manager

## What This Is

go-threads is a self-hosted digital wardrobe management application that bridges manual inventory tracking with AI-powered styling capabilities. Users can upload clothing items, create outfits through an interactive visual canvas, track wear history, and get AI-powered suggestions - all while keeping their personal photos and fashion data completely private on their own servers.

## Core Value

Privacy-first wardrobe management with local AI processing - users can leverage powerful AI features like virtual try-on and automated styling without sending personal photos or shopping habits to the cloud.

## Requirements

### Validated

- ✓ Basic mockup frontend with HTML/CSS/JS working — existing prototype
- ✓ Go backend API structure established — backend/api/main.go framework
- ✓ Docker Compose development environment — PostgreSQL, SeaweedFS, Ollama
- ✓ AI integration patterns implemented — Ollama for vision models, rembg for background removal
- ✓ Database schema designed — PostgreSQL with SQLC code generation
- ✓ Nix development environment — flake.nix with devShell

### Active

- [ ] Complete backend service layer implementation (user, item, outfit, analytics services)
- [ ] Implement proper database migrations with goose
- [ ] Build HTMX-powered frontend templates using Templ
- [ ] Create Svelte outfit canvas component for drag-drop functionality  
- [ ] Implement AI processing queue with SSE progress updates
- [ ] Add OAuth2/OIDC authentication with configurable providers
- [ ] Build PWA features (offline mode, push notifications, installable)
- [ ] Implement sharing system for public outfit links
- [ ] Add comprehensive testing (unit, integration, E2E with Playwright)
- [ ] Create CI/CD pipeline with GitLab CI
- [ ] Build NixOS module for declarative deployment
- [ ] Fix virtual try-on reliability issues (GPU quota/local optimization)
- [ ] Implement outfit segmentation from full photos
- [ ] Add voice input for natural wardrobe interaction

### Out of Scope

- Multi-tenancy — single-user/family focused for simplicity
- Mobile native apps — PWA provides mobile experience
- Cloud-first AI — local processing is core value proposition
- Real-time chat/social features — focused on personal wardrobe management
- E-commerce integration — inventory management, not shopping platform

## Context

**Existing Codebase:**
- "Vibe-coded" mockup demonstrating all major features working
- Go API server with AI integration (Ollama, rembg, virtual try-on)
- HTML mockups showing UX flow for all major features
- Docker Compose stack with all dependencies (PostgreSQL, SeaweedFS, AI services)
- Nix flake providing reproducible development environment

**Technical Environment:**
- Go 1.24.0 backend with net/http and pgx/PostgreSQL
- HTMX + Templ for server-rendered UI, Svelte for complex interactions
- Local-first AI via Ollama (llava models) with cloud fallback options
- S3-compatible storage via SeaweedFS for image management
- Task-based build system with air for live reload

**Inspiration:**
- Stylebook: calendar planning, cost-per-wear tracking, wear history
- Whering: AI outfit suggestions, seamless web scraping
- Twelve70: weather-based suggestions, menswear-focused UI
- Acloset: AI background removal, professional item catalogs

**Current Issues:**
- Virtual try-on: HuggingFace quota limits, slow local fallback (~5 min)
- Backend: prototype code needs proper architecture/error handling
- Testing: minimal test coverage, needs comprehensive test suite
- Frontend: mockup HTML needs conversion to proper Templ templates

## Constraints

- **Privacy**: All AI processing must be capable of running locally by default
- **Self-hosted**: Must be deployable on single server/NAS with reasonable resources
- **Tech Stack**: Go backend, PostgreSQL, HTMX+Templ frontend, Svelte for complex UI
- **Development**: Nix-based development environment for reproducibility
- **Performance**: Must work on CPU-only systems (GPU acceleration optional)
- **Platform**: Docker Compose primary deployment, NixOS module secondary

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Hybrid Frontend (HTMX + Svelte) | HTMX for 95% of UI (server-rendered, simple), Svelte only for complex canvas interactions | — Pending |
| Local-first AI with Cloud Fallback | Privacy-first but allow cloud APIs for better performance/accuracy | — Pending |
| SeaweedFS over MinIO | S3-compatible storage, fully open source vs MinIO's licensing restrictions | — Pending |  
| SQLC over ORM | Type-safe SQL with explicit queries vs magic/reflection of ORMs | — Pending |
| OAuth2/OIDC Auth | Integrate with existing self-hosted auth (Authelia/Authentik) vs custom auth | — Pending |
| PWA over Native Mobile | Progressive Web App provides mobile experience without platform complexity | — Pending |

---
*Last updated: 2026-02-20 after codebase analysis and project initialization*