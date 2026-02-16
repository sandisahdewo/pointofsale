# Backend Stage 5 — Master Data: Product API (Complex)

## Overview

Build the full Product CRUD API — the most complex entity in the system. Products have nested relationships: units, variants (with attributes, images, pricing tiers, rack assignments), product images, and supplier associations. All nested data is managed within the product API (no separate endpoints for units/variants).

> **Prerequisite**: Stage 4 must be complete (categories, suppliers, racks exist in the database).

---

## 1. Database Migrations

### 1.1 Products Table

```sql
CREATE TABLE products (
    id            BIGSERIAL PRIMARY KEY,
    name          VARCHAR(255) NOT NULL,
    description   TEXT,
    category_id   BIGINT NOT NULL REFERENCES categories(id),
    price_setting VARCHAR(20) NOT NULL DEFAULT 'fixed',  -- fixed, markup
    markup_type   VARCHAR(20),                            -- percentage, fixed_amount (NULL when price_setting=fixed)
    has_variants  BOOLEAN NOT NULL DEFAULT false,
    status        VARCHAR(20) NOT NULL DEFAULT 'active',  -- active, inactive
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_products_category_id ON products(category_id);
CREATE INDEX idx_products_status ON products(status);
```

### 1.2 Product Images Table

```sql
CREATE TABLE product_images (
    id         BIGSERIAL PRIMARY KEY,
    product_id BIGINT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    image_url  TEXT NOT NULL,
    sort_order INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX idx_product_images_product_id ON product_images(product_id);
```

### 1.3 Product-Suppliers Junction Table

```sql
CREATE TABLE product_suppliers (
    product_id  BIGINT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    supplier_id BIGINT NOT NULL REFERENCES suppliers(id) ON DELETE CASCADE,
    PRIMARY KEY (product_id, supplier_id)
);

CREATE INDEX idx_product_suppliers_supplier_id ON product_suppliers(supplier_id);
```

### 1.4 Product Units Table

```sql
CREATE TABLE product_units (
    id                BIGSERIAL PRIMARY KEY,
    product_id        BIGINT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    name              VARCHAR(100) NOT NULL,
    conversion_factor DECIMAL(15,4) NOT NULL DEFAULT 1,  -- how many of the referenced unit
    converts_to_id    BIGINT REFERENCES product_units(id) ON DELETE SET NULL,  -- NULL for base unit
    to_base_unit      DECIMAL(15,4) NOT NULL DEFAULT 1,  -- total conversion to base unit
    is_base           BOOLEAN NOT NULL DEFAULT false
);

CREATE INDEX idx_product_units_product_id ON product_units(product_id);
CREATE UNIQUE INDEX idx_product_units_name_per_product ON product_units(product_id, LOWER(name));
```

**Constraints:**
- Each product must have exactly one unit where `is_base = true`
- The base unit always has `conversion_factor = 1`, `converts_to_id = NULL`, `to_base_unit = 1`
- Unit name must be unique per product (case-insensitive)

### 1.5 Product Variants Table

```sql
CREATE TABLE product_variants (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id    BIGINT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    sku           VARCHAR(100),
    barcode       VARCHAR(100),
    current_stock INTEGER NOT NULL DEFAULT 0,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_product_variants_product_id ON product_variants(product_id);
CREATE INDEX idx_product_variants_sku ON product_variants(sku) WHERE sku IS NOT NULL;
CREATE INDEX idx_product_variants_barcode ON product_variants(barcode) WHERE barcode IS NOT NULL;
```

### 1.6 Variant Attributes Table

```sql
CREATE TABLE variant_attributes (
    id              BIGSERIAL PRIMARY KEY,
    variant_id      UUID NOT NULL REFERENCES product_variants(id) ON DELETE CASCADE,
    attribute_name  VARCHAR(100) NOT NULL,
    attribute_value VARCHAR(255) NOT NULL
);

CREATE INDEX idx_variant_attributes_variant_id ON variant_attributes(variant_id);
```

