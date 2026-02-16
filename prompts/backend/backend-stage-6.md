# Backend Stage 6 — Transactions: Purchase Orders & Sales

## Overview

Build the Purchase Order workflow (create, send, receive, complete, cancel) and the Sales Transaction API (product search, checkout, stock deduction, receipt). These are the transactional modules that modify stock.

> **Prerequisite**: Stage 5 must be complete (products with units, variants, and stock exist in the database).

---

## 1. Database Migrations

### 1.1 Purchase Orders Table

```sql
CREATE TABLE purchase_orders (
    id                       BIGSERIAL PRIMARY KEY,
    po_number                VARCHAR(20) NOT NULL UNIQUE,       -- format: PO-YYYY-NNNN
    supplier_id              BIGINT NOT NULL REFERENCES suppliers(id),
    date                     DATE NOT NULL,
    status                   VARCHAR(20) NOT NULL DEFAULT 'draft',  -- draft, sent, received, completed, cancelled
    notes                    TEXT,
    received_date            TIMESTAMPTZ,
    payment_method           VARCHAR(20),                       -- cash, credit_card, bank_transfer
    supplier_bank_account_id UUID REFERENCES supplier_bank_accounts(id),
    subtotal                 DECIMAL(15,2),
    total_items              INTEGER,
    created_at               TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at               TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_purchase_orders_supplier_id ON purchase_orders(supplier_id);
CREATE INDEX idx_purchase_orders_status ON purchase_orders(status);
CREATE INDEX idx_purchase_orders_po_number ON purchase_orders(po_number);
CREATE INDEX idx_purchase_orders_date ON purchase_orders(date DESC);
```

### 1.2 Purchase Order Items Table

```sql
CREATE TABLE purchase_order_items (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    purchase_order_id BIGINT NOT NULL REFERENCES purchase_orders(id) ON DELETE CASCADE,
    product_id        BIGINT NOT NULL REFERENCES products(id),
    variant_id        UUID NOT NULL REFERENCES product_variants(id),
    unit_id           BIGINT NOT NULL REFERENCES product_units(id),
    unit_name         VARCHAR(100) NOT NULL,       -- denormalized
    product_name      VARCHAR(255) NOT NULL,       -- denormalized
    variant_label     VARCHAR(255) NOT NULL,        -- denormalized, e.g., "Red / L" or "Default"
    sku               VARCHAR(100),                -- denormalized
    current_stock     INTEGER NOT NULL DEFAULT 0,  -- snapshot at PO creation time
    ordered_qty       INTEGER NOT NULL CHECK (ordered_qty > 0),
    price             DECIMAL(15,2) NOT NULL DEFAULT 0,
    received_qty      INTEGER,
    received_price    DECIMAL(15,2),
    is_verified       BOOLEAN DEFAULT false
);

CREATE INDEX idx_po_items_purchase_order_id ON purchase_order_items(purchase_order_id);
CREATE INDEX idx_po_items_variant_id ON purchase_order_items(variant_id);
```

### 1.3 Sales Transactions Table

```sql
CREATE TABLE sales_transactions (
    id                 BIGSERIAL PRIMARY KEY,
    transaction_number VARCHAR(30) NOT NULL UNIQUE,  -- format: TRX-YYYY-NNNNNN
    date               TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    subtotal           DECIMAL(15,2) NOT NULL,
    grand_total        DECIMAL(15,2) NOT NULL,
    total_items        INTEGER NOT NULL,
    payment_method     VARCHAR(20) NOT NULL,         -- cash, card, qris
    created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_sales_transactions_date ON sales_transactions(date DESC);
CREATE INDEX idx_sales_transactions_number ON sales_transactions(transaction_number);
```

### 1.4 Sales Transaction Items Table

