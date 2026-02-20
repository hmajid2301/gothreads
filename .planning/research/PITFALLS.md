# Domain Pitfalls

**Domain:** Digital Wardrobe Management + AI Integration + Self-Hosted
**Researched:** 2026-02-20
**Confidence:** HIGH

## Critical Pitfalls

### Pitfall 1: AI Service Quota Hell

**What goes wrong:**
Production breaks when free-tier cloud AI services hit quota limits, causing background removal, virtual try-on, or tagging to fail silently. Users upload items that get stuck in "processing" state forever, or worse, appear to process but actually fail.

**Why it happens:**
- HuggingFace Spaces have GPU quotas that reset unpredictably
- Free tiers seem generous during development but break under real usage
- Prototype code often lacks proper fallback handling
- Local AI alternatives are too slow for production (5+ minutes per image)

**How to avoid:**
- Implement queue-based processing with retry logic and exponential backoff
- Always have local fallback providers configured (even if slower)
- Monitor quota usage and implement circuit breakers
- Notify users when processing will be delayed vs. failed
- Consider paid AI service budget from day one

**Warning signs:**
- Background removal works in dev but fails intermittently in production
- Processing jobs stuck in "pending" state for hours
- Users reporting images not being processed
- High error rates during peak usage times
- AI service returning 402/429 status codes

**Phase to address:**
Phase 3 (AI Integration) - Implement robust provider pattern with fallbacks

---

### Pitfall 2: Monolithic Prototype Architecture

**What goes wrong:**
The 3,637-line main.go file becomes unmaintainable, causing merge conflicts, making testing impossible, and creating cascading failures when one component breaks.

**Why it happens:**
- "It's just a prototype" mentality leads to cramming everything in one file
- Quick iteration encourages shortcuts that become permanent
- Each new feature gets bolted onto the existing monolith
- Separation feels like over-engineering when building quickly

**How to avoid:**
- Refactor to proper service layer before adding more features
- Separate concerns: handlers, services, storage, AI providers
- Use dependency injection to enable proper testing
- Create clear package boundaries from the start

**Warning signs:**
- Single file over 1000 lines
- Functions doing multiple unrelated things
- Direct database calls mixed with business logic
- Impossible to test components in isolation
- "God objects" that know about everything

**Phase to address:**
Phase 2 (Backend Foundation) - Architectural refactoring must happen first

---

### Pitfall 3: Data Loss from Runtime Schema Changes

**What goes wrong:**
Database schema modifications at startup (lines 34-73 in db.go) can corrupt existing data or fail to apply, causing the entire system to become unusable with no rollback capability.

**Why it happens:**
- Prototypes often modify schema directly in startup code
- Migration tools feel like overkill for small changes
- No testing of schema changes against production-like data
- Assumption that database is always empty/fresh

**How to avoid:**
- Use proper migration tools (goose, migrate) from the start
- Version all schema changes
- Test migrations against realistic data sets
- Always have rollback procedures
- Never modify schema in application startup code

**Warning signs:**
- Schema changes mixed with application code
- No migration version tracking
- "DROP TABLE IF EXISTS" statements in production code
- Schema changes that require data manipulation
- No database backup/restore procedures

**Phase to address:**
Phase 1 (Backend Foundation) - Proper database migration setup

---

### Pitfall 4: Image Storage Chaos

**What goes wrong:**
Unorganized S3 storage structure makes it impossible to clean up deleted items, causes permission issues, and creates billing nightmares when storage grows uncontrolled.

**Why it happens:**
- Flat storage structure seems simpler initially
- No consideration for multi-user file isolation
- Raw and processed images mixed together
- No lifecycle management or cleanup procedures

**How to avoid:**
- Design clear S3 key structure from day one: `users/{user_id}/items/{type}/{uuid}.{ext}`
- Implement cleanup procedures for deleted items
- Separate raw and processed images in different paths
- Plan for GDPR-compliant user data deletion
- Set up S3 lifecycle policies for cost management

**Warning signs:**
- Files stored with random or date-based keys
- No way to find all files belonging to a user
- Processed images overwrite raw images
- Growing storage costs with no cleanup mechanism
- Permission errors when users try to access their files

**Phase to address:**
Phase 4 (Storage & Upload) - Proper S3 organization and lifecycle

---

### Pitfall 5: Authentication Bypass in Production

**What goes wrong:**
Hardcoded user ID bypasses (like `userID := int64(1) // TODO: auth` in line 3003) make it into production, causing massive security breaches where all users see each other's wardrobes.

**Why it happens:**
- Auth seems complex to implement during rapid prototyping
- Environment variables for auth skipping get enabled in production
- OAuth integration appears working but has subtle bypass paths
- Testing doesn't catch auth issues because "it works for the happy path"

