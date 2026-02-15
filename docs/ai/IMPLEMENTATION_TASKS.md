# go-threads Implementation Tasks

Combined checklist and structured tasks for both human developers and AI agents.

## Quick Reference

```bash
# Development
task dev              # Start all dev servers
task test             # Run all tests
task build            # Build backend + frontend
task db:migrate       # Run database migrations
task generate         # Generate code (sqlc, templ, mocks)
```

---

## Phase 0: Setup ✓

- [x] Create project structure (backend/, frontend/, docs/, mockups/)
- [x] Setup Nix flake with gomod2nix and Bun
- [x] Create docker-compose.yml (PostgreSQL, SeaweedFS, mock-oauth2, rembg)
- [x] Create Taskfile.yml
- [x] Write documentation (ARCHITECTURE.md, SPEC.md, GETTING_STARTED.md)
- [x] Create mockups (5 pages: index, upload, wardrobe, item-detail, outfit-builder)
- [ ] Initialize Git repository
- [ ] Create .gitlab-ci.yml

---

## Phase 1: Backend Foundation

### Project Structure
```bash
# Commands to run
cd backend
go mod init github.com/yourusername/gothreads
mkdir -p {cmd,config,service,store/db/sqlc/{migrations,seeds},transport/http/{handlers,middleware,views/{components,layouts,pages}},static/{css,js,icons},gothreadstest}
```

- [ ] Initialize Go module in `backend/`
- [ ] Create directory structure
- [ ] Copy .air.toml from go-routinely
- [ ] Create backend/main.go

### Configuration
**Files**: `backend/config/config.go`

- [ ] Create config package with sethvargo/go-envconfig
- [ ] Define Config struct with nested sections:
  - Database (host, port, user, password, dbname, ssl_mode)
  - Server (host, port, environment)
  - OAuth (jwks_url, client_id, client_secret, redirect_url, skip_auth)
  - S3 (endpoint, bucket, access_key, secret_key)
  - AI (worker_concurrency, enable_gpu)
- [ ] Add validation functions
- [ ] Add log level helpers

### Database Setup
**Files**: `backend/sqlc.yaml`, `backend/store/db/sqlc/migrations/*.sql`

- [ ] Setup sqlc.yaml configuration
- [ ] Create migration 00001: users table
  ```sql
  CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
  CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    oauth_provider TEXT NOT NULL,
    oauth_sub TEXT NOT NULL,
    email TEXT,
    name TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(oauth_provider, oauth_sub)
  );
  ```
- [ ] Create migration 00002: items table (see ARCHITECTURE.md for schema)
- [ ] Create migration 00003: outfits table
- [ ] Create migration 00004: outfit_items join table
- [ ] Create migration 00005: wear_history table
- [ ] Create migration 00006: outfit_ratings table
- [ ] Create migration 00007: ai_providers table
- [ ] Create migration 00008: ai_processing_queue table
- [ ] Create migration 00009: outfit_shares table
- [ ] Create migration 00010: push_subscriptions table
- [ ] Write initial SQL queries in `query.sql`:
  - GetUserByOAuthSub
  - CreateUser
  - ListItems (with cursor pagination)
  - CreateItem, GetItem, UpdateItem, DeleteItem
  - CreateOutfit, GetOutfit, ListOutfits
- [ ] Run `sqlc generate`
- [ ] Create `store/db/pgx.go` with connection pooling and retry logic

### Seed Data
**Files**: `backend/store/db/sqlc/seeds/*.sql`

- [ ] Create seed: AI providers (rembg-local, openai-vision)
- [ ] Create seed: Item categories (tops, bottoms, shoes, outerwear, accessories)
- [ ] Create seed: Test users for development

---

## Phase 2: Core Services

### User Service
**Files**: `backend/service/user.go`, `backend/service/user_test.go`

- [ ] Implement CreateUser(ctx, oauthProvider, oauthSub, email, name)
- [ ] Implement GetUserByOAuthSub(ctx, provider, sub)
- [ ] Implement GetUserByID(ctx, userID)
- [ ] Implement DeleteUser(ctx, userID) - GDPR compliance
- [ ] Write unit tests (90%+ coverage target)