```sql
CREATE TABLE sales_transaction_items (
    id              BIGSERIAL PRIMARY KEY,
    transaction_id  BIGINT NOT NULL REFERENCES sales_transactions(id) ON DELETE CASCADE,
    product_id      BIGINT NOT NULL REFERENCES products(id),
    variant_id      UUID NOT NULL REFERENCES product_variants(id),
    unit_id         BIGINT NOT NULL REFERENCES product_units(id),
    product_name    VARCHAR(255) NOT NULL,       -- denormalized
    variant_label   VARCHAR(255) NOT NULL,       -- denormalized
    sku             VARCHAR(100),                -- denormalized
    unit_name       VARCHAR(100) NOT NULL,       -- denormalized
    quantity        INTEGER NOT NULL CHECK (quantity > 0),
    base_qty        INTEGER NOT NULL,            -- quantity × unit.toBaseUnit
    unit_price      DECIMAL(15,2) NOT NULL,      -- price per selected unit
    total_price     DECIMAL(15,2) NOT NULL       -- quantity × unit_price
);

CREATE INDEX idx_sales_items_transaction_id ON sales_transaction_items(transaction_id);
CREATE INDEX idx_sales_items_variant_id ON sales_transaction_items(variant_id);
```

### 1.5 Stock Movements Table (Audit Trail)

```sql
CREATE TABLE stock_movements (
    id              BIGSERIAL PRIMARY KEY,
    variant_id      UUID NOT NULL REFERENCES product_variants(id),
    movement_type   VARCHAR(20) NOT NULL,        -- purchase_receive, sales, adjustment
    quantity         INTEGER NOT NULL,             -- positive for inbound, negative for outbound
    reference_type  VARCHAR(20),                  -- purchase_order, sales_transaction
    reference_id    BIGINT,                       -- ID of the PO or sale transaction
    notes           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_stock_movements_variant_id ON stock_movements(variant_id);
CREATE INDEX idx_stock_movements_type ON stock_movements(movement_type);
```

---

## 2. Purchase Order API

### 2.1 List Purchase Orders — `GET /api/v1/purchase-orders`

**Permission**: `Transaction > Purchase > read`

**Query Parameters:**

| Param | Type | Default | Description |
|-------|------|---------|-------------|
| `page` | int | 1 | Page number |
| `pageSize` | int | 10 | Items per page |
| `search` | string | — | Search by PO number or supplier name |
| `status` | string | — | Filter: `draft`, `sent`, `received`, `completed`, `cancelled` |
| `supplierId` | int | — | Filter by supplier |
| `sortBy` | string | `date` | Sort: `date`, `poNumber`, `status` |
| `sortDir` | string | `desc` | Sort direction (default newest first) |

**Response (200):**

```json
{
  "data": [
    {
      "id": 1,
      "poNumber": "PO-2026-0001",
      "supplierId": 1,
      "supplierName": "PT Sumber Makmur",
      "date": "2026-02-12",
      "status": "draft",
      "notes": "First restocking order",
      "itemCount": 4,
      "totalOrderedQty": 225,
      "estimatedTotal": 0,
      "createdAt": "2026-02-12T10:00:00Z"
    }
  ],
  "meta": { "page": 1, "pageSize": 10, "totalItems": 4, "totalPages": 1 },
  "statusCounts": {
    "all": 4,
    "draft": 1,
    "sent": 1,
    "received": 1,
    "completed": 1,
    "cancelled": 0
  }
}
```

**Notes:**
- Include `statusCounts` for the tab filters in the frontend
- Summary fields (`itemCount`, `totalOrderedQty`, `estimatedTotal`) are computed from items

### 2.2 Get Purchase Order — `GET /api/v1/purchase-orders/:id`

**Permission**: `Transaction > Purchase > read`

Returns full PO with all items and receive data.

**Response (200):**

```json
{
  "data": {
    "id": 1,
    "poNumber": "PO-2026-0001",
    "supplierId": 1,
    "supplierName": "PT Sumber Makmur",
    "date": "2026-02-12",
    "status": "received",
    "notes": "First restocking order",
    "receivedDate": "2026-02-13T14:30:00Z",
    "paymentMethod": "bank_transfer",
    "supplierBankAccountId": "uuid-bank-1",
    "supplierBankAccountLabel": "BCA - 1234567890",
    "subtotal": 18500000,
    "totalItems": 220,
    "items": [
      {
        "id": "uuid-item-1",
        "productId": 1,
        "productName": "T-Shirt",
        "variantId": "uuid-variant-1",
        "variantLabel": "Red / S",
        "sku": "TS-R-S",
        "unitId": 1,
        "unitName": "Pcs",
        "currentStock": 50,
        "orderedQty": 50,
        "price": 0,
        "receivedQty": 50,
        "receivedPrice": 45000,
        "isVerified": true
      }
    ],
    "createdAt": "2026-02-12T10:00:00Z",
    "updatedAt": "2026-02-13T14:30:00Z"
  }
}
```

