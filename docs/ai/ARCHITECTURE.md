# go-threads: Implementation Planning

## Project Overview
Privacy-first, self-hosted digital wardrobe manager built with Go, HTMX, Templ, and Svelte.

## Decisions Made

### MVP Scope (Phase 1)
**Core Inventory System**
- Item CRUD (name, price, source_url, photos)
- S3 storage integration (raw photos)
- Basic wear tracking and cost-per-wear stats

**Outfit Builder (Hybrid HTMX + Svelte)**
- HTMX-powered item selection interface
- Svelte preview component for visual outfit composition
- Basic outfit saving and retrieval

**AI Screenshot Parser**
- Upload product page screenshot
- Extract: product name, price, brand, image URL
- Automatically fetch and store product image

### Post-MVP AI Features (Priority Order)
1. **Background Removal** (rembg sidecar)
2. **Auto-tagging** (color, category, season)
3. **Outfit Compatibility Scoring** (LLM-based fashion analysis)
4. **Virtual Try-On** (OOTDiffusion integration)
5. **Smart Outfit Suggestions** (weather, occasion, wear history)
6. **Style Trend Analysis** (social media scraping)

### AI Innovation Features
- **Outfit compatibility scoring**: Rate pairings based on color theory and style rules
- **Style trend analysis**: Identify trending pieces similar to user's wardrobe

### Tech Stack
- **Backend**: Go (Golang)
- **Database**: PostgreSQL with JSONB for flexible metadata
- **Storage**: S3-compatible (MinIO for local dev)
- **Frontend**: HTMX + Templ (Core) | Svelte (Outfit Canvas & Try-On Tool)
- **AI Provider Pattern**: Swappable providers via config (local-first, cloud fallback)
- **Deployment**: Docker Compose (primary) + Nix flakes (declarative option)

### User Context
- Proficient in Go, HTMX, and PostgreSQL
- Will use both Docker and Nix, but local setup is priority
- Has boilerplate from another project to copy from

## Technical Decisions

### Database & Migrations
- **Migration Tool**: pressly/goose (embedded migrations, Go-native)

### Authentication
- **Auth Model**: OAuth2/OIDC integration (Authelia, Authentik)
- Multi-user support with identity provider integration

### Object Storage
- **Storage Backend**: SeaweedFS (S3-compatible, fast, simple, active development)
- Note: MinIO no longer fully open source, avoiding vendor lock-in
- Docker deployment with SeaweedFS container

### Project Structure
- **Boilerplate Source**: `~/Documents/go-routinely/`
- **Directory Layout** (adopting clean architecture):
  ```
  gothreads/
  ├── cmd/                    # CLI utilities (scrapers, migrations)
  ├── config/                 # Environment-based configuration
  ├── service/                # Business logic layer (wardrobe, outfits, AI)
  ├── store/                  # Data access layer (sqlc generated)
  │   └── db/
  │       ├── sqlc/migrations/  # Goose migrations
  │       └── query.sql         # Source SQL for sqlc
  ├── transport/http/
  │   ├── handlers/           # HTTP handlers by feature
  │   ├── middleware/         # Auth, logging, error handling
  │   ├── views/              # Templ templates
  │   │   ├── components/
  │   │   ├── layouts/
  │   │   └── pages/
  │   └── server.go
  ├── static/                 # Embedded assets (CSS, JS, icons)
  ├── tests/e2e/              # Playwright E2E tests
  └── main.go
  ```

### API Design
- **Pattern**: HTMX-first (HTML over the wire)
- Most endpoints return HTML fragments
- Minimal JSON endpoints (primarily for Svelte components)
- Pure hypermedia-driven application

### Authentication Implementation
- **Integration**: Direct OIDC in Go (using coreos/go-oidc)
- Handle OAuth flow in application
- Support multiple providers (Authelia, Authentik, generic OIDC)

## Testing Strategy
- **Comprehensive Testing**: Unit + Integration + E2E
  - **Unit Tests**: Business logic (wear tracking, cost calculations, outfit logic)
  - **Integration Tests**: API endpoints, database interactions, full request/response
  - **E2E Tests**: Playwright/Cypress for full user workflows and HTMX interactions

## AI Provider Architecture
- **Pattern**: Strategy pattern with provider interfaces
- Define interfaces for each capability:
  - `BgRemovalProvider` (background removal)
  - `TaggingProvider` (auto-tagging)
  - `VTOProvider` (virtual try-on)
  - `CompatibilityProvider` (outfit scoring)
- Swap implementations (local sidecars vs cloud APIs)

## Database Schema - MVP
- **Initial Tables**: Items + Outfits
  - `items`: Core wardrobe inventory
  - `outfits`: Outfit compositions and metadata
- **Later Additions**: wear_history, calendar, ratings, laundry_tracking

## Architecture Patterns from go-routinely