### 1.7 Variant Images Table

```sql
CREATE TABLE variant_images (
    id         BIGSERIAL PRIMARY KEY,
    variant_id UUID NOT NULL REFERENCES product_variants(id) ON DELETE CASCADE,
    image_url  TEXT NOT NULL,
    sort_order INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX idx_variant_images_variant_id ON variant_images(variant_id);
```

### 1.8 Variant Pricing Tiers Table

```sql
CREATE TABLE variant_pricing_tiers (
    id         BIGSERIAL PRIMARY KEY,
    variant_id UUID NOT NULL REFERENCES product_variants(id) ON DELETE CASCADE,
    min_qty    INTEGER NOT NULL CHECK (min_qty > 0),
    value      DECIMAL(15,2) NOT NULL CHECK (value >= 0)
);

CREATE INDEX idx_variant_pricing_tiers_variant_id ON variant_pricing_tiers(variant_id);
```

### 1.9 Variant-Racks Junction Table

```sql
CREATE TABLE variant_racks (
    variant_id UUID NOT NULL REFERENCES product_variants(id) ON DELETE CASCADE,
    rack_id    BIGINT NOT NULL REFERENCES racks(id) ON DELETE CASCADE,
    PRIMARY KEY (variant_id, rack_id)
);

CREATE INDEX idx_variant_racks_rack_id ON variant_racks(rack_id);
```

---

## 2. GORM Models

```go
type Product struct {
    ID           uint              `json:"id" gorm:"primaryKey"`
    Name         string            `json:"name"`
    Description  string            `json:"description,omitempty"`
    CategoryID   uint              `json:"categoryId" gorm:"column:category_id"`
    Category     *Category         `json:"category,omitempty" gorm:"foreignKey:CategoryID"`
    PriceSetting string            `json:"priceSetting" gorm:"column:price_setting;default:fixed"`
    MarkupType   *string           `json:"markupType,omitempty" gorm:"column:markup_type"`
    HasVariants  bool              `json:"hasVariants" gorm:"column:has_variants;default:false"`
    Status       string            `json:"status" gorm:"default:active"`
    Images       []ProductImage    `json:"images" gorm:"foreignKey:ProductID"`
    Suppliers    []Supplier        `json:"suppliers" gorm:"many2many:product_suppliers;"`
    Units        []ProductUnit     `json:"units" gorm:"foreignKey:ProductID"`
    Variants     []ProductVariant  `json:"variants" gorm:"foreignKey:ProductID"`
    CreatedAt    time.Time         `json:"createdAt"`
    UpdatedAt    time.Time         `json:"updatedAt"`
}

type ProductImage struct {
    ID        uint   `json:"id" gorm:"primaryKey"`
    ProductID uint   `json:"productId" gorm:"column:product_id"`
    ImageURL  string `json:"imageUrl" gorm:"column:image_url"`
    SortOrder int    `json:"sortOrder" gorm:"column:sort_order;default:0"`
}

type ProductUnit struct {
    ID               uint     `json:"id" gorm:"primaryKey"`
    ProductID        uint     `json:"productId" gorm:"column:product_id"`
    Name             string   `json:"name"`
    ConversionFactor float64  `json:"conversionFactor" gorm:"column:conversion_factor;default:1"`
    ConvertsToID     *uint    `json:"convertsToId,omitempty" gorm:"column:converts_to_id"`
    ToBaseUnit       float64  `json:"toBaseUnit" gorm:"column:to_base_unit;default:1"`
    IsBase           bool     `json:"isBase" gorm:"column:is_base;default:false"`
}

type ProductVariant struct {
    ID           string               `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
    ProductID    uint                 `json:"productId" gorm:"column:product_id"`
    SKU          string               `json:"sku,omitempty"`
    Barcode      string               `json:"barcode,omitempty"`
    CurrentStock int                  `json:"currentStock" gorm:"column:current_stock;default:0"`
    Attributes   []VariantAttribute   `json:"attributes" gorm:"foreignKey:VariantID"`
    Images       []VariantImage       `json:"images" gorm:"foreignKey:VariantID"`
    PricingTiers []VariantPricingTier `json:"pricingTiers" gorm:"foreignKey:VariantID"`
    Racks        []Rack               `json:"racks" gorm:"many2many:variant_racks;"`
    CreatedAt    time.Time            `json:"createdAt"`
    UpdatedAt    time.Time            `json:"updatedAt"`
}

