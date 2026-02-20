# Requirements: go-threads

**Defined:** 2026-02-20
**Core Value:** Privacy-first wardrobe management with local AI processing - users can leverage powerful AI features like virtual try-on and automated styling without sending personal photos or shopping habits to the cloud

## v1 Requirements

Requirements for production-ready release. Each maps to roadmap phases.

### Architecture & Backend Foundation

- [ ] **ARCH-01**: Refactor monolithic main.go (3,637 lines) into proper service layer architecture
- [ ] **ARCH-02**: Implement proper Go project structure (cmd/, service/, store/, transport/)
- [ ] **ARCH-03**: Replace prototype database code with production SQLC generated queries
- [ ] **ARCH-04**: Implement structured logging with slog (JSON output, context-aware)
- [ ] **ARCH-05**: Add comprehensive error handling with domain-specific error types
- [ ] **ARCH-06**: Implement graceful shutdown with 30-second timeout for AI jobs

### Authentication & Security

- [ ] **AUTH-01**: Replace hardcoded user bypasses with proper OAuth2/OIDC authentication
- [ ] **AUTH-02**: Implement JWT validation with JWKS endpoint support
- [ ] **AUTH-03**: Add support for multiple OIDC providers (Authelia, Authentik, generic)
- [ ] **AUTH-04**: Implement proper authorization middleware for all endpoints
- [ ] **AUTH-05**: Add file upload validation (type, size, magic number checking)
- [ ] **AUTH-06**: Implement EXIF data stripping for privacy
- [ ] **AUTH-07**: Add Content Security Policy headers

### Item Management (Core Wardrobe)

- [ ] **ITEM-01**: User can upload clothing item images with drag-drop interface
- [ ] **ITEM-02**: System automatically removes backgrounds from uploaded images
- [ ] **ITEM-03**: AI automatically extracts tags (category, color, brand) from images
- [ ] **ITEM-04**: User can edit item details (name, price, category, size, material)
- [ ] **ITEM-05**: User can delete items with confirmation
- [ ] **ITEM-06**: User can view item grid with infinite scroll pagination
- [ ] **ITEM-07**: User can search and filter items by category, color, brand
- [ ] **ITEM-08**: System tracks wear count automatically when outfits are logged

### Outfit Management

- [ ] **OUTF-01**: User can create new outfits by selecting items from wardrobe
- [ ] **OUTF-02**: User can drag and position items on interactive canvas (Svelte component)
- [ ] **OUTF-03**: User can save outfit with custom name and notes
- [ ] **OUTF-04**: User can edit existing outfits (add/remove/reposition items)
- [ ] **OUTF-05**: User can delete outfits with confirmation
- [ ] **OUTF-06**: User can view outfit gallery with thumbnail previews
- [ ] **OUTF-07**: AI can suggest outfits based on weather and occasion
- [ ] **OUTF-08**: AI can rate outfit compatibility and provide feedback

### Analytics & Insights

- [ ] **ANAL-01**: User can view dashboard with wardrobe statistics (total items, outfits, spending)
- [ ] **ANAL-02**: System calculates cost-per-wear for each item
- [ ] **ANAL-03**: User can view wear frequency charts (30/90/365 days)
- [ ] **ANAL-04**: User can see category breakdown with interactive pie charts
- [ ] **ANAL-05**: User can view most/least worn items rankings
- [ ] **ANAL-06**: System identifies wardrobe gaps and suggests essential items

### AI Processing & Queue

- [ ] **AI-01**: Implement async AI job queue with PostgreSQL backend
- [ ] **AI-02**: Support multiple AI providers (local Ollama, cloud APIs) with configurable fallback
- [ ] **AI-03**: Provide real-time progress updates via Server-Sent Events
- [ ] **AI-04**: Implement exponential backoff retry for failed AI jobs (1min, 2min, 4min, 8min)
- [ ] **AI-05**: Queue AI jobs with priority levels and concurrency limits
- [ ] **AI-06**: Send push notifications when long-running AI jobs complete

### Storage & File Management

- [ ] **STOR-01**: Organize S3 storage with proper key structure (users/{id}/items/{type}/{uuid})
- [ ] **STOR-02**: Store both raw and processed versions of uploaded images
- [ ] **STOR-03**: Implement file cleanup when items/users are deleted
- [ ] **STOR-04**: Support pre-signed URLs for direct S3 uploads
- [ ] **STOR-05**: Implement upload concurrency limits (10 per user)

### Frontend & User Experience

- [ ] **UI-01**: Convert HTML mockups to production Templ templates
- [ ] **UI-02**: Implement HTMX-driven interface with progressive enhancement
- [ ] **UI-03**: Create responsive design that works on mobile and desktop
- [ ] **UI-04**: Implement Alpine.js components for dropdowns, modals, toasts
- [ ] **UI-05**: Build Svelte outfit canvas component with drag-drop functionality
- [ ] **UI-06**: Add loading states and progress indicators for all operations
- [ ] **UI-07**: Implement offline-capable PWA with service worker