### Core Stack Decisions
- **Web Framework**: stdlib `net/http` + `go-pkgz/routegroup`
- **Database Driver**: `pgx/v5` with connection pooling
- **SQL Tool**: `sqlc` (type-safe SQL, no ORM)
- **Templating**: `templ` (Go-native, type-safe HTML)
- **Styling**: TailwindCSS + DaisyUI
- **Auth**: `golang-jwt/jwt/v5` with OIDC (coreos/go-oidc)
- **Validation**: `go-playground/validator/v10`
- **UUID**: `gofrs/uuid/v5` (UUID v7 support)
- **Config**: `sethvargo/go-envconfig`
- **Testing**: `testify`, `mockery`, `playwright-go`
- **Dev Tools**: `air` (live reload), `go-task`

### Key Patterns to Adopt
1. **Dependency Injection**: Constructor functions with interface-based services
2. **Database Layer**: Custom retry wrapper around pgxpool for transient failures
3. **Error Handling**: Domain-specific errors with `errors.Is()` discrimination
4. **Logging**: Structured logging with `slog` (JSON output, context-aware)
5. **Auth Flow**: JWT validation with JWKS, context-based user injection
6. **Testing**: Unit (mocked interfaces) + Integration (real DB) + E2E (Playwright)
7. **Configuration**: Hierarchical env var structs with validation at load time
8. **Build**: Nix flake for reproducible dev env, embedded static files with `go:embed`

### Frontend Integration Details - Hybrid Approach

**Philosophy**: HTMX-first with minimal Svelte for complex interactions

#### Core Stack
- **HTMX**: Server-driven HTML over the wire (95% of app)
- **Alpine.js**: Lightweight reactivity for dropdowns, modals, simple UI state
- **Templ**: Type-safe server-side HTML templating
- **Svelte**: Single component for outfit canvas (complex drag-drop)
- **TailwindCSS + DaisyUI**: Styling and component library

#### Why Hybrid?
- **HTMX**: Perfect for CRUD, navigation, forms, infinite scroll
- **Alpine.js**: Great for simple client-side state (dropdowns, accordions)
- **Svelte**: Best for complex canvas interactions (drag-drop positioning)

#### File Structure
```
backend/static/
├── js/
│   ├── htmx.min.js          # 14KB - Core HTMX
│   ├── alpine.min.js        # 15KB - Alpine.js
│   ├── outfit-canvas.js     # 60KB - Compiled Svelte component
│   ├── apexcharts.min.js    # Charts for analytics
│   └── app.js               # Custom Alpine components
├── css/
│   ├── input.css            # Tailwind source
│   └── output.css           # Generated (gitignored)
├── icons/                   # PWA icons
├── manifest.json
└── sw.js                    # Service worker
```

#### Outfit Canvas Implementation

**Svelte Component** (frontend/src/OutfitCanvas.svelte):
```typescript
<script lang="ts">
  import Draggable from './Draggable.svelte';

  export let outfitId: string;
  export let items: Item[];

  let canvasItems = items.map(item => ({
    ...item,
    x: item.position_x || 100,
    y: item.position_y || 100,
    z: item.z_index || 0
  }));

  function saveOutfit() {
    htmx.ajax('POST', `/outfits/${outfitId}`, {
      values: { items: JSON.stringify(canvasItems) },
      target: '#save-message'
    });
  }
</script>

<div class="relative w-full h-96 bg-gray-50 border">
  {#each canvasItems as item (item.id)}
    <Draggable bind:x={item.x} bind:y={item.y} bind:z={item.z}>
      <img src={item.imageUrl} alt={item.name} class="max-w-24" />
    </Draggable>
  {/each}

  <button on:click={saveOutfit} class="btn btn-primary">
    Save Outfit
  </button>
</div>
```

**Integration in Templ** (transport/http/views/pages/outfit_editor.templ):
```go
templ OutfitEditor(outfit Outfit) {
  <div id="outfit-canvas"
       data-outfit-id={ outfit.ID }
       data-items={ toJSON(outfit.Items) }>
    Loading canvas...
  </div>

  <script>
    // Initialize Svelte component
    new OutfitCanvas({
      target: document.getElementById('outfit-canvas'),
      props: {
        outfitId: document.getElementById('outfit-canvas').dataset.outfitId,
        items: JSON.parse(document.getElementById('outfit-canvas').dataset.items)
      }
    });
  </script>
}
```

#### Build Process

**Vite Config** (frontend/vite.config.js):
```javascript
export default {
  build: {
    lib: {
      entry: 'src/OutfitCanvas.svelte',
      name: 'OutfitCanvas',
      fileName: 'outfit-canvas',
      formats: ['iife'] // Single file for browser
    },
    outDir: '../backend/static/js',
    rollupOptions: {
      external: ['htmx.org'], // Use global htmx
    }
  }
}
```

**Task Integration**:
```bash
# Build Svelte component
task build:frontend
# → Outputs: backend/static/js/outfit-canvas.js

# Embed in Go binary
go:embed static
var staticFiles embed.FS
```