type VariantAttribute struct {
    ID             uint   `json:"id" gorm:"primaryKey"`
    VariantID      string `json:"variantId" gorm:"column:variant_id;type:uuid"`
    AttributeName  string `json:"attributeName" gorm:"column:attribute_name"`
    AttributeValue string `json:"attributeValue" gorm:"column:attribute_value"`
}

type VariantImage struct {
    ID        uint   `json:"id" gorm:"primaryKey"`
    VariantID string `json:"variantId" gorm:"column:variant_id;type:uuid"`
    ImageURL  string `json:"imageUrl" gorm:"column:image_url"`
    SortOrder int    `json:"sortOrder" gorm:"column:sort_order;default:0"`
}

type VariantPricingTier struct {
    ID        uint    `json:"id" gorm:"primaryKey"`
    VariantID string  `json:"variantId" gorm:"column:variant_id;type:uuid"`
    MinQty    int     `json:"minQty" gorm:"column:min_qty"`
    Value     float64 `json:"value"`
}
```

---

## 3. Product API Endpoints

### 3.1 List Products — `GET /api/v1/products`

**Permission**: `Master Data > Product > read`

**Query Parameters:**

| Param | Type | Default | Description |
|-------|------|---------|-------------|
| `page` | int | 1 | Page number |
| `pageSize` | int | 10 | Items per page |
| `search` | string | — | Search by product name |
| `sortBy` | string | `id` | Sort: `id`, `name`, `category`, `status` |
| `sortDir` | string | `asc` | Sort direction |
| `status` | string | — | Filter: `active`, `inactive` |
| `categoryId` | int | — | Filter by category |
| `supplierId` | int | — | Filter by supplier |

**Response (200):**

```json
{
  "data": [
    {
      "id": 1,
      "name": "T-Shirt",
      "description": "Cotton t-shirt",
      "categoryId": 1,
      "category": { "id": 1, "name": "Clothing" },
      "priceSetting": "fixed",
      "markupType": null,
      "hasVariants": true,
      "status": "active",
      "images": [
        { "id": 1, "imageUrl": "/uploads/products/tshirt-1.jpg", "sortOrder": 0 }
      ],
      "suppliers": [
        { "id": 1, "name": "PT Sumber Makmur" }
      ],
      "variantCount": 12,
      "createdAt": "2026-01-15T10:00:00Z"
    }
  ],
  "meta": { "page": 1, "pageSize": 10, "totalItems": 4, "totalPages": 1 }
}
```

**Notes:**
- List view returns a summary: product images, supplier names, and `variantCount` (computed)
- Do NOT include full variant data in the list — that's too heavy. Include variant count only.
- Eager-load: category, images (first image only for thumbnail), suppliers (names only)

### 3.2 Get Product — `GET /api/v1/products/:id`

**Permission**: `Master Data > Product > read`

Returns the **full product** with all nested data: images, suppliers, units, variants (with attributes, images, pricing tiers, racks).

**Response (200):**

```json
{
  "data": {
    "id": 1,
    "name": "T-Shirt",
    "description": "Cotton t-shirt",
    "categoryId": 1,
    "category": { "id": 1, "name": "Clothing" },
    "priceSetting": "fixed",
    "markupType": null,
    "hasVariants": true,
    "status": "active",
    "images": [
      { "id": 1, "imageUrl": "/uploads/products/tshirt-1.jpg", "sortOrder": 0 }
    ],
    "suppliers": [
      { "id": 1, "name": "PT Sumber Makmur" },
      { "id": 2, "name": "CV Jaya Abadi" }
    ],
    "units": [
      { "id": 1, "name": "Pcs", "conversionFactor": 1, "convertsToId": null, "toBaseUnit": 1, "isBase": true },
      { "id": 2, "name": "Dozen", "conversionFactor": 12, "convertsToId": 1, "toBaseUnit": 12, "isBase": false },
      { "id": 3, "name": "Box", "conversionFactor": 12, "convertsToId": 2, "toBaseUnit": 144, "isBase": false }
    ],
    "variants": [
      {
        "id": "uuid-1",
        "sku": "TS-R-S",
        "barcode": "8901234567890",
        "currentStock": 50,
        "attributes": [
          { "attributeName": "Color", "attributeValue": "Red" },
          { "attributeName": "Size", "attributeValue": "S" }
        ],
        "images": [],
        "pricingTiers": [
          { "minQty": 1, "value": 75000 },
          { "minQty": 12, "value": 70000 }
        ],
        "racks": [
          { "id": 1, "name": "Main Display" }
        ]
      }
    ],
    "createdAt": "2026-01-15T10:00:00Z",
    "updatedAt": "2026-01-15T10:00:00Z"
  }
}
```

### 3.3 Create Product — `POST /api/v1/products`

**Permission**: `Master Data > Product > create`

**Request Body:**

```json
{
  "name": "T-Shirt",
  "description": "Cotton t-shirt",
  "categoryId": 1,
  "priceSetting": "fixed",
  "markupType": null,
  "hasVariants": true,
  "status": "active",
  "supplierIds": [1, 2],
  "units": [
    { "name": "Pcs", "isBase": true },
    { "name": "Dozen", "conversionFactor": 12, "convertsToName": "Pcs" },
    { "name": "Box", "conversionFactor": 12, "convertsToName": "Dozen" }
  ],
  "variants": [
    {
      "sku": "TS-R-S",
      "barcode": "8901234567890",
      "attributes": [
        { "attributeName": "Color", "attributeValue": "Red" },
        { "attributeName": "Size", "attributeValue": "S" }
      ],
      "pricingTiers": [
        { "minQty": 1, "value": 75000 },
        { "minQty": 12, "value": 70000 }
      ],
      "rackIds": [1]
    }
  ]
}
```

**Validation:**

Product-level:
- `name`: required, 1-255 characters
- `categoryId`: required, must reference an existing category
- `priceSetting`: required, must be `fixed` or `markup`
- `markupType`: required if `priceSetting` is `markup` (must be `percentage` or `fixed_amount`), must be null/omitted if `priceSetting` is `fixed`
- `status`: must be `active` or `inactive`, default `active`
- `supplierIds`: optional, each must reference an existing active supplier

Unit-level:
- Exactly one unit with `isBase: true`
- Unit name required, unique within the product (case-insensitive)
- `conversionFactor` must be > 0 for non-base units
- No circular references in unit conversion chain
- `convertsToName` must reference another unit in the same request

Variant-level:
- At least one variant required
- If `hasVariants: false`, exactly one variant (no attributes)
- If `hasVariants: true`, at least one variant with attributes
- `sku`: optional, but must be globally unique across all products if provided
- `barcode`: optional, but must be globally unique across all products if provided
- `pricingTiers`: at least one tier required, first tier must have `minQty: 1`, tiers in ascending `minQty` order
- `rackIds`: optional, each must reference an existing active rack

**Business Logic (inside a database transaction):**
1. Create product record
2. Create product-supplier associations
3. Create units — resolve `convertsToName` references to actual IDs, calculate `toBaseUnit` for each unit
4. Create variants with attributes, pricing tiers, and rack associations
5. Return the fully loaded product

**Unit `toBaseUnit` Calculation:**
- Base unit: `toBaseUnit = 1`
- Other units: `toBaseUnit = conversionFactor × convertsTo.toBaseUnit`
- Process units in dependency order (base first, then units referencing base, etc.)

**Response (201):** Full product object (same shape as GET).

### 3.4 Update Product — `PUT /api/v1/products/:id`

**Permission**: `Master Data > Product > update`

**Request Body:** Same shape as create.

**Business Logic (inside a database transaction):**
1. Update product fields
2. Sync supplier associations (delete all, re-insert)
3. Sync units: compare existing units with request
   - **Important**: If variant stock exists (`currentStock > 0` on any variant), reject unit changes. Return `409`: "Cannot modify units while stock exists."
   - If no stock, delete existing units and recreate from request
4. Sync variants: compare by ID for existing, create new ones, delete removed ones
   - **Important**: Do NOT reset `currentStock` when updating variants. Preserve stock values.
   - If a variant is deleted that has `currentStock > 0`, return `409`: "Cannot delete variant with existing stock."
5. Sync images, pricing tiers, attributes, rack assignments for each variant
6. Return fully loaded product

**Response (200):** Full product object.

### 3.5 Delete Product — `DELETE /api/v1/products/:id`

**Permission**: `Master Data > Product > delete`

**Business Logic:**
- Check if any variant has `currentStock > 0` → return `409`: "Cannot delete product with existing stock."
- Check if product is referenced by purchase orders → return `409`: "Cannot delete product. It is referenced by {n} purchase order(s)."
- Delete product (CASCADE removes all nested data)

**Response (200):**

```json
{
  "message": "Product deleted successfully"
}
```

---

## 4. Image Upload Endpoints

### 4.1 Upload Product Image — `POST /api/v1/products/:id/images`

**Permission**: `Master Data > Product > update`

**Request**: `multipart/form-data` with `image` field (supports multiple files).

**Validation:**
- File type: JPEG, PNG, WebP
- Max file size: 5 MB per image
- Max 10 images per product

**Business Logic:**
1. Validate file(s)
2. Save to `backend/uploads/products/{productId}_{timestamp}_{index}.{ext}`
3. Create `product_images` records with auto-incrementing `sort_order`
4. Return created image records

**Response (201):**

```json
{
  "data": [
    { "id": 10, "imageUrl": "/uploads/products/1_1708099200_0.jpg", "sortOrder": 0 }
  ]
}
```

### 4.2 Delete Product Image — `DELETE /api/v1/products/:id/images/:imageId`

**Permission**: `Master Data > Product > update`

**Business Logic:**
1. Delete the file from disk
2. Delete the database record
3. Re-order remaining images

### 4.3 Reorder Product Images — `PUT /api/v1/products/:id/images/reorder`

**Permission**: `Master Data > Product > update`

**Request Body:**

```json
{
  "imageIds": [3, 1, 2]
}
```

Sets `sort_order` based on array position (index 0 = primary image).

### 4.4 Upload Variant Image — `POST /api/v1/products/:id/variants/:variantId/images`

Same pattern as product images but for variant-specific images. Stored in `variant_images` table.

---

## 5. Seed Data

Seed products that match the frontend mock data:

### Product 1: T-Shirt (with variants, linear + branching units, fixed wholesale pricing)

- **Category**: Clothing
- **Suppliers**: PT Sumber Makmur, CV Jaya Abadi
- **Price Setting**: fixed
- **Has Variants**: true
- **Units**: Pcs (base) → Dozen (12 Pcs) → Box (12 Dozen = 144 Pcs), Bag (50 Pcs)
- **Variants**: Color (Red, Blue) × Size (S, M, L) = 6 variants
  - Each with tiered pricing: [{minQty: 1, value: 75000}, {minQty: 12, value: 70000}]
  - Racks: Main Display
  - Stock: mixed (0, 25, 50, 100)

### Product 2: Rice (branching units, fixed pricing, no variants)

- **Category**: Food & Beverages
- **Suppliers**: UD Berkah Sentosa
- **Price Setting**: fixed
- **Has Variants**: false
- **Units**: Kg (base), Karung (50 Kg), Bag (25 Kg)
- **Single variant**: [{minQty: 1, value: 15000}, {minQty: 50, value: 14000}]
- Racks: Bulk Storage
- Stock: 200

### Product 3: Notebook (no variants, markup percentage)

- **Category**: Stationery
- **Suppliers**: CV Jaya Abadi
- **Price Setting**: markup, markup_type: percentage
- **Has Variants**: false
- **Units**: Pcs (base), Carton (48 Pcs)
- **Single variant**: [{minQty: 1, value: 25}] (25% markup)
- Racks: Main Display
- Stock: 150

### Product 4: Cooking Oil (simplest product)

- **Category**: Household
- **Suppliers**: (none)
- **Price Setting**: fixed
- **Has Variants**: false
- **Units**: Liter (base)
- **Single variant**: [{minQty: 1, value: 28000}]
- Racks: Cold Storage
- Stock: 5

---

## 6. TDD Workflow

**Follow strict TDD. Product is the most complex entity — thorough tests are critical.**

### 6.1 Unit Conversion Logic (pure function, test first)

**`services/unit_conversion_test.go`**:
- `TestCalculateToBaseUnit_BaseUnit_Returns1`
- `TestCalculateToBaseUnit_DirectReference_ReturnsCorrectValue`
- `TestCalculateToBaseUnit_ChainedReference_MultipliesCorrectly` — e.g., Box → Dozen → Pcs = 144
- `TestCalculateToBaseUnit_BranchingStructure_CalculatesIndependently`
- `TestValidateUnitCircularRef_NoCircle_ReturnsNil`
- `TestValidateUnitCircularRef_SelfReference_ReturnsError`
- `TestValidateUnitCircularRef_IndirectCircle_ReturnsError`
- `TestResolveUnitDependencyOrder_ReturnsBaseFirst`

### 6.2 Product Validation (service layer)

**`services/product_validation_test.go`**:
- `TestValidateProduct_ValidMinimal_ReturnsNil`
- `TestValidateProduct_MissingName_ReturnsError`
- `TestValidateProduct_MissingCategory_ReturnsError`
- `TestValidateProduct_MarkupWithoutMarkupType_ReturnsError`
- `TestValidateProduct_FixedWithMarkupType_ReturnsError`
- `TestValidateProduct_NoBaseUnit_ReturnsError`
- `TestValidateProduct_MultipleBaseUnits_ReturnsError`
- `TestValidateProduct_DuplicateUnitNames_ReturnsError`
- `TestValidateProduct_NoVariants_ReturnsError`
- `TestValidateProduct_HasVariantsFalseMultipleVariants_ReturnsError`
- `TestValidateProduct_DuplicateSKU_ReturnsError`
- `TestValidateProduct_PricingTiersMissingMinQty1_ReturnsError`
- `TestValidateProduct_PricingTiersNotAscending_ReturnsError`

### 6.3 Product Repository

**`repositories/product_repository_test.go`**:
- `TestCreateProduct_FullNested_CreatesAllRecords`
- `TestCreateProduct_WithUnits_ResolvesConvertsToIDs`
- `TestCreateProduct_SKUGloballyUnique_ReturnsError`
- `TestGetProduct_EagerLoadsAllRelations`
- `TestListProducts_Pagination_Works`
- `TestListProducts_FilterByCategory_Works`
- `TestListProducts_FilterBySupplier_Works`
- `TestListProducts_FilterByStatus_Works`
- `TestListProducts_Search_MatchesName`
- `TestListProducts_IncludesVariantCount`
- `TestUpdateProduct_SyncUnits_ReplacesUnits`
- `TestUpdateProduct_SyncVariants_PreservesStock`
- `TestUpdateProduct_DeleteVariantWithStock_ReturnsError`
- `TestDeleteProduct_WithStock_ReturnsError`
- `TestDeleteProduct_NoStock_CascadesAll`

### 6.4 Product Service

**`services/product_service_test.go`**:
- `TestCreateProduct_Valid_Succeeds`
- `TestCreateProduct_InvalidCategory_ReturnsNotFound`
- `TestCreateProduct_InvalidSupplier_ReturnsNotFound`
- `TestCreateProduct_CircularUnits_ReturnsError`
- `TestCreateProduct_GlobalDuplicateSKU_ReturnsConflict`
- `TestCreateProduct_GlobalDuplicateBarcode_ReturnsConflict`
- `TestUpdateProduct_UnitsWithExistingStock_ReturnsConflict`
- `TestUpdateProduct_DeleteVariantWithStock_ReturnsConflict`
- `TestUpdateProduct_PreservesStockValues`
- `TestDeleteProduct_WithStock_ReturnsConflict`
- `TestDeleteProduct_ReferencedByPO_ReturnsConflict`

### 6.5 Product Handler (Integration Tests)

**`handlers/product_handler_test.go`**:
- `TestListProducts_Returns200WithVariantCount`
- `TestListProducts_FilterByCategory_ReturnsFiltered`
- `TestListProducts_FilterBySupplier_ReturnsFiltered`
- `TestGetProduct_ReturnsFullNestedData`
- `TestGetProduct_NotFound_Returns404`
- `TestCreateProduct_MinimalValid_Returns201`
- `TestCreateProduct_FullComplex_Returns201` — product with units, variants, attributes, pricing, racks
- `TestCreateProduct_MissingName_Returns400`
- `TestCreateProduct_InvalidCategory_Returns400`
- `TestCreateProduct_NoBaseUnit_Returns400`
- `TestCreateProduct_CircularUnits_Returns400`
- `TestCreateProduct_DuplicateSKU_Returns409`
- `TestCreateProduct_NoAuth_Returns401`
- `TestCreateProduct_NoPermission_Returns403`
- `TestUpdateProduct_Valid_Returns200`
- `TestUpdateProduct_UnitsWithStock_Returns409`
- `TestDeleteProduct_NoStock_Returns200`
- `TestDeleteProduct_WithStock_Returns409`
- `TestDeleteProduct_ReferencedByPO_Returns409`

### 6.6 Image Upload Handler Tests

**`handlers/product_image_handler_test.go`**:
- `TestUploadProductImage_ValidJPEG_Returns201`
- `TestUploadProductImage_ValidPNG_Returns201`
- `TestUploadProductImage_InvalidType_Returns400`
- `TestUploadProductImage_TooLarge_Returns400`
- `TestUploadProductImage_MaxImagesExceeded_Returns400`
- `TestDeleteProductImage_Exists_Returns200`
- `TestDeleteProductImage_NotFound_Returns404`
- `TestReorderImages_ValidOrder_Returns200`
- `TestUploadVariantImage_ValidImage_Returns201`

---

## 7. Deliverables

After completing this stage:

1. All product-related tables exist with proper foreign keys and indexes
2. Product CRUD works with full nested data (units, variants, attributes, pricing, images, racks, suppliers)
3. Product creation/update handles complex transactional logic (all-or-nothing)
4. Unit conversion chains are validated (no circular refs) and `toBaseUnit` is auto-calculated
5. SKU and barcode uniqueness is enforced globally
6. Stock-related protections work (can't modify units or delete variants/products with stock)
7. Image upload, delete, and reorder work for both products and variants
8. Product list endpoint supports filtering by category, supplier, and status
9. Seed data includes 4 products with diverse configurations
10. **All tests pass** (`go test ./...`)
11. **Test coverage** ≥ 80% across handlers and services
12. **Unit conversion logic** has 100% test coverage (critical business logic)
