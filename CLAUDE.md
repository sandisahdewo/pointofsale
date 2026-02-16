# Point of Sale - Admin Panel

## Tech Stack
- **Framework**: Next.js 16 (React 19) — CSR only (`'use client'` on all pages)
- **Styling**: Tailwind CSS v4 (no shadcn-ui, no UI libraries)
- **State**: Zustand v5 (stores in `frontend/src/stores/`)
- **Language**: TypeScript
- **Package manager**: npm (lockfile: `frontend/package-lock.json`)

## Project Structure

```
frontend/src/
├── app/                          # Next.js pages (all 'use client')
│   ├── layout.tsx                # Root layout (bare html/body, no providers)
│   ├── page.tsx                  # Redirects to /login
│   ├── globals.css               # Tailwind imports
│   ├── login/page.tsx
│   ├── register/page.tsx
│   ├── reset-password/page.tsx
│   ├── dashboard/page.tsx
│   ├── master/
│   │   ├── category/page.tsx     # Simple CRUD with inline modal
│   │   ├── product/              # Complex CRUD with add/edit sub-routes
│   │   │   ├── page.tsx          # List page
│   │   │   ├── add/page.tsx      # Uses ProductForm component
│   │   │   └── edit/[id]/page.tsx
│   │   ├── supplier/page.tsx     # CRUD with inline modal
│   │   └── rack/page.tsx         # CRUD with inline modal
│   ├── transaction/
│   │   └── purchase/
│   │       ├── page.tsx          # List page
│   │       ├── add/page.tsx      # Uses PurchaseOrderForm
│   │       └── [id]/
│   │           ├── page.tsx      # Detail/view page
│   │           ├── edit/page.tsx
│   │           └── receive/page.tsx
│   └── settings/
│       ├── users/page.tsx        # CRUD with modal
│       └── roles/
│           ├── page.tsx          # CRUD with modal
│           └── [id]/permissions/page.tsx
├── components/
│   ├── layout/                   # AdminLayout, Header, Sidebar, Footer
│   ├── ui/                       # Reusable UI components (see list below)
│   ├── product/                  # ProductForm, UnitsTab, VariantsTab, VariantPricing
│   ├── purchase/                 # PurchaseOrderForm
│   ├── role/                     # RoleFormModal
│   └── user/                     # UserFormModal
├── data/                         # Mock/initial data (no backend yet)
│   ├── categories.ts, products.ts, suppliers.ts, racks.ts
│   ├── purchaseOrders.ts
│   ├── users.ts, roles.ts, permissions.ts, rolePermissions.ts
└── stores/                       # Zustand stores (one per entity)
    ├── useCategoryStore.ts, useProductStore.ts, useSupplierStore.ts
    ├── useRackStore.ts, usePurchaseOrderStore.ts
    ├── useUserStore.ts, useRoleStore.ts
    ├── useSidebarStore.ts, useToastStore.ts
```

## UI Components (`components/ui/`)

| Component | Props | Notes |
|-----------|-------|-------|
| Button | variant: primary/secondary/danger/outline, size: sm/md/lg | |
| Input | label?, error? + native input props | |
| Textarea | label?, error? + native textarea props | |
| Select | label?, error?, options[] | |
| MultiSelect | label?, options[], value[], onChange | Tag-style multi-select |
| Checkbox | label?, checked, onChange | |
| Toggle | label?, enabled, onChange | |
| TagInput | label?, tags[], onAdd, onRemove | |
| Table | columns[], data[], pagination, sorting | Generic `<T extends {id: number}>` |
| Modal | isOpen, onClose, title, children | |
| ConfirmModal | isOpen, onClose, onConfirm, title, message | |
| Card | className?, children | Simple wrapper |
| Alert | type: success/error/warning/info, message | |
| Toast | Global toast via useToastStore | |
| Dropdown | trigger, items[], align? | |
| Avatar | name, size? | Initials-based |
| Badge | variant, children | |
| StatusBadge | status string | |
| DatePicker | label?, value, onChange | |
| ImageUpload | images[], onChange, maxImages? | |
| SidebarMenu | items[] (tree structure) | |
| Tabs | tabs[], activeTab, onChange | |

## Conventions & Patterns

