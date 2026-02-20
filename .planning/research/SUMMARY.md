# Project Research Summary

**Project:** GoThreads - Digital Wardrobe Management with AI
**Domain:** Privacy-first personal wardrobe management with AI integration
**Researched:** 2026-02-20
**Confidence:** HIGH

## Executive Summary

GoThreads is a self-hosted digital wardrobe management application that prioritizes user privacy through local AI processing. The recommended approach centers on a Go backend with PostgreSQL database, HTMX for server-driven UI, and Ollama for local AI processing (background removal, auto-tagging, virtual try-on). This architecture delivers the performance users expect while maintaining complete data ownership.

The most critical challenge is managing AI processing complexity — services like background removal can take 5+ minutes without proper GPU acceleration, making user experience unacceptable. The solution is implementing an asynchronous job queue with real-time progress updates via Server-Sent Events, combined with cloud AI fallbacks for better performance. Secondary risks include the current monolithic prototype architecture and image storage organization that must be addressed before scaling.

The research strongly supports a phase-based approach that establishes solid foundations (proper Go project structure, database migrations, authentication) before tackling complex AI integration and advanced features like virtual try-on.

## Key Findings

### Recommended Stack

The Go ecosystem provides excellent foundations for this domain, with PostgreSQL offering the reliability needed for wardrobe data and HTMX enabling server-driven UI without SPA complexity. The critical insight is using local AI as the primary approach with cloud fallbacks — this differentiates GoThreads from cloud-first competitors while providing performance escape hatches.

**Core technologies:**
- Go 1.26.0: Backend API with excellent concurrency for AI processing — mature ecosystem and security updates
- PostgreSQL 18.x: Primary database with ACID compliance — essential for transaction integrity and JSON support
- HTMX 2.0.x: Frontend interactivity — server-driven approach reduces JS complexity, perfect for form-heavy apps
- Templ 0.3.x: Type-safe HTML templates — compile-time safety with perfect HTMX integration
- Ollama: Local AI processing — core privacy value proposition with vision models for tagging and virtual try-on

### Expected Features

Fashion management apps have well-established user expectations, with cost-per-wear analytics being the key financial value proposition. The research identified strong consensus on table stakes while highlighting AI-powered features as primary differentiators.

**Must have (table stakes):**
- Item Photography & Upload — core function, users can't manage wardrobes without items
- Outfit Creation & Planning — combine items into outfits for coordinated looks
- Wear Tracking & Analytics — track usage for cost-per-wear calculations
- Basic Search & Filter — find items in growing collections
- Mobile-Responsive PWA — fashion is primarily mobile activity

**Should have (competitive):**
- Local AI Background Removal — professional catalog appearance with privacy
- AI Auto-Tagging — automatic metadata extraction (color, style, brand)
- Voice Input Interface — natural language wardrobe interaction
- Weather-Based Suggestions — context-aware outfit recommendations
- Self-Hosted Deployment — complete data ownership and privacy control

**Defer (v2+):**
- Virtual Try-On — high complexity, needs GPU optimization
- Bulk Upload & Segmentation — complex UX and specialized models
- Social/Community Features — conflicts with privacy-first positioning

### Architecture Approach

The hybrid frontend pattern (HTMX + targeted JavaScript) optimally balances development speed with user experience. The service layer pattern enables proper testing and the provider strategy for AI services delivers the local-first flexibility that differentiates this product.

**Major components:**
1. HTMX Frontend — server-driven updates with minimal JavaScript complexity
2. Service Layer — business logic separated from HTTP transport for testability
3. AI Provider System — pluggable local/cloud providers with fallback capabilities
4. Async Job Queue — background AI processing with real-time progress updates
5. S3-Compatible Storage — organized file structure with proper lifecycle management

### Critical Pitfalls

The research identified seven critical pitfalls that have historically killed similar projects, with AI service reliability and architectural technical debt being the highest priority risks.

1. **AI Service Quota Hell** — free-tier AI services hit limits causing silent failures; implement robust queue system with local fallbacks
2. **Monolithic Prototype Architecture** — 3,637-line main.go becomes unmaintainable; refactor to service layer before adding features
3. **Runtime Schema Changes** — database modifications at startup can corrupt data; use proper migration tools from start
4. **Image Storage Chaos** — unorganized S3 structure makes cleanup impossible; design clear key structure: `users/{id}/items/{type}/{uuid}.ext`
5. **Authentication Bypass** — hardcoded user IDs reach production; implement proper JWT/OIDC before multi-user features

## Implications for Roadmap