#### Everything Else: HTMX + Alpine

**Item List** (HTMX infinite scroll):
```html
<div hx-get="/items?cursor=abc123"
     hx-trigger="intersect"
     hx-swap="afterend">
  <!-- Items load here -->
</div>
```

**Dropdown Menu** (Alpine.js):
```html
<div x-data="{ open: false }">
  <button @click="open = !open">Menu</button>
  <div x-show="open" x-cloak>
    <!-- Dropdown content -->
  </div>
</div>
```

**Item Upload** (HTMX + Alpine):
```html
<form hx-post="/items/upload"
      hx-encoding="multipart/form-data"
      x-data="{ uploading: false }"
      @htmx:before-request="uploading = true">

  <input type="file" name="image">
  <button :disabled="uploading">
    <span x-show="!uploading">Upload</span>
    <span x-show="uploading">Uploading...</span>
  </button>
</form>
```

#### Bundle Size Comparison

**Our Hybrid Approach**:
- HTMX: 14KB
- Alpine.js: 15KB
- Svelte component: ~60KB (only loads on outfit editor page)
- ApexCharts: ~120KB (only on analytics page)
- **Total baseline**: 29KB (HTMX + Alpine)
- **Total with canvas**: 89KB

**Full Svelte Alternative**:
- Svelte runtime: ~20KB
- Router: ~10KB
- All components: ~150KB+
- **Total**: 180KB+

#### When to Use What

| Feature | Technology | Why |
|---------|-----------|-----|
| Item CRUD | HTMX | Server-rendered, simple |
| Navigation | HTMX | Hypermedia-driven |
| Forms | HTMX | Progressive enhancement |
| Infinite scroll | HTMX | Built-in support |
| Modals/dropdowns | Alpine.js | Lightweight reactivity |
| Outfit canvas | Svelte | Complex drag-drop state |
| Charts | ApexCharts + Alpine | Rich visualizations |
| Offline sync | Service Worker | PWA standard |

#### Asset Serving
- **Development**: CDN for HTMX/Alpine, local Vite dev server for Svelte
- **Production**: Embed all in Go binary with `go:embed`

### Error Handling Implementation
- **HTMX Error Responses**: HX-Trigger headers for toast notifications
- **HTTP Handler Wrapper**: Catch errors, return appropriate status codes
- **Inline Errors**: Return partial HTML for form validation errors
- **Recovery Middleware**: Panic recovery with structured logging

## Post-MVP Feature Roadmap
1. **Background Removal AI Pipeline** (rembg sidecar)
2. **Wear Tracking and Calendar Integration** (wear_history table)
3. **Multi-Rating System** (personal/spouse/social ratings)
4. **Screenshot Product Parser** (deferred after web scraper)

## Implementation Decisions - Round 2

### AI Provider Configuration
- **Storage**: Database-driven configuration
- Admin UI to manage provider settings (enable/disable, API keys, endpoints)
- Runtime switching between local sidecars and cloud APIs
- Schema: `ai_providers` table with (name, type, enabled, config_json, priority)

### Observability Stack
- **Comprehensive Observability**: Logging + Tracing + Metrics
  - **Logging**: Structured logs with `slog` (JSON output)
  - **Tracing**: OpenTelemetry for request flows and AI pipeline debugging
  - **Metrics**: Prometheus for performance tracking (API latency, DB queries, AI duration)

### Image Processing Pipeline
- **Flow**: Synchronous processing (block until complete)
  1. User uploads image (via form or screenshot)
  2. Background removal processing (rembg or configured provider)
  3. Store processed image in SeaweedFS (S3)
  4. Create database record with S3 keys
  5. Return success response with item details
- Simple, immediate feedback, no queue complexity in MVP

### NixOS Module Design
- **Integrated Stack**: Full deployment bundle
  - gothreads service (systemd)
  - PostgreSQL with automatic migrations
  - SeaweedFS with S3 configuration
  - Optional: AI sidecars (rembg, OOTDiffusion)
- Declarative configuration via NixOS options
- Example:
  ```nix
  services.gothreads = {
    enable = true;
    database.enable = true;  # or use external
    storage.enable = true;   # SeaweedFS
    ai.localSidecars = [ "rembg" ];
  };
  ```

## Implementation Decisions - Round 3

### CI/CD & Build
- **Platform**: GitLab CI
- Self-hostable, built-in container registry
- Nix-powered reproducible builds in pipeline
- Automated testing (unit, integration, E2E)

### AI Sidecar Images
- **rembg**: Use public image (danielgatis/rembg or similar)
- **OOTDiffusion**: Custom Dockerfile in repo (`docker/ootdiffusion/`)
- Registry: GitLab Container Registry for custom images
- Tag strategy: `:latest`, `:v1.0.0`, `:commit-sha`