### Pages
- Every page starts with `'use client'`
- Admin pages wrap content in `<AdminLayout>` (provides header, sidebar, footer, toast)
- Auth pages (login, register, reset-password) do NOT use AdminLayout

### CRUD Pattern (Simple — e.g., Category, Rack)
All-in-one page file with:
1. Zustand store hook for data + CRUD actions
2. `useToastStore` for notifications
3. Search/filter with `useMemo`
4. Sort with `useMemo`
5. Client-side pagination (DEFAULT_PAGE_SIZE = 10)
6. Inline add/edit Modal with form validation
7. Delete confirmation Modal
8. Table with columns definition including action buttons

### CRUD Pattern (Complex — e.g., Product)
- List page at `master/{entity}/page.tsx`
- Separate add/edit pages with shared Form component in `components/{entity}/`
- Form component handles all tabs and complex state

### Zustand Store Pattern
```typescript
'use client';
import { create } from 'zustand';
import { initialItems } from '@/data/{entity}';

export interface Entity { id: number; /* fields */ }

interface EntityState {
  items: Entity[];
  addItem: (item: Omit<Entity, 'id'>) => void;
  updateItem: (id: number, item: Partial<Entity>) => void;
  deleteItem: (id: number) => void;
}

export const useEntityStore = create<EntityState>((set) => ({
  items: initialItems,
  addItem: (item) => set((state) => {
    const maxId = state.items.reduce((max, i) => Math.max(max, i.id), 0);
    return { items: [...state.items, { ...item, id: maxId + 1 }] };
  }),
  // ... update and delete follow same pattern
}));
```

### Mock Data Pattern
```typescript
import { Entity } from '@/stores/useEntityStore';
export const initialItems: Entity[] = [ /* seed data */ ];
```

### Sidebar Navigation
Defined in `components/layout/Sidebar.tsx` — update `menuItems` array when adding new pages.

## Running the App
```bash
cd frontend && npm run dev    # dev server
cd frontend && npm run build  # production build
cd frontend && npm run lint   # eslint
```

## Important Notes
- No shadcn-ui — all UI components are custom-built
- Tailwind uses default color palette (custom palette planned for future)
- All imports use `@/` path alias (maps to `frontend/src/`)

---

# Point of Sale - Backend API

## Tech Stack
- **Language**: Go 1.24
- **Router**: Chi v5
- **Database**: PostgreSQL 17 (via GORM for queries, goose for migrations)
- **Cache**: Redis 7 (JWT refresh tokens, token blacklisting, permission caching)
- **Mail**: Mailpit (dev SMTP on :1025, Web UI on :8025)
- **Auth**: Argon2id password hashing, JWT (access 15min + refresh 7d)
- **Infra**: Docker Compose (backend, frontend, postgres, redis, mailpit)

## Backend Project Structure

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

## TDD Methodology

**Every feature MUST follow strict Test-Driven Development:**

### Red-Green-Refactor Cycle
1. **RED**: Write a failing test that defines the expected behavior
2. **GREEN**: Write the minimum code to make the test pass
3. **REFACTOR**: Clean up the code while keeping tests green

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
cd backend && go test ./...                  # all tests
cd backend && go test ./handlers/...         # handler tests only
cd backend && go test ./services/...         # service tests only
cd backend && go test -run TestLogin ./...   # specific test
cd backend && go test -v -count=1 ./...      # verbose, no cache
cd backend && go test -race ./...            # race condition detection
cd backend && go test -cover ./...           # coverage report
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
docker compose up                  # all services (from project root)
docker compose up backend          # backend only
docker compose exec backend go test ./...  # run tests inside container
cd backend && go test ./...        # run tests locally (needs test DB)
```

## Backend Conventions
- All API routes under `/api/v1/` prefix
- JSON request/response with consistent error format: `{"error": "message", "code": "CODE"}`
- Success format: `{"data": {...}, "message": "optional"}`
- Paginated lists: `{"data": [...], "meta": {"page", "pageSize", "totalItems", "totalPages"}}`
- Use database transactions for multi-table writes
- Never expose `password_hash` in JSON responses (`json:"-"`)
- Super admin bypasses all permission checks
- Validate `sortBy` fields against allowlists to prevent SQL injection
