# Coding Conventions

**Analysis Date:** 2026-02-20

## Naming Patterns

**Files:**
- Go files: snake_case (`main.go`, `auth.go`, `config.go`)
- SQL generated files: snake_case with `.sql.go` suffix (`items.sql.go`, `users.sql.go`)
- Configuration files: snake_case with extensions (`config.yaml`, `docker-compose.yml`)
- Test files: `*_test.go` pattern (though no test files found)

**Functions:**
- Public functions: PascalCase (`GetUser`, `CreateItem`, `HandleOAuthLogin`)
- Private functions: camelCase (`getEnv`, `getUserFromRequest`, `parseClothingResponse`)
- HTTP handlers: PascalCase with `handle` prefix (`handleHealth`, `handleGetConfig`)

**Variables:**
- Global variables: camelCase (`ollamaURL`, `visionModel`, `textModel`)
- Struct fields: PascalCase for public (`ID`, `Email`, `Name`), camelCase for private
- Constants: UPPER_SNAKE_CASE (`defaultOllamaURL`, `defaultVisionModel`)
- Package-level vars: camelCase with descriptive names

**Types:**
- Structs: PascalCase (`User`, `AnalyzeRequest`, `OAuthConfig`)
- Interfaces: PascalCase (`Querier`)
- Type aliases: PascalCase following Go conventions

## Code Style

**Formatting:**
- Tool used: `gofmt` (standard Go formatter)
- Key settings: Standard Go formatting (tabs for indentation, automatic spacing)

**Linting:**
- Tool used: `golangci-lint`
- Configuration: Standard Go linting rules
- Task: `task lint:go` runs linter on backend code

## Import Organization

**Order:**
1. Standard library imports (`context`, `encoding/json`, `fmt`, `log`)
2. Third-party imports (`github.com/golang-jwt/jwt/v5`, `github.com/jackc/pgx/v5`)
3. Local imports (none observed in current structure)

**Path Aliases:**
- No custom path aliases used
- Standard Go import paths with full module names
- Module: `github.com/haseebmajid/gothreads/mockups/api`

## Error Handling

**Patterns:**
- Standard Go error handling with explicit `if err != nil` checks
- Error wrapping using `fmt.Errorf("message: %w", err)` pattern
- HTTP errors using custom `writeError` helper function
- Logging errors before returning: `log.Printf("Error: %v", err)`
- Graceful fallbacks in AI operations: try primary model, fallback to secondary

## Logging

**Framework:** Standard library `log` package

**Patterns:**
- Info messages: `log.Printf("Message: %v", value)`
- Error messages: `log.Printf("Error: %v", err)`
- Status updates: `log.Printf("✅ Success message")` with emoji prefixes
- Warning messages: `log.Printf("⚠️  Warning message")`
- Debug traces: `log.Printf("🤖 AI operation: %s", details)`

## Comments

**When to Comment:**
- Package declarations with purpose
- Complex algorithms (AI parsing, OAuth flow)
- Public APIs and exported functions
- Configuration structures with field explanations
- TODO items for known limitations: `// TODO: auth` (found in `main.go:3003`)

**JSDoc/TSDoc:**
- Not applicable (Go codebase)
- Standard Go doc comments above exported functions and types

## Function Design

**Size:** Functions range from 10-200 lines, with main.go containing some longer handlers

**Parameters:** 
- HTTP handlers: `(w http.ResponseWriter, r *http.Request)`
- Database operations: `(ctx context.Context, params StructType)`
- Configuration: Parameters passed as structs rather than many individual params

**Return Values:** 
- Error-last pattern: `(result Type, error)`
- Multiple return values common: `(string, error)` or `(response, model, error)`
- Void functions return error only: `error`

## Module Design

**Exports:** 
- Types and functions follow Go visibility rules (PascalCase = public, camelCase = private)
- Main package exports HTTP handlers and core business logic
- Database package exports generated SQLC interfaces and models

**Barrel Files:** 
- Not applicable to Go
- Single `main.go` serves as main entry point
- Separate files for logical grouping: `auth.go`, `config.go`, `db.go`

---

*Convention analysis: 2026-02-20*