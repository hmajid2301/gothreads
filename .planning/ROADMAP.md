# Roadmap: go-threads

**Project:** Privacy-first digital wardrobe management with local AI processing  
**Generated:** 2026-02-20  
**Depth:** Quick (3-5 phases, critical path focus)  
**Coverage:** 43/43 v1 requirements mapped ✓

## Phases

- [ ] **Phase 1: Backend Foundation** - Refactor prototype into production-ready service architecture
- [ ] **Phase 2: Authentication & Core UI** - Implement OAuth2/OIDC auth and HTMX-driven templates
- [ ] **Phase 3: AI Processing & Storage** - Build async AI queue with local processing capabilities
- [ ] **Phase 4: Wardrobe Management** - Complete item/outfit management with analytics dashboard

## Phase Details

### Phase 1: Backend Foundation
**Goal**: Transform prototype into production-ready service architecture with proper database layer  
**Depends on**: Nothing (foundation phase)  
**Requirements**: ARCH-01, ARCH-02, ARCH-03, ARCH-04, ARCH-05, ARCH-06  
**Success Criteria** (what must be TRUE):
  1. Developer can add new features without touching main.go monolith
  2. Database schema changes deploy reliably through migration system
  3. Logs provide structured JSON output with request tracing
  4. Server handles graceful shutdown during AI processing jobs
**Plans**: TBD

### Phase 2: Authentication & Core UI  
**Goal**: Users can securely access the application through modern web interface  
**Depends on**: Phase 1 (service layer needed for auth middleware)  
**Requirements**: AUTH-01, AUTH-02, AUTH-03, AUTH-04, AUTH-05, AUTH-06, AUTH-07, UI-01, UI-02, UI-03, UI-04  
**Success Criteria** (what must be TRUE):
  1. User can authenticate via OAuth2/OIDC provider and stay logged in
  2. Application rejects unauthenticated requests with proper redirects
  3. File uploads validate type/size and strip EXIF data automatically
  4. UI responds fluidly on mobile and desktop with HTMX interactions
**Plans**: TBD

### Phase 3: AI Processing & Storage
**Goal**: AI services process wardrobe images efficiently with real-time progress feedback  
**Depends on**: Phase 2 (authentication needed for file upload security)  
**Requirements**: AI-01, AI-02, AI-03, AI-04, AI-05, AI-06, STOR-01, STOR-02, STOR-03, STOR-04, STOR-05  
**Success Criteria** (what must be TRUE):
  1. Background removal jobs complete without blocking UI interactions
  2. Users see live progress updates for AI processing via browser notifications
  3. System falls back to cloud AI when local processing fails or times out
  4. File storage organizes images with predictable cleanup capabilities
**Plans**: TBD

### Phase 4: Wardrobe Management
**Goal**: Users manage complete wardrobe lifecycle with outfit creation and analytics  
**Depends on**: Phase 3 (AI processing needed for auto-tagging, storage for images)  
**Requirements**: ITEM-01, ITEM-02, ITEM-03, ITEM-04, ITEM-05, ITEM-06, ITEM-07, ITEM-08, OUTF-01, OUTF-02, OUTF-03, OUTF-04, OUTF-05, OUTF-06, OUTF-07, OUTF-08, ANAL-01, ANAL-02, ANAL-03, ANAL-04, ANAL-05, ANAL-06, UI-05, UI-06, UI-07  
**Success Criteria** (what must be TRUE):
  1. User can photograph clothing item and see it cataloged with AI-generated tags
  2. User can create outfits by dragging items onto visual canvas and save them
  3. Dashboard shows cost-per-wear analytics and wardrobe usage patterns
  4. Application works offline and syncs when connection returns (PWA)
  5. System provides intelligent outfit suggestions based on weather and occasion
**Plans**: TBD

## Progress

| Phase | Plans Complete | Status | Completed |
|-------|----------------|--------|-----------|
| 1. Backend Foundation | 0/TBD | Not started | - |
| 2. Authentication & Core UI | 0/TBD | Not started | - |
| 3. AI Processing & Storage | 0/TBD | Not started | - |
| 4. Wardrobe Management | 0/TBD | Not started | - |

## Architecture Dependencies

```
Phase 1 (Foundation)
    ↓
Phase 2 (Auth + UI) ← needs service layer
    ↓  
Phase 3 (AI + Storage) ← needs auth for upload security
    ↓
Phase 4 (Wardrobe) ← needs AI processing + storage
```

## Research Context

Based on project research findings:
- **Foundation-first approach**: Current prototype architecture (3,637-line main.go) must be refactored before feature development
- **Security before features**: OAuth2/OIDC required before multi-user functionality  
- **AI complexity managed**: Async processing system essential for 5+ minute background removal jobs
- **Storage organization upfront**: S3 structure changes expensive once users have data

**Research flags** for deeper investigation during planning:
- Phase 3: Ollama integration patterns, queue system design
- Phase 4: Voice processing APIs, weather service integration

---
*Generated: 2026-02-20*  
*Next: `/gsd-plan-phase 1`*