## v2 Requirements

Deferred to future release. Tracked but not in current roadmap.

### Advanced AI Features

- **AI-ADV-01**: Virtual try-on with person photos and clothing items
- **AI-ADV-02**: Bulk outfit segmentation from full outfit photos
- **AI-ADV-03**: Voice input for natural wardrobe interaction
- **AI-ADV-04**: AI outfit generation based on style preferences
- **AI-ADV-05**: Duplicate item detection with perceptual hashing

### Social & Sharing

- **SHAR-01**: Public outfit sharing with anonymous view links
- **SHAR-02**: Outfit export as PNG images with watermark
- **SHAR-03**: Wardrobe data export (JSON, CSV, PDF formats)

### Mobile & Hardware Integration  

- **MOB-01**: Native camera integration for in-app photo capture
- **MOB-02**: Weather-based outfit suggestions with location services
- **MOB-03**: Barcode scanning for product information lookup

## Out of Scope

Explicitly excluded. Documented to prevent scope creep.

| Feature | Reason |
|---------|--------|
| Multi-tenancy | Self-hosted single-user/family focus keeps complexity manageable |
| E-commerce integration | Focus on inventory management, not shopping |
| Real-time chat/messaging | Not core to wardrobe management value prop |
| Native mobile apps | PWA provides mobile experience without platform complexity |
| Cloud-first AI processing | Conflicts with privacy-first core value |
| Social networking features | Personal wardrobe management, not social platform |
| Video content support | Image-focused workflow keeps storage simple |

## Traceability

Which phases cover which requirements. Updated during roadmap creation.

| Requirement | Phase | Status |
|-------------|-------|--------|
| ARCH-01 | Phase 1 | Pending |
| ARCH-02 | Phase 1 | Pending |
| ARCH-03 | Phase 1 | Pending |
| ARCH-04 | Phase 1 | Pending |
| ARCH-05 | Phase 1 | Pending |
| ARCH-06 | Phase 1 | Pending |
| AUTH-01 | Phase 2 | Pending |
| AUTH-02 | Phase 2 | Pending |
| AUTH-03 | Phase 2 | Pending |
| AUTH-04 | Phase 2 | Pending |
| AUTH-05 | Phase 2 | Pending |
| AUTH-06 | Phase 2 | Pending |
| AUTH-07 | Phase 2 | Pending |
| AI-01 | Phase 3 | Pending |
| AI-02 | Phase 3 | Pending |
| AI-03 | Phase 3 | Pending |
| AI-04 | Phase 3 | Pending |
| AI-05 | Phase 3 | Pending |
| AI-06 | Phase 3 | Pending |
| STOR-01 | Phase 4 | Pending |
| STOR-02 | Phase 4 | Pending |
| STOR-03 | Phase 4 | Pending |
| STOR-04 | Phase 4 | Pending |
| STOR-05 | Phase 4 | Pending |
| ITEM-01 | Phase 5 | Pending |
| ITEM-02 | Phase 5 | Pending |
| ITEM-03 | Phase 5 | Pending |
| ITEM-04 | Phase 5 | Pending |
| ITEM-05 | Phase 5 | Pending |
| ITEM-06 | Phase 5 | Pending |
| ITEM-07 | Phase 5 | Pending |
| ITEM-08 | Phase 5 | Pending |
| OUTF-01 | Phase 5 | Pending |
| OUTF-02 | Phase 5 | Pending |
| OUTF-03 | Phase 5 | Pending |
| OUTF-04 | Phase 5 | Pending |
| OUTF-05 | Phase 5 | Pending |
| OUTF-06 | Phase 5 | Pending |
| OUTF-07 | Phase 5 | Pending |
| OUTF-08 | Phase 5 | Pending |
| UI-01 | Phase 6 | Pending |
| UI-02 | Phase 6 | Pending |
| UI-03 | Phase 6 | Pending |
| UI-04 | Phase 6 | Pending |
| UI-05 | Phase 6 | Pending |
| UI-06 | Phase 6 | Pending |
| UI-07 | Phase 6 | Pending |
| ANAL-01 | Phase 6 | Pending |
| ANAL-02 | Phase 6 | Pending |
| ANAL-03 | Phase 6 | Pending |
| ANAL-04 | Phase 6 | Pending |
| ANAL-05 | Phase 6 | Pending |
| ANAL-06 | Phase 6 | Pending |

**Coverage:**
- v1 requirements: 43 total
- Mapped to phases: 43  
- Unmapped: 0 ✓

---
*Requirements defined: 2026-02-20*
*Last updated: 2026-02-20 after research-informed definition*