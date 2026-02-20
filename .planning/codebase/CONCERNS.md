# Codebase Concerns

**Analysis Date:** 2026-02-20

## Tech Debt

**Authentication bypass hardcoded:**
- Issue: Hardcoded user ID `1` in chat handler bypasses authentication system
- Files: `mockups/api/main.go:3003`
- Impact: All chat requests execute as user ID 1, breaking multi-user functionality
- Fix approach: Implement proper JWT token extraction and user ID validation

**Runtime schema migrations:**
- Issue: Database schema changes executed at startup without proper migration control
- Files: `mockups/api/db.go:34-73`
- Impact: Schema drift, potential data loss, no rollback capability
- Fix approach: Move schema changes to proper migration files in `backend/store/db/migrations/`

**Monolithic main.go file:**
- Issue: 3,637 lines of code in single file mixing concerns
- Files: `mockups/api/main.go`
- Impact: Poor maintainability, difficult testing, merge conflicts
- Fix approach: Split into separate handlers, services, and middleware packages

**Dual database systems:**
- Issue: Both `mockups/api/` (direct SQL) and `backend/store/db/` (SQLC generated) exist
- Files: `mockups/api/db.go`, `backend/store/db/*.go`
- Impact: Code duplication, inconsistent data access patterns
- Fix approach: Standardize on SQLC-generated code, remove direct SQL in mockups

## Known Bugs

**Fatal server crash on startup failure:**
- Symptoms: Server exits completely if HTTP listener fails
- Files: `mockups/api/main.go:623`
- Trigger: Port already in use or permission denied
- Workaround: None - server must be restarted

**pgtype null handling inconsistency:**
- Symptoms: Generated SQLC code uses pgtype nullable types but mockup code doesn't handle them
- Files: `backend/store/db/models.go`, `mockups/api/db.go`
- Trigger: NULL database values cause type assertion failures
- Workaround: Ensure database constraints prevent NULLs

## Security Considerations

**Hardcoded secrets in config:**
- Risk: Default JWT secret exposed in code
- Files: `mockups/api/auth.go:60`
- Current mitigation: Environment variable override available
- Recommendations: Remove default, require explicit secret configuration

**Database credentials in docker-compose:**
- Risk: Plaintext database passwords in version control
- Files: `docker-compose.yml:26`, `config.yaml:46`
- Current mitigation: Development environment only
- Recommendations: Use environment variables or secrets management

**No input validation on API endpoints:**
- Risk: SQL injection, XSS, data corruption
- Files: `mockups/api/db.go` (various query methods)
- Current mitigation: Some pgx parameterization
- Recommendations: Add comprehensive input validation middleware

**CORS and authentication disabled by default:**
- Risk: Open access from any origin when SKIP_AUTH=true
- Files: `mockups/api/auth.go:59`
- Current mitigation: Development flag warning
- Recommendations: Require explicit production mode configuration

## Performance Bottlenecks

**No database connection pooling limits:**
- Problem: Unlimited connection pool growth
- Files: `mockups/api/db.go:21`
- Cause: pgxpool.New() with default settings
- Improvement path: Configure max connections, idle timeout

**No request rate limiting:**
- Problem: AI endpoint abuse potential
- Files: `mockups/api/main.go` (chat endpoint)
- Cause: No rate limiting middleware
- Improvement path: Add rate limiting per user/IP

**Large file uploads without streaming:**
- Problem: Image uploads load entirely into memory
- Files: `mockups/api/main.go` (multipart handling)
- Cause: Buffered multipart parsing
- Improvement path: Implement streaming upload to S3

## Fragile Areas

**OAuth state management:**
- Files: `mockups/api/auth.go:53-54`
- Why fragile: In-memory map without cleanup, not thread-safe
- Safe modification: Replace with database-backed state store
- Test coverage: None detected

**AI job status tracking:**
- Files: `mockups/api/main.go:248-273`
- Why fragile: Global map with complex locking, no persistence
- Safe modification: Use database or Redis for job state
- Test coverage: None detected

**Vendor directory in version control:**
- Files: `mockups/api/vendor/`
- Why fragile: Large vendor directory committed to git
- Safe modification: Add to .gitignore, use go modules properly
- Test coverage: Not applicable

## Scaling Limits

**Single-node file storage:**
- Current capacity: Local filesystem/SeaweedFS
- Limit: Disk space and I/O throughput
- Scaling path: Migrate to cloud object storage (S3, GCS)

**Synchronous AI processing:**
- Current capacity: One request at a time
- Limit: AI model processing time blocks other requests
- Scaling path: Queue-based async processing

## Dependencies at Risk

**Go 1.24.0 pre-release:**
- Risk: Using unreleased Go version
- Impact: Potential instability, build issues
- Migration plan: Downgrade to stable Go 1.23.x

**Direct database access without ORM:**
- Risk: Manual SQL management complexity
- Impact: Schema drift, query optimization issues
- Migration plan: Complete migration to SQLC or adopt proper ORM

## Missing Critical Features

**Request logging and monitoring:**
- Problem: No structured request logging
- Blocks: Debugging production issues, performance analysis

**Database migrations versioning:**
- Problem: No migration state tracking
- Blocks: Safe schema changes in production

**Graceful shutdown handling:**
- Problem: No signal handling for cleanup
- Blocks: Safe deployments, data consistency

## Test Coverage Gaps

**No automated tests found:**
- What's not tested: All API endpoints, database operations, auth flows
- Files: No test files detected
- Risk: Breaking changes undetected, regression bugs
- Priority: High - critical for production readiness

**No error handling tests:**
- What's not tested: Database failures, external service unavailability
- Files: `mockups/api/*.go`
- Risk: Poor error recovery in production
- Priority: High

**No concurrency safety tests:**
- What's not tested: Race conditions in job management, OAuth state
- Files: `mockups/api/main.go`, `mockups/api/auth.go`
- Risk: Data corruption under load
- Priority: Medium

---

*Concerns audit: 2026-02-20*