**How to avoid:**
- Implement proper JWT/OIDC auth before any multi-user features
- Never allow auth bypass flags in production builds
- Use integration tests that verify auth on every endpoint
- Implement proper user context propagation throughout the app

**Warning signs:**
- Hardcoded user IDs in handlers
- `SKIP_AUTH=true` environment variables
- Auth middleware that can be bypassed
- No per-user data isolation in database queries
- Tests that don't verify authentication

**Phase to address:**
Phase 5 (HTTP Layer) - Authentication middleware must be solid

---

### Pitfall 6: Virtual Try-On Performance Disaster

**What goes wrong:**
Virtual try-on features work in demos but are unusably slow (5+ minutes) in production, causing user abandonment and server resource exhaustion.

**Why it happens:**
- VTON models are extremely compute-intensive
- CPU-only inference is too slow for real-time use
- No progress feedback leaves users thinking the app is broken
- Memory usage can crash servers with concurrent requests

**How to avoid:**
- Budget for GPU-enabled servers or cloud GPU services
- Implement proper job queuing with progress updates
- Set realistic user expectations about processing times
- Limit concurrent VTON operations to prevent server overload
- Consider paid AI services for better performance

**Warning signs:**
- VTON requests taking more than 60 seconds
- Server memory usage spiking during AI operations
- Users reporting app "freezing" during try-on
- CPU usage hitting 100% during image processing
- Multiple concurrent VTON requests crashing the server

**Phase to address:**
Phase 3 (AI Integration) - Performance optimization and queuing

---

### Pitfall 7: Image Upload Abuse

**What goes wrong:**
Users upload extremely large files (>50MB), inappropriate content, or malicious images that crash the processing pipeline or fill up storage.

**Why it happens:**
- No client-side or server-side file size validation
- No image format validation beyond MIME type checking
- No malware scanning for uploaded files
- No content moderation for inappropriate images

**How to avoid:**
- Implement strict file size limits (10MB max)
- Validate image formats using magic number checking
- Strip EXIF data for privacy
- Consider virus scanning for uploads
- Implement user upload quotas

**Warning signs:**
- Storage costs growing rapidly
- Image processing failing on certain uploads
- Server disk space filling up
- Processing pipeline crashing on specific images
- Users reporting slow upload performance

**Phase to address:**
Phase 4 (Storage & Upload) - Upload validation and security

---

## Technical Debt Patterns

Shortcuts that seem reasonable but create long-term problems.

| Shortcut | Immediate Benefit | Long-term Cost | When Acceptable |
|----------|-------------------|----------------|-----------------|
| Direct SQL in handlers | Faster to write | Unmaintainable queries, no type safety | Never - use SQLC from start |
| Global variables for config | Simple access | Race conditions, testing issues | Only during initial prototype |
| In-memory job tracking | No database complexity | Lost jobs on restart, no persistence | Demo only, never production |
| Hardcoded OAuth secrets | Quick testing | Security breach | Development only, never commit |
| Single database connection | Simple setup | Connection pool exhaustion | Never - use pgxpool |
| Mixed raw/processed S3 keys | Fewer folders | Impossible cleanup, data corruption | Never - separate from start |
| No error wrapping | Less boilerplate | Impossible debugging | Never - wrap all errors |
| Synchronous AI processing | Simpler code | UI blocks, poor UX | MVP only, add queue by Phase 3 |

## Integration Gotchas

Common mistakes when connecting to external services.

| Integration | Common Mistake | Correct Approach |
|-------------|----------------|------------------|
| HuggingFace Spaces | Assuming 100% uptime | Implement circuit breakers and local fallbacks |
| S3/SeaweedFS | No error handling for network issues | Retry with exponential backoff |
| PostgreSQL | Direct queries without connection pooling | Use pgxpool with proper limits |
| OAuth/OIDC | Trusting JWT without signature verification | Always verify with JWKS endpoint |
| Docker AI sidecars | Assuming they're always ready | Health checks before making requests |
| Ollama models | Not checking if models are pulled | Verify model availability in startup |
| Image processing | No timeout on AI operations | Always set context timeouts (30s-2min) |

## Performance Traps

Patterns that work at small scale but fail as usage grows.

| Trap | Symptoms | Prevention | When It Breaks |
|------|----------|------------|----------------|
| No database indexes | Slow queries after 1000+ items | Add indexes on user_id, created_at | ~1000 items per user |
| Synchronous AI processing | UI freezing during operations | Async queue with SSE updates | 2+ concurrent users |
| No image size limits | Server disk fills up | 10MB max file size | 100+ users uploading |
| Global job tracking map | Memory leaks, race conditions | Database-backed job queue | 10+ concurrent operations |
| No connection pooling | "Too many connections" errors | pgxpool with 50 max connections | 20+ concurrent users |
| Single-threaded uploads | File uploads block each other | Goroutine per upload with limits | 5+ concurrent uploads |
| No image caching | S3 costs skyrocket | CloudFlare/CDN with cache headers | 100+ daily active users |