### GPU Support Strategy
- **Optional GPU Acceleration**: Detect and use if available
- Docker GPU runtime detection (`nvidia-docker`, `--gpus all`)
- CPU fallback for all operations (slower but works everywhere)
- VTO performance warning if GPU unavailable
- Config flag: `ENABLE_GPU=true` (default: auto-detect)

### Development S3 Setup
- **SeaweedFS in Docker Compose**: Production parity
- Volume mount for persistence: `./data/seaweedfs`
- S3 endpoint: `http://localhost:8333`
- Bucket auto-creation on startup
- Same S3 client code in dev and production

## Implementation Decisions - Round 4

### Backup Strategy
- **Deferred**: Not included in MVP
- Document recommended approaches in README
- Users can implement pg_dump + S3 sync as needed

### Development Environment Bootstrap
- **Single Command**: `task dev`
  - Starts docker-compose (PostgreSQL, SeaweedFS, AI sidecars)
  - Runs database migrations (goose)
  - Starts parallel live reload (air + templ + tailwind)
  - Sets up test database and fixtures
- Nix flake provides all tools, task orchestrates workflow

### Health Checks
- **Basic HTTP Endpoint**: `GET /health`
  - Returns 200 OK with minimal JSON: `{"status": "ok"}`
  - Used by Docker Compose and container orchestration
  - Future: Add detailed checks when needed

### Rate Limiting
- **Not in MVP**: Skip rate limiting initially
- Self-hosted environment, trusted users
- Can add later if abuse becomes concern

## Implementation Decisions - Round 5

### Environment Variable Management
- **direnv with .envrc**: Automatic env loading when entering directory
- Committed `.envrc.example`, local `.envrc` in .gitignore
- Nix-friendly, integrates with flake.nix
- Works for both shell and task commands

### Image Upload Constraints
- **Max Size**: 10MB per upload
- **Allowed Formats**: JPEG, PNG, WebP only
- **Validation**: Server-side MIME type checking + magic number validation
- **Rejection**: Return 400 Bad Request with clear error message

### S3 Bucket Organization
- **Structure**: `users/<user_id>/items/<type>/<uuid>.<ext>`
  - Example: `users/abc-123/items/raw/def-456.jpg`
  - Example: `users/abc-123/items/processed/def-456.png`
- Per-user isolation for access control
- Separate raw and processed images
- Easy cleanup when user deletes account

### AI Processing Timeouts
- **Background Removal**: 30 seconds
- **Virtual Try-On**: 2 minutes
- **Auto-Tagging**: 30 seconds
- **Outfit Scoring**: 30 seconds
- Return timeout error with fallback message to user

## Implementation Decisions - Round 6

### Database Schema Design

#### `items` Table
```sql
CREATE TABLE items (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    price NUMERIC(10, 2),
    source_url TEXT,
    brand TEXT,
    category TEXT NOT NULL,  -- e.g., 'tops', 'bottoms', 'shoes', 'accessories'
    color TEXT,
    season TEXT,             -- e.g., 'spring', 'summer', 'fall', 'winter', 'all-season'
    size TEXT,
    material TEXT,
    care_instructions TEXT,
    s3_key_raw TEXT NOT NULL,
    s3_key_processed TEXT,   -- After background removal
    wear_count INTEGER DEFAULT 0,
    max_wears INTEGER DEFAULT 3,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
```

#### `outfits` Table
```sql
CREATE TABLE outfits (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name TEXT,
    notes TEXT,
    vto_preview_key TEXT,    -- S3 key for virtual try-on preview
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
```

#### `outfit_items` Join Table
```sql
CREATE TABLE outfit_items (
    outfit_id UUID NOT NULL REFERENCES outfits(id) ON DELETE CASCADE,
    item_id UUID NOT NULL REFERENCES items(id) ON DELETE CASCADE,
    position_x INTEGER,       -- For visual canvas positioning
    position_y INTEGER,
    z_index INTEGER,          -- Layering order
    PRIMARY KEY (outfit_id, item_id)
);
```

#### `wear_history` Table
```sql
CREATE TABLE wear_history (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    outfit_id UUID NOT NULL REFERENCES outfits(id) ON DELETE CASCADE,
    worn_date DATE NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(user_id, outfit_id, worn_date)
);
```

#### `outfit_ratings` Table
```sql
CREATE TABLE outfit_ratings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    outfit_id UUID NOT NULL REFERENCES outfits(id) ON DELETE CASCADE,
    rating_type TEXT NOT NULL,  -- 'personal', 'spouse', 'social'
    rating_value INTEGER NOT NULL CHECK (rating_value BETWEEN 1 AND 5),
    rated_at TIMESTAMPTZ DEFAULT NOW(),
    notes TEXT,
    UNIQUE(outfit_id, rating_type)
);
```