### Item Service
**Files**: `backend/service/item.go`, `backend/service/item_test.go`

- [ ] Implement CreateItem(ctx, userID, item)
- [ ] Implement GetItem(ctx, userID, itemID)
- [ ] Implement ListItems(ctx, userID, cursor, limit) - cursor pagination
- [ ] Implement UpdateItem(ctx, userID, itemID, updates)
- [ ] Implement DeleteItem(ctx, userID, itemID)
- [ ] Implement IncrementWearCount(ctx, itemID)
- [ ] Write unit tests

### Outfit Service
**Files**: `backend/service/outfit.go`, `backend/service/outfit_test.go`

- [ ] Implement CreateOutfit(ctx, userID, name, notes)
- [ ] Implement GetOutfit(ctx, userID, outfitID)
- [ ] Implement ListOutfits(ctx, userID, cursor, limit)
- [ ] Implement UpdateOutfit(ctx, userID, outfitID, updates)
- [ ] Implement DeleteOutfit(ctx, userID, outfitID)
- [ ] Implement AddItemToOutfit(ctx, outfitID, itemID, position)
- [ ] Implement RemoveItemFromOutfit(ctx, outfitID, itemID)
- [ ] Write unit tests

### Analytics Service
**Files**: `backend/service/analytics.go`, `backend/service/analytics_test.go`

- [ ] Implement GetDashboardStats(ctx, userID) - total items, outfits, avg cost-per-wear
- [ ] Implement GetCostPerWear(ctx, userID) - per item
- [ ] Implement GetWearFrequency(ctx, userID, days) - 30/90/365 day charts
- [ ] Implement GetCategoryBreakdown(ctx, userID) - pie chart data
- [ ] Implement GetTopItems(ctx, userID, limit, sort) - most/least worn
- [ ] Write unit tests

---

## Phase 3: AI Integration

### Provider Interfaces
**Files**: `backend/service/ai/provider.go`

- [ ] Define BgRemovalProvider interface (RemoveBackground(ctx, imageData) (processedData, error))
- [ ] Define TaggingProvider interface (ExtractTags(ctx, imageData) (tags, error))
- [ ] Define VTOProvider interface (GenerateTryOn(ctx, userPhoto, items) (image, error)) - Post-MVP
- [ ] Define CompatibilityProvider interface (ScoreOutfit(ctx, items) (score, error)) - Post-MVP

### Background Removal (rembg)
**Files**: `backend/service/ai/rembg.go`, `backend/service/ai/rembg_test.go`