### 2.3 Create Purchase Order — `POST /api/v1/purchase-orders`

**Permission**: `Transaction > Purchase > create`

**Request Body:**

```json
{
  "supplierId": 1,
  "date": "2026-02-12",
  "notes": "First restocking order",
  "items": [
    {
      "productId": 1,
      "variantId": "uuid-variant-1",
      "unitId": 1,
      "orderedQty": 50,
      "price": 0
    }
  ]
}
```

**Validation:**
- `supplierId`: required, must reference an existing active supplier
- `date`: required, valid date
- `items`: required, at least one item with `orderedQty > 0`
- Each item's `productId`, `variantId`, `unitId` must reference existing records
- `unitId` must belong to the specified product

**Business Logic:**
1. Generate next PO number (`PO-YYYY-NNNN`)
2. For each item, denormalize: `productName`, `variantLabel`, `sku`, `unitName`, `currentStock` (snapshot)
3. Set status to `draft`
4. Save PO and items in a transaction

**PO Number Generation:**
- Query the latest PO number for the current year
- Increment the sequence number
- Format: `PO-{year}-{sequence zero-padded to 4 digits}`
- Handle concurrent creation with a database sequence or advisory lock

**Response (201):** Full PO object.

### 2.4 Update Purchase Order — `PUT /api/v1/purchase-orders/:id`

**Permission**: `Transaction > Purchase > update`

**Business Logic:**
- Only allow editing POs with status `draft` → return `403` otherwise
- Changing the supplier clears and replaces all items
- Sync items: delete existing, insert new (full replace)

**Response (200):** Full PO object.

### 2.5 Delete Purchase Order — `DELETE /api/v1/purchase-orders/:id`

**Permission**: `Transaction > Purchase > delete`

**Business Logic:**
- Only allow deleting POs with status `draft` → return `403` otherwise
- Delete PO and items

**Response (200):**

```json
{
  "message": "Purchase order deleted successfully"
}
```

### 2.6 Update PO Status — `PATCH /api/v1/purchase-orders/:id/status`

**Permission**: `Transaction > Purchase > update`

**Request Body:**

```json
{
  "status": "sent"
}
```

**Allowed Transitions:**

| Current Status | Allowed Next Status |
|---------------|---------------------|
| draft | sent, cancelled |
| sent | cancelled |
| received | completed |

Any other transition → return `400`: "Invalid status transition from {current} to {requested}."

**Business Logic per transition:**
- `draft → sent`: Update status. (Future: send email/notification to supplier.)
- `draft/sent → cancelled`: Update status. No stock changes.
- `received → completed`: Update status. Administrative finalization only.

**Response (200):** Updated PO object.

### 2.7 Receive Purchase Order — `POST /api/v1/purchase-orders/:id/receive`

**Permission**: `Transaction > Purchase > update`

**Precondition:** PO status must be `sent` (or `draft` if allowing direct receive — see note below).

**Request Body:**

```json
{
  "receivedDate": "2026-02-13T14:30:00Z",
  "paymentMethod": "bank_transfer",
  "supplierBankAccountId": "uuid-bank-1",
  "items": [
    {
      "itemId": "uuid-item-1",
      "receivedQty": 50,
      "receivedPrice": 45000,
      "isVerified": true
    }
  ]
}
```

**Validation:**
- `receivedDate`: required
- `paymentMethod`: required, one of `cash`, `credit_card`, `bank_transfer`
- `supplierBankAccountId`: required if `paymentMethod` is not `cash`, must belong to the PO's supplier
- Each item: `receivedQty` >= 0, `receivedPrice` >= 0
- All PO items must be included in the request

**Business Logic (inside a database transaction):**

1. Validate PO status is `sent` (or `draft`)
2. Update PO:
   - Set status to `received`
   - Set `receivedDate`, `paymentMethod`, `supplierBankAccountId`
   - Calculate `subtotal` = SUM(`receivedQty` × `receivedPrice`) for all items
   - Calculate `totalItems` = SUM(`receivedQty`) for all items
