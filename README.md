# 🧵 go-threads

Privacy-first, self-hosted digital wardrobe manager. Track clothing, build outfits, and gain style insights.

## Features

- **Digital Wardrobe** - Catalog clothing with photos, metadata, pricing
- **Outfit Builder** - Visual drag-and-drop canvas with Svelte
- **Wear Tracking** - Log outfit usage, cost-per-wear analytics
- **AI Processing** - Local background removal, auto-tagging, outfit scoring
- **PWA** - Install on mobile, offline mode, push notifications
- **Sharing** - Public outfit links via short URLs

## Quick Start

**Prerequisites:** [Nix](https://nixos.org/download.html) with flakes, Docker & Docker Compose

```bash
# Clone and enter directory
git clone https://github.com/yourusername/gothreads.git
cd gothreads

# Setup environment
cp .envrc.example .envrc
direnv allow  # or: nix develop

# Initialize project
task setup

# Start development
task dev
```

Open http://localhost:8080

## Development

```bash
task dev              # Start dev servers (backend + frontend)
task test             # Run all tests
task build            # Build backend + frontend
task db:migrate       # Run database migrations
task generate         # Generate code (sqlc, templ, mocks)
```

## Stack

- **Backend**: Go, HTMX, Templ, PostgreSQL, SeaweedFS
- **Frontend**: HTMX + Alpine.js (most UI), Svelte (outfit canvas only)
- **AI**: Local sidecars (rembg for background removal)
- **Auth**: OAuth2/OIDC (Authelia, Authentik)
- **Build**: Nix Flakes, gomod2nix, Bun

## Mockups

Interactive HTML prototypes demonstrating the full UX:

```bash
cd mockups && python3 -m http.server 8000
# Visit http://localhost:8000
```

11 pages: login, dashboard, upload, wardrobe, **outfit-builder** (drag-drop canvas), outfits, calendar, analytics, settings, share

## Documentation

- [Architecture](docs/ARCHITECTURE.md) - Technical decisions
- [Specification](docs/SPEC.md) - Product requirements
- [Getting Started](docs/GETTING_STARTED.md) - Development guide
- [Implementation Tasks](docs/ai/IMPLEMENTATION_TASKS.md) - Full checklist
- [Handoff](docs/ai/HANDOFF.md) - Project context for resuming work

## License

MIT - See [LICENSE](LICENSE)

## Acknowledgments

Inspired by [Stylebook](https://www.stylebookapp.com/), [Whering](https://whering.co.uk/), [Twelve70](https://twelve70.com/). Built with patterns from [go-routinely](https://github.com/hmajid2301/go-routinely).