- [ ] Implement HTTP client for rembg API (http://rembg:5000/api/remove)
- [ ] Implement RemoveBackground method with multipart upload
- [ ] Add 30 second timeout
- [ ] Add error handling and retry logic
- [ ] Write unit tests with mock HTTP server

### AI Queue Worker
**Files**: `backend/service/ai_queue.go`, `backend/service/ai_queue_test.go`

- [ ] Implement EnqueueJob(ctx, userID, itemID, providerType, inputData)
- [ ] Implement ProcessJob(ctx, jobID) - runs AI provider
- [ ] Create worker pool (3 workers with configurable concurrency)
- [ ] Implement exponential backoff retry (1min, 2min, 4min, 8min)
- [ ] Add SSE status updates (/upload/:job_id/stream)
- [ ] Write unit tests

---

## Phase 4: Storage & Upload

### S3 Client (SeaweedFS)
**Files**: `backend/service/storage/s3.go`, `backend/service/storage/s3_test.go`

- [ ] Implement S3 client using AWS SDK
- [ ] Implement UploadFile(ctx, key, data, contentType)
- [ ] Implement DeleteFile(ctx, key)
- [ ] Implement GeneratePresignedURL(ctx, key, expiry)
- [ ] Write unit tests with mock S3

### Image Processing
**Files**: `backend/service/image.go`, `backend/service/image_test.go`

- [ ] Implement EXIF stripping (github.com/dsoprea/go-exif)
- [ ] Implement format validation (JPEG, PNG, WebP)
- [ ] Implement size validation (10MB max)
- [ ] Implement magic number checking
- [ ] Write unit tests

### Upload Service
**Files**: `backend/service/upload.go`, `backend/service/upload_test.go`

- [ ] Implement upload flow:
  1. Validate image (format, size, magic number)
  2. Strip EXIF data
  3. Upload raw image to S3 (users/{user_id}/items/raw/{uuid}.ext)
  4. Enqueue background removal job
  5. Return job ID for SSE tracking
- [ ] Implement concurrency tracking (10 uploads per user max, sync.Map)
- [ ] Return 429 Too Many Requests if limit exceeded
- [ ] Write unit tests

---

## Phase 5: HTTP Layer

### Middleware
**Files**: `backend/transport/http/middleware/*.go`

- [ ] Create auth.go - OAuth/OIDC middleware
  - JWT validation with JWKS
  - Context-based user injection
  - Skip auth mode for development (SKIP_AUTH=true)
- [ ] Create logging.go - Structured logging with slog
- [ ] Create recovery.go - Panic recovery
- [ ] Create ratelimit.go - github.com/ulule/limiter (1000 req/hour per user)
- [ ] Create cors.go - CORS headers

### Handlers - Items
**Files**: `backend/transport/http/handlers/item.go`, `*_integration_test.go`

- [ ] POST /items - Create item
- [ ] GET /items/:id - Get item details
- [ ] GET /items - List items with cursor pagination (hx-get for infinite scroll)
- [ ] PUT /items/:id - Update item
- [ ] DELETE /items/:id - Delete item
- [ ] Write integration tests

### Handlers - Outfits
**Files**: `backend/transport/http/handlers/outfit.go`

- [ ] POST /outfits - Create outfit
- [ ] GET /outfits/:id - Get outfit details
- [ ] GET /outfits - List outfits
- [ ] PUT /outfits/:id - Update outfit (including item positions)
- [ ] DELETE /outfits/:id - Delete outfit
- [ ] Write integration tests

### Handlers - Upload
**Files**: `backend/transport/http/handlers/upload.go`

- [ ] POST /upload - Multipart image upload
- [ ] GET /upload/:job_id - Get job status (JSON)
- [ ] GET /upload/:job_id/stream - SSE endpoint for real-time updates
- [ ] Write integration tests

### Handlers - Analytics
**Files**: `backend/transport/http/handlers/analytics.go`

- [ ] GET /analytics - Render dashboard with stats and charts
- [ ] Write integration tests

### Handlers - Sharing
**Files**: `backend/transport/http/handlers/share.go`

- [ ] POST /share - Create public outfit share (generates nanoid slug)
- [ ] GET /s/:slug - Public read-only outfit view
- [ ] Increment view count on access
- [ ] Write integration tests

### Server Setup
**Files**: `backend/transport/http/server.go`

- [ ] Setup routes with go-pkgz/routegroup
- [ ] Add middleware chain (logging → recovery → auth → ratelimit)
- [ ] Serve static files with go:embed (static/css, static/js, static/icons)
- [ ] Add health check endpoint: GET /health
- [ ] Implement graceful shutdown (30 second timeout for in-flight requests)

---

## Phase 6: Templates (Templ)

### Layouts
**Files**: `backend/transport/http/views/layouts/*.templ`

- [ ] Create base.templ - HTML structure, head, body wrapper
- [ ] Create components/head.templ - Meta tags, CSS links, PWA manifest
- [ ] Create components/header.templ - Navigation bar
- [ ] Create components/footer.templ - Footer content

### Pages
**Files**: `backend/transport/http/views/pages/*.templ`

- [ ] Create home.templ - Dashboard with stats
- [ ] Create login.templ - OAuth login flow
- [ ] Create items.templ - Item grid with infinite scroll
- [ ] Create item_detail.templ - Item edit form
- [ ] Create outfits.templ - Outfit grid
- [ ] Create outfit_editor.templ - Svelte canvas integration
- [ ] Create analytics.templ - Charts and stats
- [ ] Create settings.templ - User preferences

### Components
**Files**: `backend/transport/http/views/components/*.templ`

- [ ] Create item_card.templ - Reusable item card
- [ ] Create outfit_card.templ - Reusable outfit card
- [ ] Create upload_form.templ - File upload with progress
- [ ] Create stats_widget.templ - Dashboard stat cards

---

## Phase 7: Frontend (Svelte Canvas)

### Setup
**Files**: `frontend/package.json`, `frontend/vite.config.js`

```bash
cd frontend
bun create svelte@latest .
bun install
```

- [ ] Initialize Bun project in frontend/
- [ ] Install Svelte + Vite + TypeScript
- [ ] Configure Vite to output single IIFE file:
  ```js
  build: {
    lib: {
      entry: 'src/OutfitCanvas.svelte',
      name: 'OutfitCanvas',
      fileName: 'outfit-canvas',
      formats: ['iife']
    },
    outDir: '../backend/static/js'
  }
  ```
- [ ] Setup TailwindCSS for Svelte component

### Svelte Components
**Files**: `frontend/src/*.svelte`

- [ ] Create OutfitCanvas.svelte - Main canvas component
  - Props: outfitId, items (from Templ data attributes)
  - State: canvasItems with x, y, z positions
  - Methods: addItem, saveOutfit (via htmx.ajax)
- [ ] Create Draggable.svelte - Draggable wrapper component
  - Props: x, y, z (two-way bound)
  - Events: mousedown/mousemove/mouseup for drag
- [ ] Build component: `bun run build`
- [ ] Verify output: `backend/static/js/outfit-canvas.js`

### Static Assets
**Files**: `backend/static/js/*.js`

- [ ] Download htmx.min.js (https://unpkg.com/htmx.org@2.0.0/dist/htmx.min.js)
- [ ] Download alpine.min.js (https://unpkg.com/alpinejs@3.14.0/dist/cdn.min.js)
- [ ] Download apexcharts.min.js (for analytics charts)
- [ ] Create app.js - Alpine components (dropdowns, modals, toasts)

---

## Phase 8: PWA

### Manifest & Service Worker
**Files**: `backend/static/manifest.json`, `backend/static/sw.js`

- [ ] Create manifest.json with app metadata:
  - name, short_name, description
  - icons (192x192, 512x512)
  - start_url, display: standalone
  - theme_color, background_color
- [ ] Create service worker (sw.js):
  - Cache static assets (CSS, JS, icons)
  - Implement offline upload queue using IndexedDB
  - Implement Background Sync API for queued uploads
  - Cache API responses with stale-while-revalidate
  - Handle offline/online transitions

### Push Notifications
**Files**: `backend/cmd/genvapid/main.go`, `backend/transport/http/handlers/push.go`

- [ ] Create VAPID key generator utility (cmd/genvapid)
- [ ] Implement push subscription endpoint (POST /push/subscribe)
- [ ] Implement push unsubscribe endpoint (POST /push/unsubscribe)
- [ ] Send push notification on AI job completion (from worker)
- [ ] Deep link to processed item in notification

### Icons
**Files**: `backend/static/icons/*.png`

- [ ] Generate PWA icon 192x192
- [ ] Generate PWA icon 512x512
- [ ] Generate favicon.ico
- [ ] Generate Apple Touch icons (180x180)

---

## Phase 9: Testing

### Unit Tests
**Target**: 90%+ service layer, 80%+ overall

- [ ] Write tests for user service (CreateUser, GetUser, DeleteUser)
- [ ] Write tests for item service (CRUD + pagination)
- [ ] Write tests for outfit service (CRUD + item management)
- [ ] Write tests for analytics service (stats calculations)
- [ ] Write tests for AI providers (mock HTTP responses)
- [ ] Write tests for storage (mock S3 client)
- [ ] Write tests for image processing (EXIF stripping, validation)
- [ ] Run: `task test:unit` - Verify 90%+ coverage

### Integration Tests
**Target**: 70%+ handler coverage

- [ ] Write tests for item handlers (full request/response cycle)
- [ ] Write tests for outfit handlers
- [ ] Write tests for upload handler with SSE
- [ ] Write tests for analytics handler
- [ ] Write tests for sharing handlers
- [ ] Use pgtestdb for isolated test databases
- [ ] Run: `task test:integration`

### E2E Tests (Playwright)
**Files**: `tests/e2e/*.go`

- [ ] Create auth_test.go - OAuth login flow
- [ ] Create items_test.go - Upload → Edit → Save item
- [ ] Create outfits_test.go - Create outfit → Drag items → Save
- [ ] Create analytics_test.go - View dashboard with charts
- [ ] Create ai_test.go - Background removal with SSE updates
- [ ] Create pwa_test.go - Offline mode, background sync, push notifications
- [ ] Run: `task test:e2e`

---

## Phase 10: Documentation & Polish

### Documentation
**Files**: `README.md`, `docs/*.md`

- [ ] Add screenshots to README (homepage, wardrobe, outfit builder)
- [ ] Update quickstart guide
- [ ] Write deployment guide (Docker, NixOS)
- [ ] Generate OpenAPI docs from handler annotations
- [ ] Setup Swagger UI endpoint (GET /swagger)

### Database Optimization
**Files**: Migration files

- [ ] Add index: items(user_id, created_at DESC)
- [ ] Add index: outfits(user_id, created_at DESC)
- [ ] Add index: wear_history(user_id, worn_date DESC)
- [ ] Create materialized view for analytics (daily stats aggregation)
- [ ] Setup hourly refresh for materialized views

### Performance
- [ ] Test with 1000+ items per user
- [ ] Optimize slow queries
- [ ] Add query plan analysis for common operations
- [ ] Test upload concurrency (10 simultaneous uploads)

---

## Phase 11: Deployment

### Docker
**Files**: `Dockerfile`, `docker-compose.prod.yml`

- [ ] Create production Dockerfile (multi-stage build)
- [ ] Create docker-compose.prod.yml with secrets management
- [ ] Test Docker deployment locally
- [ ] Document Docker deployment in docs/

### Nix
**Files**: `flake.nix` (NixOS module)

- [ ] Create NixOS module (services.gothreads)
- [ ] Add options for database, storage, AI sidecars
- [ ] Test NixOS deployment
- [ ] Document NixOS deployment

### CI/CD
**Files**: `.gitlab-ci.yml`

- [ ] Create GitLab CI pipeline:
  - Stage 1: Validate (lint, format check)
  - Stage 2: Test (unit, integration, E2E)
  - Stage 3: Build (Nix build, Docker container)
  - Stage 4: Deploy (staging auto, production manual)
- [ ] Test pipeline
- [ ] Document CI/CD setup

---

## Phase 12: Launch Preparation

### Security
- [ ] Add CSP headers (Content-Security-Policy)
- [ ] Security audit (OWASP top 10 check)
- [ ] GDPR compliance review (data export, deletion)
- [ ] Penetration testing (basic vulnerability scan)

### Polish
- [ ] Beta testing with 5-10 users
- [ ] Bug fixes from beta feedback
- [ ] Performance testing under load
- [ ] Mobile responsiveness testing

### Release
- [ ] Write release notes for v1.0.0
- [ ] Tag release in Git
- [ ] Publish to GitHub/GitLab
- [ ] Announce on r/selfhosted
- [ ] Create demo video/screenshots

---

## Metadata

**Tech Stack**:
- Backend: Go, HTMX, Templ, PostgreSQL, SeaweedFS
- Frontend: HTMX, Alpine.js, Svelte (canvas only), TailwindCSS
- AI: rembg (local), OpenAI Vision (optional)
- Auth: OAuth2/OIDC (Authelia, Authentik)
- Build: Nix flakes, gomod2nix, Bun

**Key Patterns**:
- Hybrid frontend (HTMX + Alpine for most UI, Svelte for canvas)
- Cursor-based pagination
- SSE for real-time updates
- Strategy pattern for AI providers
- PWA with full offline mode

**Reference Projects**:
- go-routinely (Go patterns, migrations, testing)
- Stylebook, Whering, Twelve70 (UX inspiration)

---

## Phase 13: Pending Features

### Virtual Try-On (VTON) Feature - HIGH PRIORITY
**Status**: Partially implemented but needs fixing

**Current State**:
- Cloud backend: HuggingFace Space (ajay-projects/vton-backend) works but frequently out of GPU quota
- Local fallback: CatVTON Docker service works but very slow (~5 minutes on CPU)
- Progress updates: Implemented (every 10 seconds)
- NSFW placeholder: Created in Docker build process

**Tasks**:
- [ ] Find reliable free HuggingFace Space for VTON with available GPU quota
- [ ] OR: Implement FASHN paid API integration ($0.075/image, 10 free credits)
  - API docs: https://docs.fashn.ai/
  - Endpoint: POST to FASHN API with model_name='tryon-v1.6'
  - Add FASHN_API_KEY to environment config
- [ ] OR: Deploy own VTON Space to HuggingFace account with personal GPU quota
- [ ] OR: Optimize local CatVTON (add GPU support, reduce inference steps)
- [ ] Add better progress feedback (show inference steps, estimated time remaining)
- [ ] Add retry mechanism for quota errors with exponential backoff
- [ ] Consider implementing queue system for batch processing when quota limited

**Files to modify**:
- `mockups/api/main.go` - callVtonBackend(), runTryOnJob()
- `mockups/api/.env.example` - Add FASHN_API_KEY if implementing paid API
- `mockups/js/app.js` - handleTryOn(), potentially add queue UI

### Upload Full Outfit & Auto-Segmentation - MEDIUM PRIORITY
**Status**: Not implemented

**Feature Description**:
Upload a single photo of a complete outfit (on model, mannequin, or flat lay) and automatically segment it into individual clothing pieces (top, bottom, shoes, accessories, etc.).

**Tasks**:
- [ ] Research clothing segmentation models:
  - YOLO-based fashion detection models
  - Detectron2 with fashion dataset
  - Fashionpedia dataset/models
  - ModaNet dataset/models
  - DeepFashion2 segmentation models
- [ ] Choose segmentation approach:
  - Option A: Use HuggingFace Space with fashion segmentation model
  - Option B: Deploy local model (CPU or GPU)
  - Option C: Use third-party API (Clarifai, Google Vision, etc.)
- [ ] Implement segmentation API endpoint:
  - POST /api/segment-outfit
  - Input: Full outfit image
  - Output: Array of detected items with bounding boxes and categories
- [ ] Update upload.html UI:
  - Add "Upload Full Outfit" mode toggle
  - Show detected pieces preview with category labels
  - Allow user to adjust bounding boxes
  - Allow user to remove/add detected pieces
  - Show category dropdown for each piece
- [ ] Implement cropping and processing pipeline:
  - Extract each detected piece using bounding box
  - Run background removal on each piece
  - Generate AI tags and description for each piece
  - Store outfit_set_id to link related items
- [ ] Add batch item creation endpoint:
  - POST /api/items/batch
  - Create multiple items in single transaction
  - Return array of created item IDs
- [ ] Auto-create outfit from uploaded pieces:
  - Option to automatically create outfit with all detected pieces
  - Set item positions based on bounding box coordinates
- [ ] Add outfit set management:
  - Display items that are part of the same uploaded outfit
  - Filter wardrobe by outfit set
  - "View Original Photo" option to see full outfit

**Files to create/modify**:
- `mockups/api/main.go` - Add segmentOutfit(), batchCreateItems() endpoints
- `mockups/upload.html` - Add upload mode toggle and preview UI
- `mockups/js/app.js` - Add outfit segmentation handling
- `mockups/css/style.css` - Add bounding box preview styles
- Consider adding: `mockups/segment-outfit/` - Segmentation service (similar to catvton/)

**Database schema additions**:
```sql
-- Add outfit_set_id to items table
ALTER TABLE items ADD COLUMN outfit_set_id UUID;
ALTER TABLE items ADD COLUMN original_outfit_image_url TEXT;
CREATE INDEX idx_items_outfit_set ON items(outfit_set_id);
```

**Tech considerations**:
- Segmentation latency: 5-30 seconds depending on model
- May need GPU for reasonable performance
- Consider queue system for batch processing
- Edge cases: overlapping items, accessories, partial occlusion
- Quality threshold for detection confidence (e.g., only accept >0.7 confidence)