Based on research, suggested phase structure prioritizes foundational stability before complex AI features:

### Phase 1: Backend Foundation
**Rationale:** Current prototype architecture must be refactored before adding features — monolithic structure creates cascading failures
**Delivers:** Proper Go project structure, service layer separation, database migration system
**Addresses:** Basic Item CRUD, Search & Filter from table stakes
**Avoids:** Monolithic Architecture pitfall and Runtime Schema Changes

### Phase 2: Authentication & HTTP Layer  
**Rationale:** Multi-user features require solid auth foundation — security bypasses are critical failures
**Delivers:** JWT authentication, HTMX-driven UI templates, middleware stack
**Uses:** Templ templates with HTMX for server-driven interactions
**Implements:** HTTP transport layer with proper error handling

### Phase 3: AI Integration & Processing
**Rationale:** AI features are core differentiator but require complex queue system — must handle long processing times gracefully
**Delivers:** Ollama integration, background job queue, real-time progress updates via SSE
**Addresses:** AI Background Removal, AI Auto-Tagging from differentiators
**Avoids:** AI Service Quota Hell and VTON Performance Issues pitfalls

### Phase 4: Storage & Upload System
**Rationale:** File organization must be designed properly from start — reorganizing later requires expensive data migrations
**Delivers:** S3-compatible storage (SeaweedFS), organized key structure, upload validation
**Uses:** Structured S3 keys for user isolation and cleanup
**Avoids:** Image Storage Chaos and Upload Abuse pitfalls

### Phase 5: Core Wardrobe Features
**Rationale:** With solid foundations, can implement core table stakes features users expect
**Delivers:** Outfit creation, wear tracking, cost-per-wear analytics
**Implements:** Calendar integration and analytics dashboard components

### Phase 6: Advanced Features (v1.x)
**Rationale:** After validation of core concept, add competitive differentiators
**Delivers:** Weather suggestions, voice input, outfit sharing
**Uses:** Established AI provider system for new capabilities

### Phase Ordering Rationale

- **Foundation first approach:** Architecture refactoring must precede feature development to avoid cascading failures
- **Security before features:** Authentication system required before any multi-user functionality to prevent data breaches  
- **AI complexity isolated:** Background processing system established before complex AI features to manage performance expectations
- **Storage organization upfront:** File structure changes are expensive once users have data — design correctly from start

### Research Flags

Phases likely needing deeper research during planning:
- **Phase 3 (AI Integration):** Complex integration with Ollama, queue system design patterns need specific research
- **Phase 6 (Advanced Features):** Weather API integration and voice processing need API research

Phases with standard patterns (skip research-phase):
- **Phase 1 (Backend Foundation):** Go project structure is well-established with clear patterns
- **Phase 2 (HTTP Layer):** HTMX + Templ patterns are well-documented
- **Phase 4 (Storage System):** S3 integration follows standard Go patterns

## Confidence Assessment

| Area | Confidence | Notes |
|------|------------|-------|
| Stack | HIGH | Based on established Go ecosystem, existing prototype validation |
| Features | HIGH | Strong domain knowledge from existing project context and competitor analysis |
| Architecture | HIGH | Standard patterns for Go web applications, proven HTMX approaches |
| Pitfalls | HIGH | Direct experience from prototype codebase analysis, industry best practices |

**Overall confidence:** HIGH

### Gaps to Address

Research was comprehensive but some areas need validation during implementation:

- **GPU Performance Requirements:** Local AI processing speed needs real-world testing with target hardware configurations
- **Storage Cost Modeling:** S3-compatible storage costs at scale need validation for self-hosted users
- **Mobile PWA Patterns:** Offline capabilities and background sync need specific research during Phase 5 implementation

## Sources

### Primary (HIGH confidence)
- Existing prototype codebase analysis (mockups/api/main.go, 3,637 lines) — Current architecture and pain points
- Project documentation (SPEC.md, AI_FEATURES.md, PROJECT.md) — Requirements and feature specifications
- Go 1.26 Release Notes and PostgreSQL 18 documentation — Technology compatibility verification

### Secondary (MEDIUM confidence)  
- Competitive analysis of Stylebook, Whering, Twelve70 — Feature landscape and user expectations
- Reddit r/selfhosted community analysis — Privacy-focused user requirements
- Production Go applications survey 2025 — Library adoption patterns

### Tertiary (LOW confidence)
- AI processing performance estimates — Based on Ollama documentation, needs real-world validation
- Storage scaling assumptions — Based on typical self-hosted usage patterns

---
*Research completed: 2026-02-20*
*Ready for roadmap: yes*