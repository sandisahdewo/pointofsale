# Point of Sale - Backend API

## Tech Stack
- **Language**: Go 1.24
- **Router**: Chi v5
- **Database**: PostgreSQL 17 (via GORM for queries, goose for migrations)
- **Cache**: Redis 7 (JWT refresh tokens, token blacklisting, permission caching)
- **Mail**: Mailpit (dev SMTP on :1025, Web UI on :8025)
- **Auth**: Argon2id password hashing, JWT (access 15min + refresh 7d)

## Project Structure

```
backend/
├── cmd/server/main.go           # Entry point
├── config/config.go             # Env config loader
├── handlers/                    # HTTP handlers (one file per domain)
│   └── *_test.go                # Handler/integration tests
├── middleware/                   # Auth, CORS, permissions, logging
│   └── *_test.go
├── models/                      # GORM model structs
├── repositories/                # Database queries (GORM-based)
│   └── *_test.go                # Repository tests
├── services/                    # Business logic layer
│   └── *_test.go                # Service unit tests
├── migrations/                  # Versioned .sql files (goose up/down)
├── seeds/                       # Database seed data
├── routes/routes.go             # Route definitions
├── utils/                       # Helpers (password, JWT, validation)
│   └── *_test.go                # Utility unit tests
├── testutil/                    # Shared test helpers and fixtures
│   ├── db.go                    # Test database setup/teardown
│   ├── fixtures.go              # Factory functions for test data
│   ├── auth.go                  # JWT helper for authenticated requests
│   └── assert.go                # Custom assertion helpers
├── Dockerfile
├── .env.example
├── .air.toml
└── go.mod
```

## TDD Methodology — MANDATORY

**STRICT RULE: Tests MUST be written BEFORE implementation code. No exceptions.**

You are FORBIDDEN from writing implementation code first. The workflow is always: test first, then code. If you catch yourself about to write a handler, service, repository, or utility function — STOP. Write the test for it first.

### Red-Green-Refactor Cycle (follow this EXACTLY)

**Step 1 — RED (write the test FIRST):**
- Write a `_test.go` file with test cases that define the expected behavior
- The test MUST reference functions/types/endpoints that DO NOT EXIST yet
- Run the test with `go test` — it MUST FAIL (compile error or assertion failure)
- If the test passes, you wrote it wrong — the behavior already exists or the test is trivial

**Step 2 — GREEN (write the MINIMUM implementation):**
- ONLY NOW write the production code (handler, service, repository, etc.)
- Write just enough code to make the failing test pass — nothing more
- Run the test again — it MUST PASS now
- Do NOT add extra features, edge case handling, or optimizations not covered by a test

**Step 3 — REFACTOR (clean up while green):**
- Clean up code (rename, extract, simplify) while keeping all tests passing
- Run tests after each refactor to confirm nothing broke

### Workflow Per Feature (step by step)

For each new feature (e.g., "Create Category endpoint"), do this:

1. **Write the test file** (e.g., `handlers/category_test.go`) with all test cases
2. **Run the test** — confirm it fails (RED)
3. **Write the model** if needed (only what the test requires)
4. **Write the repository/service/handler** — only enough to pass the test
5. **Run the test** — confirm it passes (GREEN)
6. **Refactor** if needed, run tests again
7. **Repeat** for the next test case or feature

### NEVER do these:
- Write a complete handler/service file and then write tests after
- Write tests and implementation in the same step without running the test in between
- Skip the RED step (you MUST see the test fail before writing implementation)
- Write implementation code "to save time" and add tests later

### Test Layers (write in this order per feature)

| Layer | Directory | Tests What | Depends On |
|-------|-----------|------------|------------|
| **1. Utils** | `utils/*_test.go` | Pure functions (hashing, JWT, validation) | Nothing |
| **2. Repository** | `repositories/*_test.go` | Database queries against real test DB | Test database |
| **3. Service** | `services/*_test.go` | Business logic with mocked repositories | Interfaces/mocks |
| **4. Handler** | `handlers/*_test.go` | HTTP request/response integration tests | Running test server + test DB |

### Test Naming Convention
```go
func TestFunctionName_Scenario_ExpectedResult(t *testing.T) {
    // Examples:
    // TestHashPassword_ValidPassword_ReturnsHash
    // TestLogin_InvalidEmail_Returns401
    // TestCreateCategory_DuplicateName_Returns409
    // TestReceivePO_SentStatus_UpdatesStock
}
```

