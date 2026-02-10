# go-threads Project Handoff Document

**Generated**: 2026-02-08
**Status**: Planning complete, mockups ready, awaiting implementation

---

## Quick Context

**go-threads** is a privacy-first, self-hosted digital wardrobe manager. Users can:
- Upload clothing photos (AI removes backgrounds)
- Organize items in a digital wardrobe
- Create outfits by dragging items on a canvas
- Track wear frequency and cost-per-wear analytics
- Plan outfits on a calendar with weather integration
- Share outfits via public links

---

## Current State

### Completed
- [x] Project structure (`backend/`, `frontend/`, `docs/`, `mockups/`)
- [x] Nix flake with gomod2nix + Bun support
- [x] Docker Compose (PostgreSQL, SeaweedFS, mock-oauth2, rembg)
- [x] Taskfile.yml for development automation
- [x] Architecture documentation (docs/ARCHITECTURE.md)
- [x] Product specification (docs/SPEC.md)
- [x] Implementation task list (docs/ai/IMPLEMENTATION_TASKS.md)
- [x] **11 interactive HTML mockups** demonstrating full UX

### Not Started
- [ ] Go backend code
- [ ] Database migrations
- [ ] Svelte outfit canvas component
- [ ] HTMX templates (Templ)
- [ ] Tests
- [ ] CI/CD pipeline

---

## Tech Stack (Decided)

| Layer | Technology | Notes |
|-------|-----------|-------|
| Backend | Go 1.23+ | Using go-routinely patterns |
| Web | HTMX + Templ | Server-side rendering |
| Interactivity | Alpine.js | Dropdowns, modals, toasts |
| Outfit Canvas | Svelte | Single component, compiled to IIFE |
| Database | PostgreSQL 18 | With sqlc for type-safe queries |
| Migrations | goose | SQL-based migrations |
| Storage | SeaweedFS | S3-compatible object storage |
| Auth | OAuth2/OIDC | Authelia, Authentik, or similar |
| AI - Background | rembg (local) | Background removal sidecar |
| AI - Vision/Text | Ollama (local) | llava:7b + llama3.2:3b |
| Package Manager | Bun | For frontend (Svelte) build |
| Build | Nix Flakes | Reproducible builds |

---

## Mockup Pages

All mockups are in `mockups/` directory. Open with browser or local server.

| Page | Purpose | Key Interactions |
|------|---------|------------------|
| `login.html` | OAuth login | OAuth buttons, demo mode |
| `index.html` | Dashboard | Stats overview, quick actions |
| `upload.html` | File upload | Drag-drop, progress indicator |
| `item-detail.html` | Edit item | Form, log wear, delete |
| `wardrobe.html` | Item grid | Category filter, click to edit |
| `outfit-builder.html` | **Canvas** | Drag-drop, z-index, save |
| `outfits.html` | Outfit gallery | Click to edit, create new |
| `calendar.html` | Outfit planning | Day selection, weather |
| `analytics.html` | Charts & stats | Tables, chart placeholders |
| `settings.html` | Preferences | AI config, export, delete |
| `share.html` | Public share | Read-only outfit view |

**JavaScript features** (mockups/js/app.js):
- LocalStorage persistence (simulates backend)
- Full drag-and-drop canvas with z-index control
- Modal system for confirmations
- Toast notifications
- Upload with progress simulation
- Category filtering
- Outfit save/load

---

## Key Architecture Decisions

### Hybrid Frontend (HTMX + Alpine + Svelte)
- **HTMX**: 95% of UI (CRUD, navigation, forms, infinite scroll)
- **Alpine.js**: Simple client state (dropdowns, modals)
- **Svelte**: Outfit canvas only (complex drag-drop positioning)
- **Rationale**: Simpler than full SPA, smaller bundle, better SSR

### No ClamAV in MVP
- Using file validation instead (type, size, magic number)
- Can add virus scanning later if needed

### Cursor Pagination
- Not offset-based (better performance at scale)
- Pattern from go-routinely

### SSE for Real-time Updates
- AI processing status via Server-Sent Events
- Not WebSockets (simpler, HTTP-native)