3. Update each PO item with `receivedQty`, `receivedPrice`, `isVerified`
4. **Update variant stock** for each item:
   - Calculate base-unit quantity: `stockDelta = receivedQty × unit.toBaseUnit`
   - `UPDATE product_variants SET current_stock = current_stock + stockDelta WHERE id = variantId`
5. **Create stock movement records** for each item:
   - `movement_type`: `purchase_receive`
   - `quantity`: positive (stockDelta)
   - `reference_type`: `purchase_order`
   - `reference_id`: PO ID
6. Return updated PO

**Response (200):** Full PO object with receive data.

> **Note on Draft → Received**: Optionally allow receiving directly from `draft` status (for walk-in purchases where goods arrive immediately). This skips the `sent` step. To support this, accept status `draft` or `sent` as precondition.

### 2.8 Get Products for PO — `GET /api/v1/purchase-orders/products`

**Permission**: `Transaction > Purchase > read`

Helper endpoint for the PO form to load products matching a supplier.

**Query Parameters:**

| Param | Type | Description |
|-------|------|-------------|
| `supplierId` | int | Required. Supplier ID to filter products by. |
| `search` | string | Optional. Search product name. |

**Business Logic:**
- Return products where:
  - Product has the specified supplier in `product_suppliers`, OR
  - Product has no suppliers assigned (`product_suppliers` is empty)