### Test Database Setup
- Use a **separate PostgreSQL database** for tests (e.g., `pointofsale_test`)
- Run goose migrations before tests
- Each test function gets a **clean database** (truncate tables in `TestMain` or per-test setup)
- Use transactions that rollback for isolation where possible
- NEVER run tests against the development database

```go
// testutil/db.go pattern
func SetupTestDB(t *testing.T) *gorm.DB {
    // Connect to pointofsale_test database
    // Run migrations
    // Return DB handle
}

func CleanupTestDB(t *testing.T, db *gorm.DB) {
    // Truncate all tables in reverse dependency order
}
```

### Test Fixtures / Factory Functions
```go
// testutil/fixtures.go pattern
func CreateTestUser(t *testing.T, db *gorm.DB, overrides ...func(*models.User)) *models.User
func CreateTestRole(t *testing.T, db *gorm.DB, overrides ...func(*models.Role)) *models.Role
func CreateTestCategory(t *testing.T, db *gorm.DB, overrides ...func(*models.Category)) *models.Category
func CreateTestProduct(t *testing.T, db *gorm.DB, overrides ...func(*models.Product)) *models.Product
// ... etc for each entity
```

### Handler Test Pattern (HTTP Integration Tests)
```go
func TestListCategories_Success(t *testing.T) {
    // Setup: test DB, seed data, create authenticated request
    db := testutil.SetupTestDB(t)
    defer testutil.CleanupTestDB(t, db)

    // Create test data
    testutil.CreateTestCategory(t, db, func(c *models.Category) {
        c.Name = "Electronics"
    })

    // Create authenticated request
    req := testutil.AuthenticatedRequest(t, "GET", "/api/v1/categories", nil, testutil.SuperAdminToken)

    // Execute
    rr := httptest.NewRecorder()
    router.ServeHTTP(rr, req)

    // Assert
    assert.Equal(t, http.StatusOK, rr.Code)
    var response map[string]interface{}
    json.Unmarshal(rr.Body.Bytes(), &response)
    // ... assert response shape and data
}
```

### What to Test Per Endpoint

For **every** API endpoint, write tests for:
1. **Happy path** — valid request returns expected response
2. **Validation errors** — missing/invalid fields return 400
3. **Auth required** — unauthenticated request returns 401
4. **Permission denied** — user without permission returns 403
5. **Not found** — invalid ID returns 404
6. **Business rules** — domain-specific constraints (e.g., can't delete super admin, can't edit non-draft PO)
7. **Edge cases** — empty lists, duplicate names, concurrent operations

### Running Tests
```bash
go test ./...                  # all tests
go test ./handlers/...         # handler tests only
go test ./services/...         # service tests only
go test -run TestLogin ./...   # specific test
go test -v -count=1 ./...      # verbose, no cache
go test -race ./...            # race condition detection
go test -cover ./...           # coverage report
```

### Mocking Pattern (for service tests)
```go
// Define interfaces in services package
type UserRepository interface {
    FindByEmail(email string) (*models.User, error)
    Create(user *models.User) error
    // ...
}

// Mock in test file
type mockUserRepo struct {
    findByEmailFn func(string) (*models.User, error)
    createFn      func(*models.User) error
}

func (m *mockUserRepo) FindByEmail(email string) (*models.User, error) {
    return m.findByEmailFn(email)
}
```

## Running the Backend
```bash
docker compose up                              # all services (from project root)
docker compose up backend                      # backend only
docker compose exec backend go test ./...      # run tests inside container
go test ./...                                  # run tests locally (needs test DB)
```

## Conventions
- All API routes under `/api/v1/` prefix
- JSON request/response with consistent error format: `{"error": "message", "code": "CODE"}`
- Success format: `{"data": {...}, "message": "optional"}`
- Paginated lists: `{"data": [...], "meta": {"page", "pageSize", "totalItems", "totalPages"}}`
- Use database transactions for multi-table writes
- Never expose `password_hash` in JSON responses (`json:"-"`)
- Super admin bypasses all permission checks
- Validate `sortBy` fields against allowlists to prevent SQL injection
