# Backend Stage 4 — Master Data: Category, Supplier & Rack APIs

## Overview

Build CRUD APIs for the three simple master data entities: Category, Supplier (with bank accounts), and Rack. These include database migrations, GORM models, repositories, services, and handlers.

> **Prerequisite**: Stage 3 must be complete (auth, users, roles, permissions, authorization middleware).

---

## 1. Database Migrations

### 1.1 Categories Table

```sql
CREATE TABLE categories (
    id          BIGSERIAL PRIMARY KEY,
    name        VARCHAR(255) NOT NULL,
    description TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

### 1.2 Suppliers Table

```sql
CREATE TABLE suppliers (
    id         BIGSERIAL PRIMARY KEY,
    name       VARCHAR(255) NOT NULL,
    address    TEXT NOT NULL,
    phone      VARCHAR(50),
    email      VARCHAR(255),
    website    VARCHAR(255),
    active     BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_suppliers_active ON suppliers(active);
```

### 1.3 Supplier Bank Accounts Table

```sql
CREATE TABLE supplier_bank_accounts (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    supplier_id    BIGINT NOT NULL REFERENCES suppliers(id) ON DELETE CASCADE,
    account_name   VARCHAR(255) NOT NULL,
    account_number VARCHAR(100) NOT NULL
);

CREATE INDEX idx_supplier_bank_accounts_supplier_id ON supplier_bank_accounts(supplier_id);
```

### 1.4 Racks Table

```sql
CREATE TABLE racks (
    id          BIGSERIAL PRIMARY KEY,
    name        VARCHAR(255) NOT NULL,
    code        VARCHAR(50) NOT NULL UNIQUE,
    location    VARCHAR(255) NOT NULL,
    capacity    INTEGER NOT NULL CHECK (capacity > 0),
    description TEXT,
    active      BOOLEAN NOT NULL DEFAULT true,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_racks_code_lower ON racks(LOWER(code));
CREATE INDEX idx_racks_active ON racks(active);
```

---

## 2. GORM Models

```go
type Category struct {
    ID          uint      `json:"id" gorm:"primaryKey"`
    Name        string    `json:"name"`
    Description string    `json:"description,omitempty"`
    CreatedAt   time.Time `json:"createdAt"`
    UpdatedAt   time.Time `json:"updatedAt"`
}

type Supplier struct {
    ID           uint                  `json:"id" gorm:"primaryKey"`
    Name         string                `json:"name"`
    Address      string                `json:"address"`
    Phone        string                `json:"phone,omitempty"`
    Email        string                `json:"email,omitempty"`
    Website      string                `json:"website,omitempty"`
    Active       bool                  `json:"active" gorm:"default:true"`
    BankAccounts []SupplierBankAccount `json:"bankAccounts" gorm:"foreignKey:SupplierID"`
    CreatedAt    time.Time             `json:"createdAt"`
    UpdatedAt    time.Time             `json:"updatedAt"`
}

type SupplierBankAccount struct {
    ID            string `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
    SupplierID    uint   `json:"supplierId" gorm:"column:supplier_id"`
    AccountName   string `json:"accountName" gorm:"column:account_name"`
    AccountNumber string `json:"accountNumber" gorm:"column:account_number"`
}

type Rack struct {
    ID          uint      `json:"id" gorm:"primaryKey"`
    Name        string    `json:"name"`
    Code        string    `json:"code" gorm:"uniqueIndex"`
    Location    string    `json:"location"`
    Capacity    int       `json:"capacity"`
    Description string    `json:"description,omitempty"`
    Active      bool      `json:"active" gorm:"default:true"`
    CreatedAt   time.Time `json:"createdAt"`
    UpdatedAt   time.Time `json:"updatedAt"`
}
```

---

## 3. Category API

### 3.1 List Categories — `GET /api/v1/categories`

**Permission**: `Master Data > Category > read`

**Query Parameters:**

| Param | Type | Default | Description |
|-------|------|---------|-------------|
| `page` | int | 1 | Page number |
| `pageSize` | int | 10 | Items per page |
| `search` | string | — | Search by name or description |
| `sortBy` | string | `id` | Sort: `id`, `name`, `description` |
| `sortDir` | string | `asc` | Sort direction |

**Response (200):** Paginated list of categories.

### 3.2 Get Category — `GET /api/v1/categories/:id`

**Permission**: `Master Data > Category > read`

**Response (200):** Single category object.

### 3.3 Create Category — `POST /api/v1/categories`

**Permission**: `Master Data > Category > create`

**Request Body:**

```json
{
  "name": "Electronics",
  "description": "Electronic devices and accessories"
}
```

**Validation:**
- `name`: required, 1-255 characters
- `description`: optional

**Response (201):** Created category object.

### 3.4 Update Category — `PUT /api/v1/categories/:id`

**Permission**: `Master Data > Category > update`

**Response (200):** Updated category object.

### 3.5 Delete Category — `DELETE /api/v1/categories/:id`

**Permission**: `Master Data > Category > delete`

**Business Logic:**
- Check if category is referenced by any products
- If referenced, return `409` with message: "Cannot delete category. It is referenced by {n} product(s)."
- If not referenced, delete and return success

**Response (200):**

```json
{
  "message": "Category deleted successfully"
}
```

---

## 4. Supplier API

### 4.1 List Suppliers — `GET /api/v1/suppliers`

**Permission**: `Master Data > Supplier > read`

**Query Parameters:**

| Param | Type | Default | Description |
|-------|------|---------|-------------|
| `page` | int | 1 | Page number |
| `pageSize` | int | 10 | Items per page |
| `search` | string | — | Search by name, address, email |
| `sortBy` | string | `id` | Sort: `id`, `name`, `active` |
| `sortDir` | string | `asc` | Sort direction |
| `active` | bool | — | Filter by active status |

**Response (200):** Paginated list of suppliers with bank accounts eager-loaded.

### 4.2 Get Supplier — `GET /api/v1/suppliers/:id`

**Permission**: `Master Data > Supplier > read`

**Response (200):** Single supplier with bank accounts.

### 4.3 Create Supplier — `POST /api/v1/suppliers`

**Permission**: `Master Data > Supplier > create`

**Request Body:**

```json
{
  "name": "PT Sumber Makmur",
  "address": "Jl. Industri No. 45, Jakarta",
  "phone": "+62-21-5550001",
  "email": "order@sumbermakmur.co.id",
  "website": "sumbermakmur.co.id",
  "bankAccounts": [
    {
      "accountName": "BCA - Main Account",
      "accountNumber": "1234567890"
    },
    {
      "accountName": "Mandiri - Operations",
      "accountNumber": "0987654321"
    }
  ]
}
```

**Validation:**
- `name`: required, 1-255 characters
- `address`: required
- `phone`: optional
- `email`: optional, valid email format if provided
- `website`: optional
- `bankAccounts`: optional array
  - If provided, each item requires both `accountName` and `accountNumber`

**Business Logic:**
1. Create supplier
2. Create bank accounts in a single transaction
3. Return supplier with bank accounts

**Response (201):** Created supplier with bank accounts.

### 4.4 Update Supplier — `PUT /api/v1/suppliers/:id`

**Permission**: `Master Data > Supplier > update`

**Request Body:** Same as create, plus `active` field.

**Business Logic:**
1. Update supplier fields
2. Sync bank accounts: delete existing, insert new ones (full replace strategy)
3. Use a database transaction for atomicity

**Response (200):** Updated supplier with bank accounts.

### 4.5 Delete Supplier — `DELETE /api/v1/suppliers/:id`

**Permission**: `Master Data > Supplier > delete`

**Business Logic:**
1. Check if supplier is referenced by products (`product_suppliers`) or purchase orders
2. If referenced by purchase orders → return `409`: "Cannot delete supplier. It is referenced by {n} purchase order(s)."
3. If referenced only by products → delete supplier, also clean up `product_suppliers` junction entries
4. CASCADE deletes bank accounts automatically

**Response (200):**

```json
{
  "message": "Supplier deleted successfully"
}
```

---

## 5. Rack API

### 5.1 List Racks — `GET /api/v1/racks`

**Permission**: `Master Data > Category > read` (racks fall under general master data access; alternatively create a separate permission in the future)

> **Note**: Racks don't have their own permission entry in the current permission seed data. Use `Master Data > Product > read` as the controlling permission for rack access (since racks are a product attribute). Or add a new permission row for racks — either approach is fine. Recommendation: use the same Product permission since racks are closely tied to products.

**Query Parameters:**

| Param | Type | Default | Description |
|-------|------|---------|-------------|
| `page` | int | 1 | Page number |
| `pageSize` | int | 10 | Items per page |
| `search` | string | — | Search by name, code, location |
| `sortBy` | string | `id` | Sort: `id`, `name`, `code`, `location`, `active` |
| `sortDir` | string | `asc` | Sort direction |
| `active` | bool | — | Filter by active status |

**Response (200):** Paginated list of racks.

### 5.2 Get Rack — `GET /api/v1/racks/:id`

**Response (200):** Single rack object.

### 5.3 Create Rack — `POST /api/v1/racks`

**Request Body:**

```json
{
  "name": "Main Display",
  "code": "R-001",
  "location": "Store Front",
  "capacity": 100,
  "description": "Primary display shelf near entrance"
}
```

**Validation:**
- `name`: required, 1-255 characters
- `code`: required, unique (case-insensitive), 1-50 characters
- `location`: required, 1-255 characters
- `capacity`: required, must be > 0
- `description`: optional

**Response (201):** Created rack object.

### 5.4 Update Rack — `PUT /api/v1/racks/:id`

**Request Body:** Same as create, plus `active` field.

**Validation:** Code uniqueness check must exclude current rack.

**Response (200):** Updated rack object.

### 5.5 Delete Rack — `DELETE /api/v1/racks/:id`

**Business Logic:**
- Check if rack is referenced by product variants (`variant_racks`)
- If referenced, clean up `variant_racks` junction entries (don't block deletion)
- Delete rack

**Response (200):**

```json
{
  "message": "Rack deleted successfully"
}
```

---

## 6. Seed Data

### 6.1 Categories

| Name | Description |
|------|-------------|
| Clothing | Apparel and garments |
| Food & Beverages | Food items and drinks |
| Stationery | Office and school supplies |
| Household | Home and kitchen essentials |

### 6.2 Suppliers

| Name | Address | Phone | Email | Website | Bank Accounts | Active |
|------|---------|-------|-------|---------|---------------|--------|
| PT Sumber Makmur | Jl. Industri No. 45, Jakarta | +62-21-5550001 | order@sumbermakmur.co.id | sumbermakmur.co.id | BCA - 1234567890, Mandiri - 0987654321 | true |
| CV Jaya Abadi | Jl. Perdagangan No. 12, Surabaya | +62-31-5550002 | sales@jayaabadi.com | — | BCA - 1122334455 | true |
| UD Berkah Sentosa | Jl. Pasar Baru No. 8, Bandung | — | — | — | — | true |
| PT Global Supplies | Jl. Raya Serpong No. 100, Tangerang | +62-21-5550004 | info@globalsupplies.co.id | globalsupplies.co.id | BNI - 5566778899, BRI - 9988776655 | false |

### 6.3 Racks

| Name | Code | Location | Capacity | Description | Active |
|------|------|----------|----------|-------------|--------|
| Main Display | R-001 | Store Front | 100 | Primary display shelf near entrance | true |
| Electronics Shelf | R-002 | Store Front | 50 | Dedicated electronics display | true |
| Cold Storage | R-003 | Warehouse Zone A | 200 | Refrigerated storage area | true |
| Bulk Storage | R-004 | Warehouse Zone B | 500 | Large item storage | true |
| Clearance Rack | R-005 | Store Back | 30 | Discounted items | false |

---

## 7. Route Registration

```go
r.Route("/api/v1", func(r chi.Router) {
    r.Use(middleware.RequireAuth)

    // Categories
    r.Route("/categories", func(r chi.Router) {
        r.With(RequirePermission("Master Data", "Category", "read")).Get("/", handlers.ListCategories)
        r.With(RequirePermission("Master Data", "Category", "read")).Get("/{id}", handlers.GetCategory)
        r.With(RequirePermission("Master Data", "Category", "create")).Post("/", handlers.CreateCategory)
        r.With(RequirePermission("Master Data", "Category", "update")).Put("/{id}", handlers.UpdateCategory)
        r.With(RequirePermission("Master Data", "Category", "delete")).Delete("/{id}", handlers.DeleteCategory)
    })

    // Suppliers
    r.Route("/suppliers", func(r chi.Router) {
        r.With(RequirePermission("Master Data", "Supplier", "read")).Get("/", handlers.ListSuppliers)
        r.With(RequirePermission("Master Data", "Supplier", "read")).Get("/{id}", handlers.GetSupplier)
        r.With(RequirePermission("Master Data", "Supplier", "create")).Post("/", handlers.CreateSupplier)
        r.With(RequirePermission("Master Data", "Supplier", "update")).Put("/{id}", handlers.UpdateSupplier)
        r.With(RequirePermission("Master Data", "Supplier", "delete")).Delete("/{id}", handlers.DeleteSupplier)
    })

    // Racks
    r.Route("/racks", func(r chi.Router) {
        r.With(RequirePermission("Master Data", "Product", "read")).Get("/", handlers.ListRacks)
        r.With(RequirePermission("Master Data", "Product", "read")).Get("/{id}", handlers.GetRack)
        r.With(RequirePermission("Master Data", "Product", "create")).Post("/", handlers.CreateRack)
        r.With(RequirePermission("Master Data", "Product", "update")).Put("/{id}", handlers.UpdateRack)
        r.With(RequirePermission("Master Data", "Product", "delete")).Delete("/{id}", handlers.DeleteRack)
    })
})
```

---

## 8. TDD Workflow

**Follow strict TDD. Write tests BEFORE implementations for each entity.**

### 8.1 Category — TDD Order

**`repositories/category_repository_test.go`**:
- `TestCreateCategory_Valid_Succeeds`
- `TestListCategories_Pagination_Works`
- `TestListCategories_Search_FiltersByNameAndDescription`
- `TestListCategories_Sort_OrdersCorrectly`
- `TestGetCategory_Exists_ReturnsCategory`
- `TestGetCategory_NotFound_ReturnsNil`
- `TestUpdateCategory_Valid_UpdatesFields`
- `TestDeleteCategory_NoReferences_Succeeds`

**`services/category_service_test.go`**:
- `TestCreateCategory_Valid_Succeeds`
- `TestDeleteCategory_ReferencedByProducts_ReturnsConflict`
- `TestDeleteCategory_Unreferenced_Succeeds`

**`handlers/category_handler_test.go`**:
- `TestListCategories_Returns200WithPagination`
- `TestListCategories_WithSearch_FiltersResults`
- `TestListCategories_WithSort_OrdersResults`
- `TestGetCategory_Exists_Returns200`
- `TestGetCategory_NotFound_Returns404`
- `TestCreateCategory_ValidBody_Returns201`
- `TestCreateCategory_MissingName_Returns400`
- `TestCreateCategory_NoAuth_Returns401`
- `TestCreateCategory_NoPermission_Returns403`
- `TestUpdateCategory_ValidBody_Returns200`
- `TestUpdateCategory_NotFound_Returns404`
- `TestDeleteCategory_Unreferenced_Returns200`
- `TestDeleteCategory_ReferencedByProduct_Returns409`

### 8.2 Supplier — TDD Order

**`repositories/supplier_repository_test.go`**:
- `TestCreateSupplier_WithBankAccounts_CreatesAll`
- `TestGetSupplier_EagerLoadsBankAccounts`
- `TestListSuppliers_FilterByActive_Works`
- `TestListSuppliers_SearchByNameAddressEmail_Works`
- `TestUpdateSupplier_SyncBankAccounts_ReplacesAll`
- `TestDeleteSupplier_CascadesBankAccounts`

**`services/supplier_service_test.go`**:
- `TestCreateSupplier_Valid_Succeeds`
- `TestCreateSupplier_BankAccountMissingFields_ReturnsValidation`
- `TestUpdateSupplier_SyncsBankAccountsAtomically`
- `TestDeleteSupplier_ReferencedByPO_ReturnsConflict`
- `TestDeleteSupplier_ReferencedByProductsOnly_CleansUpAndDeletes`

**`handlers/supplier_handler_test.go`**:
- `TestListSuppliers_Returns200WithBankAccounts`
- `TestListSuppliers_FilterActive_ReturnsActiveOnly`
- `TestCreateSupplier_WithBankAccounts_Returns201`
- `TestCreateSupplier_MissingName_Returns400`
- `TestCreateSupplier_InvalidEmail_Returns400`
- `TestCreateSupplier_BankAccountIncomplete_Returns400`
- `TestUpdateSupplier_ReplacesBankAccounts_Returns200`
- `TestDeleteSupplier_NoReferences_Returns200`
- `TestDeleteSupplier_ReferencedByPO_Returns409`

### 8.3 Rack — TDD Order

**`repositories/rack_repository_test.go`**:
- `TestCreateRack_Valid_Succeeds`
- `TestCreateRack_DuplicateCode_ReturnsError`
- `TestCreateRack_DuplicateCodeDifferentCase_ReturnsError`
- `TestListRacks_SearchByNameCodeLocation_Works`
- `TestListRacks_FilterByActive_Works`

**`services/rack_service_test.go`**:
- `TestCreateRack_Valid_Succeeds`
- `TestCreateRack_DuplicateCode_ReturnsConflict`
- `TestUpdateRack_CodeUniqueExcludesSelf`
- `TestDeleteRack_CleansUpVariantRacks`

**`handlers/rack_handler_test.go`**:
- `TestListRacks_Returns200`
- `TestCreateRack_ValidBody_Returns201`
- `TestCreateRack_DuplicateCode_Returns409`
- `TestCreateRack_MissingCode_Returns400`
- `TestCreateRack_ZeroCapacity_Returns400`
- `TestUpdateRack_DuplicateCodeOtherRack_Returns409`
- `TestUpdateRack_SameCodeSelf_Returns200`
- `TestDeleteRack_Returns200_CleansVariantRacks`

### 8.4 Seed Tests

**`seeds/seed_test.go`**:
- `TestSeedCategories_CreatesExpectedData`
- `TestSeedSuppliers_CreatesWithBankAccounts`
- `TestSeedRacks_CreatesExpectedData`
- `TestSeedIdempotent_RunTwice_NoErrors` — seeds can be run multiple times without duplicating data

---

## 9. Deliverables

After completing this stage:

1. Category CRUD works with pagination, search, and sorting
2. Category deletion is blocked when referenced by products
3. Supplier CRUD works with inline bank account management
4. Supplier bank accounts are created/updated/deleted atomically with the supplier
5. Supplier deletion cleans up product-supplier associations
6. Rack CRUD works with unique code validation (case-insensitive)
7. Rack deletion cleans up variant-rack associations
8. All endpoints are protected by permission middleware
9. Seed data is populated for all three entities
10. **All tests pass** (`go test ./...`)
11. **Test coverage** ≥ 80% across handlers and services