- Only return active products
- For each product, include: units and variants (with `currentStock`, `sku`, attributes)
- Exclude variants with no data (shouldn't happen, but defensive)

**Response (200):**

```json
{
  "data": [
    {
      "id": 1,
      "name": "T-Shirt",
      "categoryName": "Clothing",
      "units": [
        { "id": 1, "name": "Pcs", "toBaseUnit": 1, "isBase": true },
        { "id": 2, "name": "Dozen", "toBaseUnit": 12, "isBase": false }
      ],
      "variants": [
        {
          "id": "uuid-1",
          "sku": "TS-R-S",
          "label": "Red / S",
          "currentStock": 50
        }
      ]
    }
  ]
}
```

---

## 3. Sales Transaction API

### 3.1 Product Search — `GET /api/v1/sales/products/search`

**Permission**: `Transaction > Sales > read`

Endpoint for the POS search bar.

**Query Parameters:**

| Param | Type | Description |
|-------|------|-------------|
| `q` | string | Required, min 3 characters. Search query. |

**Business Logic:**
- Search against: product name, variant SKU, variant barcode (case-insensitive, partial match)
- Exclude inactive products (`status != 'active'`)
- Limit to 10 products max
- For each product, return: images, units, variants with stock/attributes/images/pricing

**Response (200):**

```json
{
  "data": [
    {
      "id": 1,
      "name": "T-Shirt",
      "description": "Cotton t-shirt",
      "hasVariants": true,
      "priceSetting": "fixed",
      "markupType": null,
      "images": [
        { "imageUrl": "/uploads/products/tshirt-1.jpg", "sortOrder": 0 }
      ],
      "units": [
        { "id": 1, "name": "Pcs", "toBaseUnit": 1, "isBase": true },
        { "id": 2, "name": "Dozen", "toBaseUnit": 12, "isBase": false }
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
          ]
        }
      ]
    }
  ]
}
```

### 3.2 Checkout — `POST /api/v1/sales/checkout`

**Permission**: `Transaction > Sales > create`

**Request Body:**

```json
{
  "paymentMethod": "cash",
  "items": [
    {
      "productId": 1,
      "variantId": "uuid-1",
      "unitId": 1,
      "quantity": 2
    },
    {
      "productId": 2,
      "variantId": "uuid-5",
      "unitId": 5,
      "quantity": 1
    }
  ]
}
```

**Validation:**
- `paymentMethod`: required, one of `cash`, `card`, `qris`
- `items`: required, at least one item
- Each item: `quantity` > 0, valid `productId`, `variantId`, `unitId`
- Each `unitId` must belong to the specified product

**Business Logic (inside a database transaction):**

1. For each item:
   a. Load variant with current stock
   b. Load unit to get `toBaseUnit`
   c. Calculate `baseQty = quantity × unit.toBaseUnit`
   d. **Stock check**: if `baseQty > variant.currentStock` → return `400`: "Insufficient stock for {productName} ({variantLabel}). Available: {stock} {baseUnitName}."
   e. Calculate price using tiered pricing:
      - Find the tier where `baseQty >= tier.minQty` with the highest matching `minQty`
      - `unitPrice = tier.value × unit.toBaseUnit`
      - `totalPrice = quantity × unitPrice`
   f. Denormalize: `productName`, `variantLabel`, `sku`, `unitName`

2. Calculate totals:
   - `subtotal` = SUM of all item `totalPrice`
   - `grandTotal` = subtotal (no tax/discount for now)
   - `totalItems` = COUNT of items

3. Generate transaction number: `TRX-YYYY-NNNNNN` (6-digit sequence)

4. Create `sales_transactions` record

5. Create `sales_transaction_items` records

6. **Deduct stock** for each item:
   - `UPDATE product_variants SET current_stock = current_stock - baseQty WHERE id = variantId`
   - Validate `current_stock` doesn't go negative (use `CHECK` constraint or application check)

7. **Create stock movement records** for each item:
   - `movement_type`: `sales`
   - `quantity`: negative (-baseQty)
   - `reference_type`: `sales_transaction`
   - `reference_id`: transaction ID

8. Return receipt data

**Response (201):**

```json
{
  "data": {
    "id": 1,
    "transactionNumber": "TRX-2026-000001",
    "date": "2026-02-16T14:30:00Z",
    "items": [
      {
        "productName": "T-Shirt",
        "variantLabel": "Red / S",
        "sku": "TS-R-S",
        "unitName": "Pcs",
        "quantity": 2,
        "unitPrice": 75000,
        "totalPrice": 150000
      }
    ],
    "totalItems": 2,
    "subtotal": 178000,
    "grandTotal": 178000,
    "paymentMethod": "cash"
  },
  "message": "Transaction completed successfully"
}
```

### 3.3 List Sales Transactions — `GET /api/v1/sales/transactions`

**Permission**: `Transaction > Sales > read`

**Query Parameters:**

| Param | Type | Default | Description |
|-------|------|---------|-------------|
| `page` | int | 1 | Page number |
| `pageSize` | int | 10 | Items per page |
| `search` | string | — | Search by transaction number |
| `dateFrom` | date | — | Filter from date |
| `dateTo` | date | — | Filter to date |
| `paymentMethod` | string | — | Filter: `cash`, `card`, `qris` |
| `sortBy` | string | `date` | Sort: `date`, `transactionNumber`, `grandTotal` |
| `sortDir` | string | `desc` | Sort direction |

**Response (200):** Paginated list of transactions with summary (no items detail).

### 3.4 Get Sales Transaction — `GET /api/v1/sales/transactions/:id`

**Permission**: `Transaction > Sales > read`

Returns the full transaction with items (receipt data).

**Response (200):** Full transaction object (same as checkout response).

---

## 4. Currency & Formatting

All monetary values are stored as `DECIMAL(15,2)` in the database. The API returns raw numbers. Currency formatting (IDR, Rp, dot separator) is handled by the frontend.

---

## 5. Seed Data

### 5.1 Purchase Orders

Seed 4 POs matching the frontend mock data:

| PO Number | Supplier | Date | Status | Items | Notes |
|-----------|----------|------|--------|-------|-------|
| PO-2026-0001 | PT Sumber Makmur | 2026-02-05 | completed | T-Shirt (3 variants × 50 pcs each), Notebook (1 × 100 pcs) | First restocking order |
| PO-2026-0002 | CV Jaya Abadi | 2026-02-08 | received | Notebook (1 × 50 pcs) | Urgent notebook restock |
| PO-2026-0003 | UD Berkah Sentosa | 2026-02-10 | sent | Rice (1 × 100 kg) | Monthly rice order |
| PO-2026-0004 | PT Sumber Makmur | 2026-02-12 | draft | T-Shirt (2 variants × 25 pcs each) | Pending review |

For completed/received POs, include realistic receive data and ensure variant `currentStock` values are consistent with received quantities.

### 5.2 Stock Movements

Create corresponding stock movement records for received POs (PO-2026-0001 and PO-2026-0002).

---

## 6. TDD Workflow

**Follow strict TDD. Transaction logic (stock changes, status workflow) is the highest-risk area — tests are essential.**

### 6.1 PO Number / Transaction Number Generation

**`services/sequence_test.go`**:
- `TestGeneratePONumber_FirstOfYear_ReturnsPO_YYYY_0001`
- `TestGeneratePONumber_Increment_ReturnsPO_YYYY_0002`
- `TestGeneratePONumber_NewYear_ResetsSequence`
- `TestGenerateTrxNumber_FirstEver_ReturnsTRX_YYYY_000001`
- `TestGenerateTrxNumber_Increment_Works`

### 6.2 PO Status Transitions (pure logic)

**`services/po_status_test.go`**:
- `TestValidateStatusTransition_DraftToSent_Valid`
- `TestValidateStatusTransition_DraftToCancelled_Valid`
- `TestValidateStatusTransition_SentToCancelled_Valid`
- `TestValidateStatusTransition_ReceivedToCompleted_Valid`
- `TestValidateStatusTransition_CompletedToAnything_Invalid`
- `TestValidateStatusTransition_CancelledToAnything_Invalid`
- `TestValidateStatusTransition_ReceivedToDraft_Invalid`
- `TestValidateStatusTransition_SentToDraft_Invalid`

### 6.3 Purchase Order Repository

**`repositories/po_repository_test.go`**:
- `TestCreatePO_WithItems_CreatesAll`
- `TestGetPO_EagerLoadsItems`
- `TestListPOs_FilterByStatus_Works`
- `TestListPOs_FilterBySupplier_Works`
- `TestListPOs_SearchByPONumberOrSupplier_Works`
- `TestListPOs_StatusCounts_ReturnsCorrectCounts`
- `TestUpdatePO_DraftOnly_Succeeds`
- `TestDeletePO_DraftOnly_Succeeds`

### 6.4 Purchase Order Service

**`services/po_service_test.go`**:
- `TestCreatePO_Valid_GeneratesPONumber`
- `TestCreatePO_DenormalizesItemFields`
- `TestCreatePO_InactiveSupplier_ReturnsError`
- `TestCreatePO_NoItemsWithQty_ReturnsValidation`
- `TestUpdatePO_NonDraft_ReturnsForbidden`
- `TestDeletePO_NonDraft_ReturnsForbidden`
- `TestUpdatePOStatus_ValidTransition_Succeeds`
- `TestUpdatePOStatus_InvalidTransition_ReturnsError`

### 6.5 PO Receive (critical — stock mutations)

**`services/po_receive_test.go`**:
- `TestReceivePO_SentStatus_Succeeds`
- `TestReceivePO_DraftStatus_Succeeds` (if allowing direct receive)
- `TestReceivePO_ReceivedStatus_ReturnsError` (can't receive twice)
- `TestReceivePO_CompletedStatus_ReturnsError`
- `TestReceivePO_UpdatesVariantStock_CorrectDelta`
- `TestReceivePO_MultipleItems_UpdatesAllVariants`
- `TestReceivePO_UnitConversion_CalculatesBaseQtyCorrectly` — e.g., 2 boxes × 144 toBaseUnit = 288 pcs added
- `TestReceivePO_CreatesStockMovements_ForEachItem`
- `TestReceivePO_StockMovement_PositiveQuantity`
- `TestReceivePO_CalculatesSubtotal_Correctly`
- `TestReceivePO_CalculatesTotalItems_Correctly`
- `TestReceivePO_BankTransferRequiresBankAccount`
- `TestReceivePO_CashNoBankAccount_Succeeds`
- `TestReceivePO_MissingItems_ReturnsValidation`

### 6.6 PO Handler (Integration Tests)

**`handlers/po_handler_test.go`**:
- `TestListPOs_Returns200WithStatusCounts`
- `TestListPOs_FilterByStatus_Returns200`
- `TestGetPO_WithItems_Returns200`
- `TestGetPO_NotFound_Returns404`
- `TestCreatePO_ValidBody_Returns201`
- `TestCreatePO_InvalidSupplier_Returns400`
- `TestCreatePO_NoItems_Returns400`
- `TestUpdatePO_DraftPO_Returns200`
- `TestUpdatePO_SentPO_Returns403`
- `TestDeletePO_DraftPO_Returns200`
- `TestDeletePO_SentPO_Returns403`
- `TestUpdatePOStatus_DraftToSent_Returns200`
- `TestUpdatePOStatus_InvalidTransition_Returns400`
- `TestReceivePO_ValidBody_Returns200_UpdatesStock`
- `TestReceivePO_NonSentPO_Returns400`
- `TestReceivePO_BankTransferNoBankAccount_Returns400`
- `TestGetProductsForPO_ReturnsFilteredProducts`

### 6.7 Tiered Pricing Calculation (pure function)

**`services/pricing_test.go`**:
- `TestCalculateTieredPrice_SingleTier_ReturnsBasePrice`
- `TestCalculateTieredPrice_QtyMatchesSecondTier_ReturnsSecondTierPrice`
- `TestCalculateTieredPrice_QtyBetweenTiers_UsesLowerTier`
- `TestCalculateTieredPrice_LargeQty_UsesHighestTier`
- `TestCalculateTieredPrice_WithUnitConversion_ConvertsToBaseFirst`
- `TestCalculateTieredPrice_EmptyTiers_ReturnsError`

### 6.8 Sales Checkout (critical — stock mutations)

**`services/sales_service_test.go`**:
- `TestCheckout_Valid_DeductsStock`
- `TestCheckout_InsufficientStock_ReturnsError`
- `TestCheckout_MultipleItems_DeductsAll`
- `TestCheckout_UnitConversion_DeductsCorrectBaseQty`
- `TestCheckout_TieredPricing_AppliesCorrectTier`
- `TestCheckout_TieredPricingWithUnitConversion_CalculatesCorrectly`
- `TestCheckout_CalculatesSubtotalAndGrandTotal`
- `TestCheckout_GeneratesTransactionNumber`
- `TestCheckout_CreatesStockMovements_NegativeQuantity`
- `TestCheckout_EmptyCart_ReturnsValidation`
- `TestCheckout_InvalidPaymentMethod_ReturnsValidation`
- `TestCheckout_ZeroQuantity_ReturnsValidation`
- `TestCheckout_ConcurrentCheckout_NoOverselling` — two checkouts competing for last stock item

### 6.9 Sales Handler (Integration Tests)

**`handlers/sales_handler_test.go`**:
- `TestProductSearch_MinChars_Returns200`
- `TestProductSearch_TooShort_Returns400`
- `TestProductSearch_ByName_ReturnsMatching`
- `TestProductSearch_BySKU_ReturnsMatching`
- `TestProductSearch_ByBarcode_ReturnsMatching`
- `TestProductSearch_InactiveProducts_Excluded`
- `TestProductSearch_Max10Results`
- `TestCheckout_ValidBody_Returns201WithReceipt`
- `TestCheckout_InsufficientStock_Returns400`
- `TestCheckout_NoAuth_Returns401`
- `TestCheckout_NoPermission_Returns403`
- `TestCheckout_VerifyStockDeducted` — check variant stock after checkout
- `TestCheckout_VerifyStockMovementCreated`
- `TestListTransactions_Returns200WithPagination`
- `TestListTransactions_FilterByDate_Works`
- `TestListTransactions_FilterByPaymentMethod_Works`
- `TestGetTransaction_ReturnsReceiptData`

### 6.10 Stock Movement Verification

**`repositories/stock_movement_repository_test.go`**:
- `TestCreateStockMovement_PurchaseReceive_PositiveQty`
- `TestCreateStockMovement_Sales_NegativeQty`
- `TestGetStockMovementsByVariant_ReturnsChronological`
- `TestGetStockMovementsByReference_ReturnsMatching`

---

## 7. Deliverables

After completing this stage:

1. Purchase order CRUD works with full lifecycle (draft → sent → received → completed / cancelled)
2. PO number auto-generation works with `PO-YYYY-NNNN` format
3. PO receive flow correctly updates variant stock and creates stock movement records
4. Products-for-PO helper endpoint returns filtered products with variants and units
5. POS product search works across product name, SKU, and barcode
6. Sales checkout validates stock, calculates tiered pricing, deducts stock, and returns receipt data
7. Transaction number auto-generation works with `TRX-YYYY-NNNNNN` format
8. Stock movement audit trail records all stock changes (inbound from PO, outbound from sales)
9. All endpoints are protected by permission middleware
10. Seed data includes purchase orders in various statuses with consistent stock
11. **All tests pass** (`go test ./...`)
12. **Test coverage** ≥ 80% across handlers and services
13. **Stock mutation logic** (receive, checkout) has 100% test coverage
14. **Concurrent checkout test** proves no overselling under race conditions