#### `ai_providers` Table (Database-driven AI config)
```sql
CREATE TABLE ai_providers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    name TEXT NOT NULL UNIQUE,           -- 'rembg-local', 'openai-vision', etc.
    provider_type TEXT NOT NULL,         -- 'background_removal', 'tagging', 'vto', 'scoring'
    enabled BOOLEAN DEFAULT true,
    priority INTEGER DEFAULT 0,          -- Higher priority = preferred provider
    config JSONB,                        -- Provider-specific config (endpoints, API keys, etc.)
    timeout_seconds INTEGER DEFAULT 30,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
```

## Implementation Decisions - Round 7

### OAuth Provider in Development
- **mock-oauth2-server Container**: Same pattern as go-routinely
- Docker Compose service on port 8090
- Configurable claims for test users
- E2E tests use mock OAuth flow

### AI Provider Fallback Strategy
- **Queue for Retry**: Don't fail immediately
- Save failed processing job to `ai_processing_queue` table
- Background worker retries failed jobs
- User sees "Processing..." state, updates when complete
- Notification when processing completes or fails permanently

### AI Provider Config Caching
- **In-Memory Cache**: Load on startup, refresh every 5 minutes
- Background goroutine reloads from database
- Manual reload via admin API: `POST /admin/reload-providers`
- Cache invalidation on provider config changes

### Test Fixtures and Seeding
- **Hybrid Approach**: SQL seed files + Go helpers
- `store/db/sqlc/seeds/`: SQL files for base data
- `gothreadstest/`: Go test helpers for dynamic fixtures
- Factory pattern for complex objects (outfits with items)
- Seed data: AI providers, categories, test users

### AI Processing Queue (New Table)
```sql
CREATE TABLE ai_processing_queue (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    item_id UUID REFERENCES items(id) ON DELETE CASCADE,
    provider_type TEXT NOT NULL,  -- 'background_removal', 'tagging', etc.
    provider_id UUID REFERENCES ai_providers(id),
    status TEXT NOT NULL,         -- 'pending', 'processing', 'completed', 'failed'
    retry_count INTEGER DEFAULT 0,
    max_retries INTEGER DEFAULT 3,
    input_data JSONB,             -- Provider-specific input
    output_data JSONB,            -- Result data
    error_message TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    completed_at TIMESTAMPTZ
);
```

## Implementation Decisions - Round 8

### AI Processing Workers
- **Concurrency**: 2-3 workers (configurable via `AI_WORKER_CONCURRENCY`)
- Conservative resource usage for self-hosted environments
- Workers poll `ai_processing_queue` for pending jobs
- Exponential backoff retry: 1min, 2min, 4min, 8min (max 4 retries)

### Schema Strategy
- **MVP Schema Only**: Ship with core tables
- Future migrations for:
  - `calendar_events` (when implementing outfit planning)
  - `laundry_tracking` (when adding wash cycle features)
  - `screenshot_uploads` (when adding screenshot parser)
- Keep migrations incremental and focused

### AI Job Retry Strategy
- **Exponential Backoff**: 1min → 2min → 4min → 8min
- Max 4 retries before marking as permanently failed
- Notify user on permanent failure
- Exponential: `delay = base_delay * 2^(attempt-1)` with jitter

## Implementation Decisions - Round 9

### Graceful Shutdown
- **Timeout**: 30 seconds for in-flight AI jobs
- Context cancellation signal sent immediately
- Workers save incomplete jobs back to queue
- Fast deployments, jobs resume after restart

### Item Creation Flow
- **Upload Image First**: Immediate AI processing
  1. User uploads image (drag-drop or file picker)
  2. Background removal starts (SSE updates progress)
  3. AI tagging pre-fills form (if enabled)
  4. User reviews/edits fields
  5. Submit saves item to database

### Outfit Builder Integration
- **Hybrid HTMX + Svelte**:
  - HTMX loads item list as HTML fragments
  - Svelte provides drag-drop canvas overlay
  - Items fetched via HTMX: `GET /wardrobe/items` returns HTML
  - Canvas state saved via HTMX: `POST /outfits` with form data
  - Svelte only for visual positioning, no API calls

### Real-time AI Processing Updates
- **Server-Sent Events (SSE)**: One-way server→client
- Endpoint: `GET /ai/status/:job_id`
- Events: `processing`, `progress`, `completed`, `failed`
- HTMX triggers SSE connection after upload
- Close connection on completion or error

### PWA Support (NEW REQUIREMENT)
- **Progressive Web App**: Installable, offline-capable
- Features from go-routinely reference:
  - `manifest.json` with app metadata
  - Service worker (`sw.js`) for offline caching
  - Cache static assets (CSS, JS, icons)
  - Cache API responses (items, outfits) with stale-while-revalidate
  - Installable on mobile devices
  - App-like experience

## Implementation Decisions - Round 10

