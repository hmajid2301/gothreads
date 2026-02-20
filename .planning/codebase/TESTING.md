# Testing Patterns

**Analysis Date:** 2026-02-20

## Test Framework

**Runner:**
- `gotestsum` with `testname` format
- Config: No explicit test config files found, uses Go defaults

**Assertion Library:**
- Standard Go testing package (no third-party assertion libraries detected)

**Run Commands:**
```bash
task test                    # Run all tests (unit + integration + e2e)
task test:unit               # Run unit tests with race detection
task test:integration        # Run integration tests  
task test:e2e               # Run E2E tests with verbose output
```

## Test File Organization

**Location:**
- Unit tests: Co-located with source code (`./service/...`, `./store/...`)
- Integration tests: In handlers directory (`./transport/http/handlers/...`)
- E2E tests: Separate directory (`../tests/e2e/...`)

**Naming:**
- `*_test.go` pattern (standard Go convention)
- Integration tests use build tag: `-tags=integration`

**Structure:**
```
backend/
├── service/         # Unit tests here
├── store/          # Unit tests here
├── transport/http/handlers/  # Integration tests here
└── ../tests/e2e/   # E2E tests here
```

## Test Structure

**Suite Organization:**
```go
// Standard Go test pattern (inferred from task configuration)
func TestFunctionName(t *testing.T) {
    // Test implementation
}
```

**Patterns:**
- Race detection enabled: `-race` flag in unit and integration tests
- Coverage reporting: `-coverprofile=coverage.out`
- Verbose E2E output: `-v` flag for detailed logging
- Test categories separated by build tags and directory structure

## Mocking

**Framework:** `mockery` (code generation tool)

**Patterns:**
```bash
task generate:mocks  # Generate mocks with mockery
```

**What to Mock:**
- Database interfaces (likely mocking `Querier` interface)
- External service calls (OAuth providers, AI services)
- HTTP clients for integration tests

**What NOT to Mock:**
- Standard library functions
- Simple data structures
- Pure functions without side effects

## Fixtures and Factories

**Test Data:**
```sql
-- Database seeding for tests
task db:seed  # Load test data from store/db/sqlc/seeds/seed.sql
```

**Location:**
- SQL seed files: `backend/store/db/sqlc/seeds/seed.sql`
- Test database management via Docker Compose

## Coverage

**Requirements:** Coverage tracking enabled with `coverage.out`

**View Coverage:**
```bash
task test:coverage  # Opens HTML coverage report
```

## Test Types

**Unit Tests:**
- Scope: Individual functions and methods in `service/` and `store/` packages
- Approach: Fast, isolated tests with mocked dependencies
- Run with: `gotestsum --format testname -- -race -coverprofile=coverage.out`

**Integration Tests:**
- Scope: HTTP handlers in `transport/http/handlers/`
- Approach: Test request/response cycles with real database
- Tagged with: `-tags=integration`
- Run with: Full request/response testing

**E2E Tests:**
- Scope: End-to-end workflows across the entire application
- Approach: Full system testing with real database and services
- Location: `../tests/e2e/`
- Run with: Verbose logging for debugging

## Common Patterns

**Async Testing:**
```go
// Standard Go pattern for testing concurrent operations
// Uses channels and timeouts for coordination
```

**Error Testing:**
```go
// Standard Go error testing pattern
if err == nil {
    t.Error("expected error, got nil")
}
```

## Test Environment

**Database:**
- PostgreSQL via Docker Compose
- Test database reset: `task db:reset` (drop, create, migrate, seed)
- Data cleanup: `task db:clear` (truncate tables, keep schema)

**Services:**
- Required services: `task mockups:services` (postgres, rembg, ollama)
- Port isolation: Different ports for test environments

**Configuration:**
- Environment variables for test configuration
- `.env` files ignored in version control
- Test-specific database URLs via `GOTHREADS_DB_DATABASE_URL`

---

*Testing analysis: 2026-02-20*