---

## File Locations

```
gothreads/
├── backend/                    # Go backend (empty, needs scaffolding)
│   ├── cmd/                    # Entry points
│   ├── config/                 # Configuration
│   ├── service/                # Business logic
│   ├── store/db/sqlc/          # Database (migrations, queries)
│   ├── transport/http/         # HTTP layer (handlers, middleware, views)
│   └── static/                 # Embedded assets (js, css, icons)
├── frontend/                   # Svelte (outfit canvas only)
├── mockups/                    # HTML/CSS/JS prototypes ✓
│   ├── css/style.css           # All styles
│   ├── js/app.js               # All JavaScript
│   └── *.html                  # 11 pages
├── docs/
│   ├── ARCHITECTURE.md         # Technical decisions ✓
│   ├── SPEC.md                 # Product requirements ✓
│   ├── GETTING_STARTED.md      # Dev setup guide ✓
│   └── ai/
│       ├── IMPLEMENTATION_TASKS.md  # Full task checklist ✓
│       └── HANDOFF.md          # This document ✓
├── docker-compose.yml          # Development services ✓
├── flake.nix                   # Nix environment ✓
├── Taskfile.yml                # Task runner ✓
└── README.md                   # Project overview ✓
```

---

## Next Steps (Priority Order)

### Phase 1: Backend Foundation
1. Initialize Go module: `cd backend && go mod init`
2. Create directory structure
3. Copy patterns from go-routinely (config, pgx, middleware)
4. Create database migrations (see IMPLEMENTATION_TASKS.md)
5. Run `sqlc generate`
6. Create basic services (User, Item)

### Phase 2: HTTP Layer
1. Setup routes with go-pkgz/routegroup
2. Implement auth middleware (OIDC)
3. Create handlers for items CRUD
4. Add Templ templates (convert mockup HTML)

### Phase 3: Upload & AI
1. Implement S3 client for SeaweedFS
2. Create upload handler with SSE progress
3. Integrate rembg for background removal
4. Create AI queue worker

### Phase 4: Svelte Canvas
1. Initialize Bun project in frontend/
2. Create OutfitCanvas.svelte component
3. Build to single JS file
4. Integrate with Templ

---

## Reference Projects

- **go-routinely** (same author): Patterns for config, database, testing
  - Location: `/home/haseebmajid/Documents/go-routinely/go-routinely-main/`
  - Key files: `config/`, `store/db/`, `transport/http/middleware/`

---

## Commands

```bash
# Enter development environment
nix develop
# or: direnv allow

# Start services (PostgreSQL, SeaweedFS, mock-oauth2, rembg, ollama)
docker compose up -d

# Pull AI models (first time only, or run ollama-pull service)
docker exec -it gothreads-ollama-1 ollama pull llava:7b
docker exec -it gothreads-ollama-1 ollama pull llama3.2:3b

# Run database migrations
task db:migrate

# Generate code (sqlc, templ, mocks)
task generate

# Start development servers
task dev

# Run tests
task test

# Build for production
task build
```

---

## Open Questions (For Next Developer)

1. **Database provider**: Use existing PostgreSQL or provision new?
2. **OAuth provider**: Which OIDC provider to test with first?
3. **Domain**: What domain/subdomain for deployment?
4. **CI/CD**: GitLab CI or GitHub Actions?
5. **Secrets management**: Vault, SOPS, or simple env vars?

---

## Resumption Tips

1. **Review mockups first**: Open `mockups/index.html` in browser to understand the UX
2. **Check IMPLEMENTATION_TASKS.md**: Has detailed checklist with file paths
3. **Reference go-routinely**: Copy patterns, don't reinvent
4. **Start small**: Get one endpoint working end-to-end first
5. **Test early**: Use pgtestdb pattern for isolated test databases

---

## Contact

- Repository: https://github.com/yourusername/gothreads
- Issues: Use GitHub Issues for bugs/features
- Documentation: See `docs/` directory

---

*This document provides everything needed to resume work on go-threads. All planning is complete - implementation can begin immediately.*