### PWA Capabilities
- **Full Offline Mode with Background Sync**:
  - Cache static assets (CSS, JS, icons, fonts)
  - Cache API responses (items, outfits) with IndexedDB
  - Offline uploads queue in IndexedDB
  - Background Sync API syncs queued uploads when reconnected
  - Service worker handles offline/online transitions
  - Cache strategy: Network-first for API, cache-first for static assets

### Push Notifications
- **For Long-Running Jobs**: VTO and background removal
- Web Push API with VAPID keys
- Notification permissions requested after first successful upload
- Subscription stored in database: `push_subscriptions` table
- Worker triggers push on AI job completion
- Deep link to processed item/outfit

### Pagination Strategy
- **Infinite Scroll with Cursor**: Modern, efficient
- Cursor-based pagination using UUID + created_at
- Load 24 items per page
- HTMX `hx-trigger="intersect"` for scroll loading
- Backend: `WHERE (created_at, id) < ($1, $2) ORDER BY created_at DESC, id DESC LIMIT 24`

### Additional Feature Areas (Post-MVP)

#### Sharing & Social Features
- Export outfits as images (PNG with branded watermark)
- Share outfit links (public read-only view)
- Outfit inspiration feed (community-shared looks)
- Friend connections and outfit recommendations
- Privacy controls (private, friends-only, public)

#### Analytics & Insights
- Wardrobe statistics dashboard
- Spending analysis (total invested, cost-per-wear trends)
- Wear frequency heatmaps
- Category balance (too many shoes? not enough tops?)
- Seasonal coverage analysis
- ROI calculator (cost per wear vs price)

#### Mobile-Specific Features
- Native camera integration (capture photos in-app)
- Barcode/QR scanning for product lookup
- AR virtual try-on (device camera + overlay)
- Voice notes for outfit descriptions
- Location-based outfit suggestions (weather, occasion)
- Apple Watch / Wear OS outfit-of-the-day widget

## Implementation Decisions - Round 11

### Social Sharing (MVP Scope)
- **Public Outfit Links**: Share via unique URL
- New table: `outfit_shares` (id, outfit_id, slug, public, created_at)
- Generate short slug (e.g., `/s/abc123`)
- Public read-only view of outfit with items
- Privacy toggle: private (default) or public
- No user profiles or gallery in MVP
- No friend connections in MVP

### Analytics (MVP Scope)
- **Extended Stats + Charts**: Rich insights from day one
- Dashboard includes:
  - Total items, total spent, average cost-per-wear
  - Wear frequency chart (last 30/90/365 days)
  - Category breakdown (pie chart)
  - Top 5 most/least worn items
  - Spending over time (line chart)
  - Cost-per-wear distribution
- Charts: Use Chart.js or similar (lightweight, Templ-compatible)
- Real-time calculation (no batch processing for MVP)

### Mobile Features (All Four - Phased Implementation)
1. **Phase 1 (MVP)**:
   - Camera integration (capture photos in-app)
   - Screenshot-based product parser (already planned)
2. **Phase 2 (Post-MVP)**:
   - Weather-based suggestions (location + weather API)
3. **Phase 3 (Post-MVP)**:
   - AR virtual try-on (complex, requires OOTDiffusion refinement)