## Security Mistakes

Domain-specific security issues beyond general web security.

| Mistake | Risk | Prevention |
|---------|------|------------|
| No user isolation in S3 keys | Users access each other's images | Structure: `users/{id}/items/{uuid}` |
| EXIF data not stripped | Location/device info leaked | Strip all EXIF before S3 upload |
| No image format validation | RCE via malicious images | Magic number + format validation |
| Auth bypass for development | Production deployment with SKIP_AUTH | Never allow auth skip in prod builds |
| No input validation on AI prompts | Prompt injection attacks | Sanitize all user inputs to AI |
| Hardcoded secrets in config | API keys in version control | Environment variables only |
| No rate limiting on uploads | Storage abuse | 10 uploads/hour per user limit |

## UX Pitfalls

Common user experience mistakes in this domain.

| Pitfall | User Impact | Better Approach |
|---------|-------------|-----------------|
| No progress feedback on AI processing | Users think app is broken | SSE with real-time updates |
| Synchronous background removal | UI freezes for 30+ seconds | Async with progress bar |
| No fallback for failed AI operations | Items get stuck "processing" | Show failure + retry option |
| Complex outfit builder on first use | Users abandon without creating outfits | Progressive disclosure, guided tour |
| No wear tracking automation | Users forget to log outfits | Calendar integration, wear reminders |
| Overwhelming analytics dashboard | Users ignore cost-per-wear insights | Simple widgets, progressive detail |
| No offline support | App useless without internet | PWA with offline item browsing |

## "Looks Done But Isn't" Checklist

Things that appear complete but are missing critical pieces.

- [ ] **AI Processing:** Often missing timeout handling — verify 30s timeouts on all AI calls
- [ ] **File Uploads:** Often missing size/format validation — verify 10MB limit + magic numbers
- [ ] **User Authentication:** Often missing proper JWT verification — verify JWKS endpoint checking
- [ ] **Database Queries:** Often missing user isolation — verify all queries filter by user_id
- [ ] **S3 Operations:** Often missing error handling — verify network failure retry logic
- [ ] **Background Jobs:** Often missing failure recovery — verify jobs resume after server restart
- [ ] **Image Processing:** Often missing EXIF stripping — verify no metadata in processed images
- [ ] **OAuth Integration:** Often missing state verification — verify CSRF protection
- [ ] **AI Provider Fallbacks:** Often missing local alternatives — verify CPU-only processing works
- [ ] **Data Cleanup:** Often missing user deletion cascade — verify GDPR-compliant data removal

## Recovery Strategies

When pitfalls occur despite prevention, how to recover.

| Pitfall | Recovery Cost | Recovery Steps |
|---------|---------------|----------------|
| Monolithic architecture | HIGH | Extract services incrementally, add tests, refactor handlers |
| No database migrations | HIGH | Export data, recreate schema with migrations, import with validation |
| Broken AI processing | MEDIUM | Implement queue system, add fallbacks, reprocess failed items |
| S3 storage chaos | MEDIUM | Script to reorganize keys, update database references, test access |
| Authentication bypass | HIGH | Audit all endpoints, add auth middleware, invalidate all sessions |
| Missing indexes | LOW | Add indexes with CONCURRENTLY, monitor query performance |
| Image processing failures | LOW | Requeue failed jobs, add timeout handling, notify users |

## Pitfall-to-Phase Mapping

How roadmap phases should address these pitfalls.

| Pitfall | Prevention Phase | Verification |
|---------|------------------|--------------|
| AI Service Quota Hell | Phase 3 (AI Integration) | Load test with quota limits, verify fallbacks work |
| Monolithic Architecture | Phase 2 (Backend Foundation) | Each service has >80% test coverage |
| Runtime Schema Changes | Phase 1 (Database Setup) | All schema changes via versioned migrations |
| Image Storage Chaos | Phase 4 (Storage & Upload) | User deletion removes all associated files |
| Authentication Bypass | Phase 5 (HTTP Layer) | All endpoints require valid JWT token |
| VTON Performance Issues | Phase 3 (AI Integration) | VTON completes in <2 minutes with progress |
| Upload Abuse | Phase 4 (Storage & Upload) | File validation rejects >10MB and invalid formats |

## Sources

- Project codebase analysis (`mockups/api/main.go`, concerns, implementation tasks)
- Digital wardrobe app domain knowledge (Stylebook, Whering patterns)
- Self-hosted application scaling patterns
- AI integration production challenges (HuggingFace quota limits, processing timeouts)
- Go web application best practices (SQLC, pgxpool, JWT handling)
- S3/object storage organization patterns for multi-user applications

---
*Pitfalls research for: Digital Wardrobe Management + AI Integration + Self-Hosted*
*Researched: 2026-02-20*