### Upload Concurrency
- **Limit**: 10 uploads per user concurrently
- Tracked in-memory (user_id → active upload count)
- Return 429 Too Many Requests if limit exceeded
- Global limit: None (trust multi-user won't overwhelm)

### New Tables for Sharing & Mobile

#### `outfit_shares` Table
```sql
CREATE TABLE outfit_shares (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    outfit_id UUID NOT NULL REFERENCES outfits(id) ON DELETE CASCADE,
    slug TEXT NOT NULL UNIQUE,
    public BOOLEAN DEFAULT false,
    view_count INTEGER DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
```

#### `weather_suggestions` Table (Phase 2)
```sql
CREATE TABLE weather_suggestions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    location TEXT NOT NULL,
    weather_condition TEXT,  -- 'sunny', 'rainy', 'cold', etc.
    temperature_range TEXT,  -- '10-15°C'
    suggested_outfit_ids UUID[],
    created_at TIMESTAMPTZ DEFAULT NOW()
);
```

## Implementation Decisions - Round 12

### Outfit Share Slugs
- **nanoid**: Short, URL-safe, collision-resistant
- Length: 10 characters (e.g., `V1StGXR8_Z`)
- Library: `github.com/matoous/go-nanoid`
- Example URL: `https://gothreads.app/s/V1StGXR8_Z`

### Analytics Charts
- **ApexCharts**: Modern, interactive, responsive
- Lightweight enough for PWA
- Beautiful defaults, good mobile UX
- Integrate via CDN or bundled with Vite (for Svelte components)

### EXIF Data Handling
- **Strip All EXIF**: Privacy-first approach
- Remove GPS, camera info, timestamps before S3 upload
- Library: `github.com/dsoprea/go-exif` or `github.com/rwcarlsen/goexif`
- Process during image upload, before background removal

### Upload Concurrency Tracking
- **In-Memory Map**: `sync.Map` of user_id → active upload count
- Increment on upload start, decrement on completion/error
- Thread-safe for concurrent requests
- Reset on server restart (acceptable for MVP)

## Implementation Decisions - Round 13

### Advanced AI Features (All Post-MVP, Phased)
1. **Duplicate Detection** (Phase 2):
   - Perceptual hashing (pHash) for image similarity
   - Warn user when uploading similar item
   - Merge or keep separate option

2. **Style Profile Quiz** (Phase 3):
   - Onboarding quiz: lifestyle, climate, aesthetic preferences
   - Store profile: `user_style_profiles` table
   - AI uses profile for outfit suggestions

3. **AI Outfit Generation** (Phase 3):
   - Constraints: occasion, weather, color palette, mood
   - LLM-based generation using wardrobe items
   - Multiple suggestions, user picks favorite

4. **Capsule Wardrobe Builder** (Phase 4):
   - AI analyzes wardrobe, identifies gaps
   - Suggests minimal essential items
   - Build capsule subsets (work, casual, travel)

### Data Portability
- **Export to Multiple Formats**:
  - JSON (complete wardrobe backup)
  - CSV (items list, spreadsheet-compatible)
  - PDF (visual wardrobe report with images)
- Export button on settings page
- Async generation for large wardrobes
- Download via S3 pre-signed URL

### Third-Party Integrations
- **None in MVP**: Self-contained application
- Future considerations:
  - Product APIs (price updates, availability)
  - Social media import (Instagram, Pinterest)
  - Webhooks for automation (IFTTT, Zapier)

### Performance Optimizations (All Implemented)
1. **Image CDN**: imgproxy or CloudFlare Images
   - On-the-fly resizing (thumbnails, previews)
   - Format conversion (WebP for modern browsers)
   - Caching and compression

2. **S3 Pre-Signed URLs**: Direct uploads
   - Client requests upload URL from backend
   - Client uploads directly to S3
   - Backend validates and processes post-upload
   - Reduces server bandwidth

3. **Database Optimization**:
   - Composite indexes on hot queries
   - Materialized views for analytics (refresh hourly)
   - Connection pooling tuning (50 max connections)
   - Query plan analysis for slow queries

4. **Frontend Optimization**:
   - Code splitting (Svelte components lazy loaded)
   - Image lazy loading (intersection observer)
   - Service worker caching strategy
   - Vite build optimization (tree shaking, minification)

## Implementation Decisions - Round 14

### Security & Privacy (Comprehensive)
1. **Content Security Policy Headers**:
   - Strict CSP to prevent XSS
   - Allow self-hosted assets only
   - Whitelist CDN for ApexCharts

2. **Virus Scanning for Uploads**:
   - ClamAV container in docker-compose
   - Scan images before S3 upload
   - Reject infected files, notify user

3. **GDPR Compliance**:
   - Data export (JSON, CSV, PDF)
   - Account deletion with cascade
   - Privacy policy page (Templ template)
   - Consent tracking for push notifications

4. **API Rate Limiting**:
   - Global: 1000 requests/hour per user
   - Upload: 10 concurrent uploads per user
   - AI processing: queue-based natural limiting
   - Middleware: `github.com/ulule/limiter`

### Scaling Strategy
- **Single Instance**: Optimized for personal/family self-hosting
- Stateless design (easy to scale horizontally if needed)
- No multi-tenancy complexity
- Keep architecture simple and maintainable

### MVP Scope Validation
- **Scope Approved**: Full feature set as planned
- Core: Inventory, outfits, wear tracking
- AI: Background removal, tagging, VTO (post-MVP), outfit scoring
- Analytics: Extended stats + charts
- Social: Public outfit links
- Mobile: Camera integration, screenshot parser
- PWA: Full offline mode, push notifications

## Implementation Decisions - Round 15

### Git Workflow
- **Trunk-Based Development**:
  - Main branch always deployable
  - Short-lived feature branches (<2 days)
  - Merge via pull requests with CI checks
  - Fast iteration, continuous integration

### Commit Conventions
- **Conventional Commits**: Strict format
  - `feat: Add outfit sharing`
  - `fix: Resolve background removal timeout`
  - `docs: Update API documentation`
  - `chore: Update dependencies`
- Enables automatic changelog generation
- Commitlint pre-commit hook

### Documentation Strategy
- **Comprehensive Documentation**:
  1. **README.md**: Quickstart, features, screenshots
  2. **Docs Site**: VitePress or Docusaurus
     - User Guide (setup, usage, features)
     - Developer Guide (architecture, contributing, testing)
     - Deployment Guide (Docker, Nix, production)
  3. **API Documentation**: OpenAPI/Swagger
     - Generate from code annotations
     - Interactive API reference
  4. **Inline Code Comments**: For complex logic

## Deep Dive: Testing Strategy

### Test Coverage Goals
- **Overall**: 80% code coverage minimum
- **Service Layer**: 90%+ (business logic critical)
- **Handlers**: 70%+ (integration tests cover most)
- **Database**: 100% of queries tested

### Mocking Patterns (from go-routinely)
- **Mockery-Generated Mocks**: Interface-based mocking
- **Test Doubles**: For external services (S3, AI providers)
- **Fixture Factories**: Builder pattern for test data

### E2E Test Scenarios
1. **User Onboarding**:
   - OAuth login flow
   - Style profile quiz
   - First item upload

2. **Core Workflows**:
   - Add item (upload → BG removal → form → save)
   - Create outfit (select items → arrange canvas → save)
   - Log wear history
   - View analytics dashboard

3. **AI Features**:
   - Background removal with progress
   - Auto-tagging validation
   - Outfit scoring

4. **PWA Features**:
   - Offline mode (cache, view items)
   - Background sync (upload while offline)
   - Push notifications (VTO completion)

5. **Sharing**:
   - Generate public link
   - View shared outfit (anonymous)
   - Track view count

### Test Organization
```
tests/
├── unit/              # Unit tests (co-located with source)
├── integration/       # Integration tests (handlers, DB)
└── e2e/              # Playwright E2E tests
    ├── auth_test.go
    ├── items_test.go
    ├── outfits_test.go
    ├── analytics_test.go
    ├── ai_test.go
    └── pwa_test.go
```

## Deep Dive: CI/CD Pipeline

### GitLab CI Stages
```yaml
stages:
  - validate
  - test
  - build
  - deploy

# Stage 1: Validate
lint:
  stage: validate
  script: golangci-lint run

format:
  stage: validate
  script: gofmt -d .

commitlint:
  stage: validate
  script: commitlint --from HEAD~1

# Stage 2: Test
unit-tests:
  stage: test
  script: gotestsum --format testname -- -race -coverprofile=coverage.out ./...

integration-tests:
  stage: test
  services: [postgres, seaweedfs, mock-oauth2]
  script: task integration:tests

e2e-tests:
  stage: test
  services: [postgres, seaweedfs, rembg]
  script: task tests:e2e
  artifacts:
    paths: [tests/videos/, tests/traces/]

# Stage 3: Build
build-binary:
  stage: build
  script: nix build
  artifacts:
    paths: [result/]

build-container:
  stage: build
  script: nix build .#container && docker load < result

# Stage 4: Deploy (manual trigger)
deploy-staging:
  stage: deploy
  when: manual
  script: ./scripts/deploy-staging.sh
```

### Deployment Automation
- **Staging**: Auto-deploy on main branch (after tests pass)
- **Production**: Manual approval required
- **Rollback**: GitLab environment history with one-click rollback

## Deep Dive: Error Handling Patterns

### Error Types
```go
// Domain errors
var (
    ErrItemNotFound      = errors.New("item not found")
    ErrInvalidCategory   = errors.New("invalid category")
    ErrOutfitNotFound    = errors.New("outfit not found")
    ErrUnauthorized      = errors.New("unauthorized")
    ErrInvalidImageFormat = errors.New("invalid image format")
    ErrImageTooLarge     = errors.New("image exceeds size limit")
    ErrAIProviderUnavailable = errors.New("AI provider unavailable")
    ErrAIProcessingTimeout   = errors.New("AI processing timeout")
)
```

### Error Wrapping
```go
// Service layer
if err := s.store.CreateItem(ctx, item); err != nil {
    return nil, fmt.Errorf("create item: %w", err)
}

// Handler discriminates
if errors.Is(err, service.ErrItemNotFound) {
    http.Error(w, "Item not found", http.StatusNotFound)
    return
}
```

### User-Facing Error Messages
```go
// Error response format
type ErrorResponse struct {
    Error   string `json:"error"`
    Message string `json:"message"`  // User-friendly message
    Code    string `json:"code"`     // Machine-readable code
}

// Example
{
  "error": "AI_PROCESSING_TIMEOUT",
  "message": "Background removal took too long. Your item has been saved and will be processed later.",
  "code": "AI_TIMEOUT"
}
```

### Logging Conventions
```go
// Structured logging with context
logger.ErrorContext(ctx, "failed to process image",
    slog.String("user_id", userID.String()),
    slog.String("item_id", itemID.String()),
    slog.String("provider", providerName),
    slog.Any("error", err),
)

// Log levels
// - DEBUG: Detailed flow, variable values
// - INFO: Significant events (user login, item created)
// - WARN: Recoverable errors (AI timeout, retry)
// - ERROR: Unexpected errors requiring investigation
```

## Ready to Start Building

All major decisions have been documented. Next steps:
1. Scaffold project structure from go-routinely
2. Generate initial database migrations
3. Implement core services and handlers
4. Build frontend templates and components
5. Integrate AI providers
6. Write tests
7. Setup CI/CD
8. Deploy!
