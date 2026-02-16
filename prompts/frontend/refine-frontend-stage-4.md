# Frontend App - Admin Panel (Phase 4)

## Overview

Phase 4 adds **Supplier** and **Rack** master data, simplifies variant pricing (tiered-only), introduces a **product-level Price tab**, and builds the full **Purchase Order** workflow including a receive/verification flow.

> **Backend note**: All data is client-side (Zustand) for now. Design stores and data models so they can be replaced with backend API calls in the future without changing the UI.

---

## 1. Supplier Master Data (`/master/supplier`)

Add a new master data page for managing suppliers.

### 1.1 Supplier Data Model

```typescript
interface BankAccount {
  id: string;          // auto-generated UUID
  accountName: string; // free-text, e.g., "BCA - Main Account"
  accountNumber: string;
}

interface Supplier {
  id: number;
  name: string;
  address: string;
  phone: string;           // optional
  email: string;           // optional
  website: string;         // optional
  bankAccounts: BankAccount[];  // optional, multiple
  active: boolean;         // default: true
  createdAt: string;       // ISO date string
}
```

### 1.2 Supplier List View

Table-based list following existing table design (same as Master Category):

**Table columns:**

| Column | Sortable | Notes |
|--------|----------|-------|
| ID | Yes | Auto-increment |
| Name | Yes | Supplier name |
| Address | No | Supplier address |
| Phone | No | Phone number, show "—" if empty |
| Email | No | Email address, show "—" if empty |
| Status | Yes | Badge: green for Active, gray for Inactive |
| Actions | No | Edit, Delete buttons |

**Features:**
- **Search**: text search across name, address, and email.
- **Sorting**: sortable on ID, Name, Status columns.
- **Pagination**: with items-per-page selector (reuse existing Table component).
- **Add Supplier** button at the top → opens modal form.

### 1.3 Create / Edit Supplier Modal

Use a modal (not a full page). Bank accounts are managed as inline rows within the modal.

**Form fields:**

| Field | Type | Required | Notes |
|-------|------|----------|-------|
| Name | Text input | Yes | Supplier name |
| Address | Textarea | Yes | Supplier address |
| Phone | Text input | No | Phone number |
| Email | Text input | No | Must be valid email format if provided |
| Website | Text input | No | URL format if provided |
| Active | Toggle | Yes | Default: true. Only shown in edit mode (new suppliers default to active). |

**Bank Accounts section** (inside the modal, below main fields):

```
Bank Accounts (optional)

| Account Name       | Account Number      | Actions |
|--------------------|---------------------|---------|
| [text input]       | [text input]        | [🗑]   |
| [text input]       | [text input]        | [🗑]   |
[+ Add Bank Account]
```

- Each row has: Account Name (text input) + Account Number (text input) + Remove button.
- Clicking **[+ Add Bank Account]** adds a new empty row.
- Remove button deletes the row. If only one row exists, removing it is allowed (bank accounts are optional).
- **Validation**: If a bank account row exists, both `accountName` and `accountNumber` are required. Show inline error on the incomplete field.

**Create mode:**
- Title: "Create Supplier"
- On save, show toast: "Supplier created successfully."

**Edit mode:**
- Title: "Edit Supplier"
- Pre-fill all fields including bank account rows.
- On save, show toast: "Supplier updated successfully."

### 1.4 Delete Supplier

Use the existing `ConfirmModal`:

| Property | Value |
|----------|-------|
| Title | Delete Supplier |
| Message | Are you sure you want to delete **{name}**? This action cannot be undone. |
| Cancel | Cancel |
| Confirm | Delete (danger variant) |

On confirm, remove from state and show toast: "Supplier {name} has been deleted."

> **Note**: If the supplier is referenced by products or purchase orders, show a warning in the confirmation message: "This supplier is referenced by {n} product(s) and {n} purchase order(s). Deleting it will remove the supplier association from those records." Proceed with deletion but clean up references (remove supplier ID from products' supplier arrays).

### 1.5 State Management

Create `useSupplierStore` with:

- `suppliers: Supplier[]`
- `addSupplier(supplier)` — auto-generates ID, createdAt, and bank account IDs.
- `updateSupplier(id, data)` — update supplier fields.
- `deleteSupplier(id)` — remove supplier.
- `getActiveSuppliers()` — returns suppliers where `active === true`.

### 1.6 Mock Data

Provide **4 predefined suppliers**:

| # | Name | Address | Phone | Email | Website | Bank Accounts | Active |
|---|------|---------|-------|-------|---------|---------------|--------|
| 1 | PT Sumber Makmur | Jl. Industri No. 45, Jakarta | +62-21-5550001 | order@sumbermakmur.co.id | sumbermakmur.co.id | BCA - 1234567890, Mandiri - 0987654321 | true |
| 2 | CV Jaya Abadi | Jl. Perdagangan No. 12, Surabaya | +62-31-5550002 | sales@jayaabadi.com | — | BCA - 1122334455 | true |
| 3 | UD Berkah Sentosa | Jl. Pasar Baru No. 8, Bandung | — | — | — | — | true |
| 4 | PT Global Supplies | Jl. Raya Serpong No. 100, Tangerang | +62-21-5550004 | info@globalsupplies.co.id | globalsupplies.co.id | BNI - 5566778899, BRI - 9988776655 | false |

---

## 2. Product Form: Supplier Field

Add a supplier selection field to the **top section** of the Product form (after Category, before Images).

**Field details:**

| Field | Type | Required | Notes |
|-------|------|----------|-------|
| Suppliers | MultiSelect | No | Select from active suppliers. Uses existing `MultiSelect` component. |

- Only active suppliers appear in the dropdown.
- A product can have zero or multiple suppliers.
- This is used later by Purchase Orders to filter products by supplier.

**Data model change** — add to `Product` interface:

```typescript
interface Product {
  // ... existing fields ...
  supplierIds: number[];  // array of supplier IDs, default: []
}
```

**Update mock products**: Assign suppliers to some existing products:
- T-Shirt → PT Sumber Makmur, CV Jaya Abadi
- Rice → UD Berkah Sentosa
- Notebook → CV Jaya Abadi
- Cooking Oil → (none)

---

## 3. Rack Master Data (`/master/rack`)

Add a new master data page for managing racks (physical storage locations in the store).

### 3.1 Rack Data Model

```typescript
interface Rack {
  id: number;
  name: string;
  code: string;           // unique code, e.g., "R-001", "A1-TOP"
  location: string;       // e.g., "Warehouse Zone A", "Store Front"
  capacity: number;       // generic number, user decides the unit
  description: string;    // optional
  active: boolean;        // default: true
  createdAt: string;
}
```

### 3.2 Rack List View

Table-based list following existing table design:

**Table columns:**

| Column | Sortable | Notes |
|--------|----------|-------|
| ID | Yes | Auto-increment |
| Name | Yes | Rack name |
| Code | Yes | Unique rack code |
| Location | Yes | Physical location |
| Capacity | No | Number value |
| Status | Yes | Badge: green for Active, gray for Inactive |
| Actions | No | Edit, Delete buttons |

**Features:**
- **Search**: text search across name, code, and location.
- **Sorting**: sortable on ID, Name, Code, Location, Status columns.
- **Pagination**: with items-per-page selector.
- **Add Rack** button at the top → opens modal form.

### 3.3 Create / Edit Rack Modal

**Form fields:**

| Field | Type | Required | Notes |
|-------|------|----------|-------|
| Name | Text input | Yes | Rack name |
| Code | Text input | Yes | Must be unique across all racks (case-insensitive) |
| Location | Text input | Yes | Physical location description |
| Capacity | Number input | Yes | Must be a positive number > 0 |
| Description | Textarea | No | Optional description |
| Active | Toggle | Yes | Default: true. Only shown in edit mode. |

**Validation:**
- Name is required.
- Code is required and must be unique (case-insensitive). Show inline error: "Rack code already exists."
- Location is required.
- Capacity is required and must be > 0.

**Create mode:**
- Title: "Create Rack"
- On save, show toast: "Rack created successfully."

**Edit mode:**
- Title: "Edit Rack"
- On save, show toast: "Rack updated successfully."

### 3.4 Delete Rack

Use `ConfirmModal`:

| Property | Value |
|----------|-------|
| Title | Delete Rack |
| Message | Are you sure you want to delete rack **{name}** ({code})? This action cannot be undone. |
| Cancel | Cancel |
| Confirm | Delete (danger variant) |

On confirm, remove from state, clean up rack references from variants, and show toast: "Rack {name} has been deleted."

### 3.5 State Management

Create `useRackStore` with:

- `racks: Rack[]`
- `addRack(rack)` — auto-generates ID and createdAt.
- `updateRack(id, data)` — update rack fields.
- `deleteRack(id)` — remove rack.
- `getActiveRacks()` — returns racks where `active === true`.

### 3.6 Mock Data

Provide **5 predefined racks**:

| # | Name | Code | Location | Capacity | Description | Active |
|---|------|------|----------|----------|-------------|--------|
| 1 | Main Display | R-001 | Store Front | 100 | Primary display shelf near entrance | true |
| 2 | Electronics Shelf | R-002 | Store Front | 50 | Dedicated electronics display | true |
| 3 | Cold Storage | R-003 | Warehouse Zone A | 200 | Refrigerated storage area | true |
| 4 | Bulk Storage | R-004 | Warehouse Zone B | 500 | Large item storage | true |
| 5 | Clearance Rack | R-005 | Store Back | 30 | Discounted items | false |

### 3.7 Sidebar Update

Add **Rack** to the sidebar under Master Data:

```
Master Data
  Product
  Category
  Supplier   ← already planned
  Rack       ← NEW
```

---

## 4. Variant: Rack Field

Add a rack assignment field to each variant (both simple and complex mode).

**Field details:**

| Field | Type | Required | Notes |
|-------|------|----------|-------|
| Racks | MultiSelect | No | Select from active racks. Indicates where this variant is stored/displayed. |

**Data model change** — add to `ProductVariant` interface:

```typescript
interface ProductVariant {
  // ... existing fields ...
  rackIds: number[];  // array of rack IDs, default: []
}
```

**UI placement in VariantsTab:**
- **Simple mode** (Has Variants = No): Show the Racks MultiSelect below the Barcode field.
- **Complex mode** (Has Variants = Yes): Show the Racks MultiSelect in the expanded variant detail section (alongside images and pricing).

**Update mock variants**: Assign racks to some existing product variants:
- T-Shirt variants → Main Display, Electronics Shelf
- Rice variants → Bulk Storage
- Notebook → Main Display
- Cooking Oil → Cold Storage

---

## 5. Pricing Simplification: Remove Price Type

Remove the `priceType` field (`retail` | `wholesale`) from variants. **All variants now use tiered pricing exclusively.**

### 5.1 What Changes

**Remove from `ProductVariant` interface:**
- `priceType` field
- `retailPrice` field
- `retailMarkup` field

**Keep:**
- `wholesaleTiers: PricingTier[]` — rename to `pricingTiers: PricingTier[]` for clarity (no longer specific to wholesale).

**Updated interface:**

```typescript
interface ProductVariant {
  id: string;
  sku: string;
  barcode: string;
  attributes: Record<string, string>;
  pricingTiers: PricingTier[];  // was wholesaleTiers
  images: string[];
  rackIds: number[];            // NEW from section 4
  currentStock: number;         // NEW, default: 0 (updated by PO receive)
}

interface PricingTier {
  minQty: number;
  value: number;  // sell price (fixed) or markup amount/percentage depending on product price setting
}
```

### 5.2 UI Changes

**VariantsTab:**
- Remove the "Price Type" radio selector from each variant.
- Always show tiered pricing editor (the current wholesale pricing UI).

**VariantPricing component:**
- Remove the retail/wholesale branching logic.
- Always show the tiered pricing table.
- The first tier (minQty = 1) cannot be removed.

**Add an info note** at the top of the pricing section in each variant:

> "All pricing uses tiered structure. For retail (single price), set one tier with Min Qty = 1. Add more tiers for volume/wholesale pricing."

### 5.3 Impact on existing behavior

The pricing rules still depend on the **product-level Price Setting** (fixed vs. markup) and **Markup Type** (percentage vs. fixed amount):

| Price Setting | Tier Column Label | Value Meaning |
|---------------|-------------------|---------------|
| Fixed Price | Sell Price | Exact sell price for this tier |
| Markup (Percentage) | Markup % | Percentage markup over cost price |
| Markup (Fixed Amount) | Markup Amount | Fixed amount added to cost price |

---

## 6. Product Form: Price Tab

Add a new **Price** tab as the first tab in the product form. Move pricing-related fields from the top section into this tab.

### 6.1 New Tab Order

```
[ Price ] [ Units ] [ Variants ]
```

(Previously: Units | Variants. Now Price is added at the beginning.)

### 6.2 Fields Moved to Price Tab

Remove these fields from the top section of the product form and place them in the Price tab:

- **Price Setting** (radio: Fixed Price / Markup Price)
- **Markup Type** (conditional radio: Percentage / Fixed Amount — only visible when Price Setting = Markup)

These remain at the top of the Price tab, above the product-level pricing table.

### 6.3 Product-Level Pricing Table

Below Price Setting and Markup Type, add a **product-level tiered pricing table**. This is optional and acts as a **default/template** that gets applied to all variants.

```
┌─────────────────────────────────────────────────────────────────┐
│  Price Tab                                                       │
│                                                                   │
│  Price Setting:   ○ Fixed Price  ○ Markup Price                  │
│  Markup Type:     ○ Percentage   ○ Fixed Amount  (if markup)     │
│                                                                   │
│  ─────────────────────────────────────────────────────────────── │
│                                                                   │
│  Default Pricing (optional)                            [Edit]    │
│                                                                   │
│  ⓘ Pricing set here will be applied to ALL variants. Variants   │
│    can override these values individually in the Variants tab.   │
│    However, saving these values will REPLACE all variant          │
│    pricing, including any overrides.                              │
│                                                                   │
│  | Min Qty | Sell Price | Actions |  ← label changes by setting │
│  |---------|------------|---------|                               │
│  | 1       | [input]    | —       |  ← first tier, can't remove │
│  | 10      | [input]    | [🗑]   |                               │
│  | 100     | [input]    | [🗑]   |                               │
│  [+ Add Tier]                                                    │
│                                                                   │
│                                         [Cancel] [Save Pricing]  │
│                                                                   │
└─────────────────────────────────────────────────────────────────┘
```

**Column labels** adapt to the Price Setting:

| Price Setting | Column 2 Label |
|---------------|----------------|
| Fixed Price | Sell Price |
| Markup (Percentage) | Markup % |
| Markup (Fixed Amount) | Markup Amount |

**Behavior:**
- This pricing table is **optional**. If left empty, each variant must define its own pricing.
- If the user fills in tiers here, those tiers are **copied** to all variants' `pricingTiers` arrays.
- The product-level pricing is stored temporarily in the form state and used as a template — it is NOT persisted separately on the product model. It only serves to populate variant pricing.

### 6.4 Edit / Save / Broadcast Behavior

The product-level pricing table starts in **read-only mode** (inputs disabled). The user must explicitly click **[Edit]** to enable editing, and then **[Save Pricing]** to apply changes. This prevents accidental edits and avoids disruptive confirmation modals while the user is still typing.

**Flow:**

1. **Initial state**: Pricing table is read-only. An **[Edit]** button is shown at the top-right of the "Default Pricing" section.
2. **Edit mode**: Clicking [Edit] enables all pricing inputs. The [Edit] button is replaced with **[Cancel]** and **[Save Pricing]** buttons at the bottom of the pricing table.
3. **Cancel**: Reverts all changes made during this edit session. Returns to read-only mode.
4. **Save Pricing**: When the user clicks [Save Pricing] and variants already have pricing data:
   - Show a **confirmation modal**:
     - Title: "Update All Variant Pricing"
     - Message: "This will replace pricing on all variants, including any custom values. Are you sure you want to continue?"
     - Cancel / Update All
   - If confirmed, copy the product-level tiers to every variant's `pricingTiers`. Return to read-only mode.
   - If cancelled, stay in edit mode (do not revert — let the user continue editing).
5. **Save Pricing (no existing variant pricing)**: If no variants have pricing data yet, skip the confirmation modal and apply directly.

> **Backend note (for future implementation)**: When the backend is built, consider storing product-level default pricing separately and resolving the effective price at query time (product default → variant override). For now (frontend-only), we flatten it by copying to variants.

### 6.5 Data Model Note

The product-level pricing is NOT a new field on the `Product` interface. It's a **form-only** concept used as a convenience tool to bulk-set variant pricing. When the product is saved, only the individual variant `pricingTiers` are persisted.

---

## 7. Variant: Stock Field

Add a `currentStock` field to each variant. This is updated when a Purchase Order is received (section 8).

```typescript
interface ProductVariant {
  // ... existing fields ...
  currentStock: number;  // default: 0
}
```

- This field is **read-only** in the product form (displayed but not editable by the user).
- Display it in the variant table/detail as "Stock: {n}" with a subtle label.
- Stock is modified only through the Purchase Order receive flow.
- For now, all mock product variants have `currentStock: 0`.

---

## 8. Purchase Order (`/transaction/purchase`)

Full workflow for creating, sending, receiving, and completing purchase orders.

### 8.1 Purchase Order Data Model

```typescript
interface PurchaseOrderItem {
  id: string;             // auto-generated UUID
  productId: number;
  productName: string;    // denormalized for display
  variantId: string;
  variantLabel: string;   // denormalized, e.g., "Red / L" or "Default"
  sku: string;            // denormalized for display
  unitId: string;         // selected unit ID from the product's units
  unitName: string;       // denormalized unit label, e.g., "pcs", "kg", "box"
  currentStock: number;   // snapshot at time of PO creation (read-only)
  orderedQty: number;     // how many to order
  price: number;          // unit price (latest from supplier, or 0 if unknown)
  // Receive fields (filled during receive flow)
  receivedQty?: number;
  receivedPrice?: number; // if different from ordered price
  isVerified?: boolean;   // user confirmed qty matches
}

interface PurchaseOrder {
  id: number;
  poNumber: string;            // auto-generated, format: "PO-YYYY-NNNN"
  supplierId: number;
  supplierName: string;        // denormalized for display
  date: string;                // ISO date, PO creation date
  status: 'draft' | 'sent' | 'received' | 'completed' | 'cancelled';
  items: PurchaseOrderItem[];
  notes: string;               // optional, internal notes

  // Receive data (filled when going to "received" status)
  receivedDate?: string;       // ISO datetime
  paymentMethod?: 'cash' | 'credit_card' | 'bank_transfer';
  supplierBankAccountId?: string;  // selected from supplier's bank accounts
  subtotal?: number;           // sum of (receivedQty × receivedPrice) for all items
  totalItems?: number;         // sum of all receivedQty

  createdAt: string;
  updatedAt: string;
}
```

### 8.2 PO Number Generation

Auto-generate PO numbers in the format `PO-YYYY-NNNN`:
- `YYYY` = current year (e.g., 2026)
- `NNNN` = sequential number, zero-padded (e.g., 0001, 0002)
- Example: `PO-2026-0001`, `PO-2026-0002`
- The sequence resets each year (future backend concern — for now, just auto-increment).

### 8.3 Status Workflow

```
              ┌──────────┐
              │  Draft    │──────────┐
              └────┬─────┘          │
                   │ Mark as Sent   │ Cancel
                   ▼                ▼
              ┌──────────┐    ┌───────────┐
              │  Sent     │───→│ Cancelled │
              └────┬─────┘    └───────────┘
                   │ Receive
                   ▼
              ┌──────────┐
              │ Received  │
              └────┬─────┘
                   │ Complete
                   ▼
              ┌──────────┐
              │ Completed │
              └──────────┘
```

**Rules:**
- Status flow is **one-directional** — no going backward.
- **Draft** → can be edited, marked as sent, or cancelled.
- **Sent** → can be received or cancelled. (Future: triggers email/whatsapp to supplier.)
- **Received** → goods verified, stock updated. Can be marked as completed.
- **Completed** → final state. View-only.
- **Cancelled** → final state. View-only. Can only be reached from draft or sent.

**Status badge colors:**

| Status | Color |
|--------|-------|
| Draft | gray |
| Sent | blue |
| Received | yellow/amber |
| Completed | green |
| Cancelled | red |

### 8.4 Routing Structure

```
/transaction/purchase                  - PO list page (simplified cards)
/transaction/purchase/add              - Create new PO (full page form)
/transaction/purchase/[id]             - PO detail view (full page)
/transaction/purchase/[id]/edit        - Edit PO (only available for draft)
/transaction/purchase/[id]/receive     - Receive PO (separate full page)
```

### 8.5 PO List Page (Simplified / Card-Based)

Instead of a full data table, use a **card-based list** for a more visual PO overview.

**Page layout:**

```
┌────────────────────────────────────────────────────────────────┐
│  Purchase Orders                               [+ New Order]   │
│                                                                 │
│  [All] [Draft] [Sent] [Received] [Completed] [Cancelled]      │
│                                                                 │
│  🔍 Search by PO number or supplier...                         │
│                                                                 │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │ PO-2026-0001                          [Draft]            │  │
│  │ PT Sumber Makmur                                         │  │
│  │ 12 Feb 2026 · 5 items · Rp 2,500,000                    │  │
│  │                                    [View] [Edit] [Delete]│  │
│  └──────────────────────────────────────────────────────────┘  │
│                                                                 │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │ PO-2026-0002                          [Sent]             │  │
│  │ CV Jaya Abadi                                            │  │
│  │ 10 Feb 2026 · 3 items · Rp 750,000                      │  │
│  │                                    [View] [Receive]      │  │
│  └──────────────────────────────────────────────────────────┘  │
│                                                                 │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │ PO-2026-0003                          [Completed]        │  │
│  │ UD Berkah Sentosa                                        │  │
│  │ 5 Feb 2026 · 8 items · Rp 4,200,000                     │  │
│  │                                              [View]      │  │
│  └──────────────────────────────────────────────────────────┘  │
│                                                                 │
│  Showing 1-10 of 25 orders           [< Prev] [Next >]        │
└────────────────────────────────────────────────────────────────┘
```

**Card contents:**
- PO number (bold, top-left)
- Status badge (top-right)
- Supplier name
- Date · Item count · Total price (formatted as currency)
- Action buttons (bottom-right, based on status)

**Status filter tabs**: Filter cards by status. "All" shows everything. Tabs show count: "Draft (3)", "Sent (2)", etc.

**Search**: Filter by PO number or supplier name.

**Sorting**: Cards sorted by date (newest first).

**Pagination**: Simple prev/next with item count ("Showing 1-10 of 25 orders").

**Action buttons per status:**

| Status | Available Actions |
|--------|-------------------|
| Draft | View, Edit, Delete |
| Sent | View, Receive |
| Received | View, Complete |
| Completed | View |
| Cancelled | View |

- **Delete** is only available for draft POs. Show `ConfirmModal`.
- **Cancel** action: available from the detail page (not the card) for draft and sent POs.

### 8.6 PO Create / Edit Page (`/transaction/purchase/add`, `/transaction/purchase/[id]/edit`)

Full page form for creating or editing a purchase order. Only available for **draft** POs.

**Page layout:**

```
┌─────────────────────────────────────────────────────────────────┐
│  ← Back to Purchase Orders                     [Save] [Cancel]  │
├─────────────────────────────────────────────────────────────────┤
│                                                                   │
│  New Purchase Order  (or "Edit Purchase Order — PO-2026-0001")   │
│                                                                   │
│  Supplier:    [Select supplier ▼]  (required, active only)       │
│  Date:        [📅 12/02/2026    ]  (default: today)              │
│  Notes:       [________________________]  (optional, textarea)   │
│                                                                   │
├─────────────────────────────────────────────────────────────────┤
│                                                                   │
│  Order Items                                                     │
│                                                                   │
│  ⓘ Showing products linked to the selected supplier and         │
│    products without supplier assignment.                          │
│                                                                   │
│  🔍 Search products...                   [+ Add Product]         │
│                                                                   │
│  ▼ T-Shirt (Clothing)                                            │
│  ┌─────────────────────────────────────────────────────────────┐ │
│  │ Variant    │ SKU    │ Stock │ Unit      │ Order Qty │ Price │ │
│  │────────────│────────│───────│───────────│───────────│───────│ │
│  │ Red / S    │ TS-R-S │ 0     │ [pcs  ▼]  │ [  0  ]   │ 0    │ │
│  │ Red / M    │ TS-R-M │ 0     │ [pcs  ▼]  │ [  0  ]   │ 0    │ │
│  │ Red / L    │ TS-R-L │ 0     │ [pcs  ▼]  │ [  0  ]   │ 0    │ │
│  │ Blue / S   │ TS-B-S │ 0     │ [pcs  ▼]  │ [  0  ]   │ 0    │ │
│  └─────────────────────────────────────────────────────────────┘ │
│                                                         [Remove] │
│                                                                   │
│  ▼ Rice (Food & Beverages)                                       │
│  ┌─────────────────────────────────────────────────────────────┐ │
│  │ Variant    │ SKU    │ Stock │ Unit      │ Order Qty │ Price │ │
│  │────────────│────────│───────│───────────│───────────│───────│ │
│  │ Default    │ RC-001 │ 0     │ [kg   ▼]  │ [  0  ]   │ 0    │ │
│  └─────────────────────────────────────────────────────────────┘ │
│                                                         [Remove] │
│                                                                   │
│  ─────────────────────────────────────────────────────────────── │
│  Summary: 6 variants · Total Qty: 0 · Est. Total: Rp 0          │
│                                                                   │
└─────────────────────────────────────────────────────────────────┘
```

**Product selection behavior when a supplier is chosen:**

1. Auto-populate the item list with products that have the selected supplier in their `supplierIds`.
2. Also include products with **empty** `supplierIds` (no supplier assigned) — these are available to any PO.
3. The user can manually **[+ Add Product]** to add any other product via a search/select dialog, regardless of supplier assignment. This gives flexibility for one-off orders.
4. The user can **[Remove]** a product group from the PO.
5. If the user changes the supplier, show a confirmation modal: "Changing the supplier will reset the product list. Continue?" If confirmed, clear items and repopulate with the new supplier's products.

**Per-variant fields:**

| Field | Editable | Notes |
|-------|----------|-------|
| Variant | No | Variant label (attribute combination or "Default") |
| SKU | No | From variant data |
| Stock | No | `currentStock` from variant (read-only) |
| Unit | Yes | Dropdown populated from the **product's units** (base unit + conversion units). Defaults to the product's base unit. User can select a different unit for ordering (e.g., order in "box" instead of "pcs"). |
| Order Qty | Yes | Number input, default: 0. User enters how many to order in the selected unit. |
| Price | No | Latest purchase price from this supplier. Set to 0 for now (no purchase history). Will be populated from previous POs in the future. |

> **Backend note**: In the future, the Price column will show the latest cost price from the most recent completed PO for this supplier + variant + unit combination. Price calculation based on the selected unit (e.g., if a "box" contains 12 pcs, the price should reflect the box price) will also be handled by the backend. For now (mock), always show 0 regardless of the selected unit.

**Summary bar** at the bottom:
- Total variant count (variants with Order Qty > 0)
- Total quantity (sum of all Order Qty)
- Estimated total price (sum of Order Qty × Price for all items)

**Validation on save:**
- Supplier is required.
- Date is required.
- At least one item must have Order Qty > 0.

### 8.7 PO Detail Page (`/transaction/purchase/[id]`)

Read-only view of a purchase order with status-appropriate actions.

**Page layout:**

```
┌─────────────────────────────────────────────────────────────────┐
│  ← Back to Purchase Orders                                      │
├─────────────────────────────────────────────────────────────────┤
│                                                                   │
│  PO-2026-0001                                    [Draft]         │
│                                                                   │
│  Supplier:  PT Sumber Makmur                                     │
│  Date:      12 Feb 2026                                          │
│  Notes:     First order for new collection                       │
│                                                                   │
│  ─────────────────────────────────────────────────────────────── │
│                                                                   │
│  Order Items                                                     │
│                                                                   │
│  ▼ T-Shirt                                                       │
│  │ Red / S  │ SKU: TS-R-S │ Unit: pcs │ Ordered: 50  │ Rp 0    │
│  │ Red / M  │ SKU: TS-R-M │ Unit: pcs │ Ordered: 100 │ Rp 0    │
│  │ Red / L  │ SKU: TS-R-L │ Unit: pcs │ Ordered: 75  │ Rp 0    │
│                                                                   │
│  ─────────────────────────────────────────────────────────────── │
│  Total: 225 items · Rp 0                                         │
│                                                                   │
│  ── Receive Information (only shown when status ≥ received) ──  │
│  Received: 12 Feb 2026, 14:30                                    │
│  Payment:  Bank Transfer → BCA - 1234567890                      │
│  Total Received: 220 items · Rp 18,500,000                       │
│                                                                   │
├─────────────────────────────────────────────────────────────────┤
│  Actions:                                                        │
│  [Edit] [Mark as Sent] [Cancel Order]       ← for Draft         │
│  [Receive] [Cancel Order]                   ← for Sent          │
│  [Mark as Completed]                        ← for Received      │
│  (no actions)                               ← for Completed     │
│  (no actions)                               ← for Cancelled     │
└─────────────────────────────────────────────────────────────────┘
```

**Action behavior:**

| Action | Behavior |
|--------|----------|
| Edit | Navigate to `/transaction/purchase/[id]/edit` (draft only) |
| Mark as Sent | Confirmation modal → update status to `sent`, show toast. Message: "Mark this PO as sent to the supplier?" |
| Receive | Navigate to `/transaction/purchase/[id]/receive` |
| Mark as Completed | Confirmation modal → update status to `completed`, show toast. Message: "Mark this PO as completed? This action cannot be undone." |
| Cancel Order | Confirmation modal (danger) → update status to `cancelled`, show toast. Message: "Are you sure you want to cancel this purchase order? This action cannot be undone." |

> **Future note**: "Mark as Sent" will trigger email/WhatsApp notification to the supplier. For now, it only updates the status.

### 8.8 PO Receive Page (`/transaction/purchase/[id]/receive`)

A dedicated full page for verifying received goods. Only accessible when PO status is `sent`.

**Page layout:**

```
┌─────────────────────────────────────────────────────────────────┐
│  ← Back to PO-2026-0001                            [Save]      │
├─────────────────────────────────────────────────────────────────┤
│                                                                   │
│  Receive — PO-2026-0001                                          │
│  Supplier: PT Sumber Makmur                                      │
│                                                                   │
│  Received Date:  [📅 13/02/2026 14:30]  (default: now)          │
│  Payment Method: [Cash           ▼]                              │
│  Bank Account:   [Select account ▼]   ← only if not cash        │
│                                                                   │
├─────────────────────────────────────────────────────────────────┤
│                                                                   │
│  ⚠️ Received quantity matches ordered quantity.                  │
│     ☐ I understand, don't show this message again.              │
│                                                                   │
│  ▼ T-Shirt                                                       │
│  ┌───────────────────────────────────────────────────────────┐  │
│  │       │ Unit │ Ordered │ Received │ Price   │ Status      │  │
│  │───────│──────│─────────│──────────│─────────│─────────────│  │
│  │ Red/S │ pcs  │ 50      │ [50  ]   │ [  0  ] │ ☑ OK        │  │
│  │ Red/M │ pcs  │ 100     │ [95  ]   │ [  0  ] │ ⚠ Mismatch  │  │
│  │ Red/L │ pcs  │ 75      │ [75  ]   │ [8500 ] │ ☑ OK        │  │
│  └───────────────────────────────────────────────────────────┘  │
│                                                                   │
│  ─────────────────────────────────────────────────────────────  │
│                                                                   │
│  Subtotal:       Rp 18,500,000                                   │
│  Total Items:    220                                              │
│  Total Price:    Rp 18,500,000                                   │
│                                                                   │
│                                                  [Save Receive]  │
└─────────────────────────────────────────────────────────────────┘
```

**Top section:**
- **Received Date**: Date + time picker, defaults to current date/time.
- **Payment Method**: Dropdown with options: Cash, Credit Card, Bank Transfer. Default: Cash.
- **Bank Account**: Only shown when payment method is NOT cash. Dropdown populated from the **supplier's bank accounts** (from supplier master data). Required when visible.

**Item verification table:**

For each variant in the PO:

| Column | Type | Notes |
|--------|------|-------|
| Variant | Read-only | Variant label |
| Unit | Read-only | The unit selected during ordering (e.g., "pcs", "kg", "box") |
| Ordered | Read-only | Original ordered quantity in the selected unit |
| Received | Input (number) | Default: same as ordered qty. User can adjust. |
| Price | Input (number) | Default: ordered price (or 0). User can input actual purchase price. |
| Status | Auto-calculated | See below |

**Status column behavior:**

- If `receivedQty === orderedQty` AND price unchanged: Show **checkbox** (☑ OK). The checkbox auto-checks when quantities match. User can uncheck to manually edit.
- If `receivedQty !== orderedQty` OR price changed: Show **"⚠ Mismatch"** label. No checkbox — user must manually verify.

**"Don't show again" warning:**

When ALL items have matching quantities (receivedQty === orderedQty), show a **dismissible info banner** at the top:

> "⚠️ Received quantity matches ordered quantity. ☐ I understand, don't show this message again."

- The "don't show again" preference is stored in **localStorage** (`po_receive_match_warning_dismissed`).
- Once dismissed, the banner never shows again (across all POs).
- When the banner is dismissed or checkbox is checked, the input fields for matching items become **disabled** (read-only) since the user confirmed they're correct. User can still click the ☑ checkbox to uncheck and re-enable editing.

**Summary section** (bottom):
- **Subtotal**: Sum of `receivedQty × receivedPrice` for all items.
- **Total Items**: Sum of all `receivedQty`.
- **Total Price**: Same as subtotal (no tax/discount in this phase).

These totals **update in real-time** as the user modifies received quantities and prices.

**Save behavior:**

When the user clicks **[Save Receive]**:

1. **Validation**:
   - Received date is required.
   - Payment method is required.
   - If payment method is not cash, bank account is required.
   - All received quantities must be ≥ 0.
   - All prices must be ≥ 0.

2. **On save**:
   - Update PO status to `received`.
   - Store received date, payment method, bank account, subtotal, total items.
   - Store per-item `receivedQty` and `receivedPrice`.
   - **Update variant stock**: For each item, add `receivedQty` to the variant's `currentStock`.
   - Show toast: "Purchase order received successfully. Stock has been updated."
   - Navigate back to PO detail page.

> **Important**: Stock update happens at receive time, not at completion time. Completion is an administrative/bookkeeping finalization step.

### 8.9 State Management

Create `usePurchaseOrderStore` with:

- `purchaseOrders: PurchaseOrder[]`
- `addPurchaseOrder(po)` — auto-generates ID, poNumber, createdAt, updatedAt.
- `updatePurchaseOrder(id, data)` — update PO fields and updatedAt.
- `deletePurchaseOrder(id)` — remove PO (only draft status).
- `updateStatus(id, status)` — transition PO status (validates one-directional flow).
- `receivePurchaseOrder(id, receiveData)` — stores receive data, updates status to `received`, and **updates variant stock** in `useProductStore`.
- `completePurchaseOrder(id)` — update status to `completed`.
- `cancelPurchaseOrder(id)` — update status to `cancelled` (only from draft/sent).
- `getNextPoNumber()` — generates next PO number in sequence.

### 8.10 Mock Data

Provide **4 predefined purchase orders** in different statuses:

| # | PO Number | Supplier | Date | Status | Items | Notes |
|---|-----------|----------|------|--------|-------|-------|
| 1 | PO-2026-0001 | PT Sumber Makmur | 2026-02-05 | completed | T-Shirt (3 variants), Notebook (1 variant) | First restocking order |
| 2 | PO-2026-0002 | CV Jaya Abadi | 2026-02-08 | received | Notebook (1 variant) | Urgent notebook restock |
| 3 | PO-2026-0003 | UD Berkah Sentosa | 2026-02-10 | sent | Rice (1 variant) | Monthly rice order |
| 4 | PO-2026-0004 | PT Sumber Makmur | 2026-02-12 | draft | T-Shirt (2 variants) | Pending review |

For completed/received POs, include realistic receive data (receivedQty, receivedPrice, payment method, etc.) and update the corresponding variant `currentStock` values in mock product data to reflect the received quantities.

---

## 9. New UI Components

Build or extend these components as needed:

| Component | Purpose | Notes |
|-----------|---------|-------|
| **DatePicker** | Date and date-time input | For PO date and received date. Can use a native `<input type="date">` / `<input type="datetime-local">` styled to match the design system, or build a custom calendar picker. |
| **StatusBadge** | Colored status label | Reusable component for PO status, supplier status, etc. Takes `status` and `colorMap` props. Consider extracting from existing inline badge implementations. |
| **Card** | Card container for PO list | Extend existing Card component if it exists, or create a styled container with header, body, and footer sections. |
| **SearchInput** | Search input with icon | If not already a standalone component, extract from existing pages for reuse. |

---

## 10. Validation Summary

### Supplier-level
- Name is required.
- Address is required.
- Email must be valid format if provided.
- Bank accounts: if a row exists, both account name and account number are required.

### Rack-level
- Name is required.
- Code is required and must be unique (case-insensitive).
- Location is required.
- Capacity is required and must be > 0.

### Product-level (updated)
- Name is required.
- Category is required.
- At least one unit (base unit) must be defined.
- Price Setting must be selected (now in Price tab).

### Variant-level (updated)
- SKU must be unique across all variants in this product (if filled).
- Barcode must be unique across all variants in this product (if filled).
- Pricing tiers: first tier must have Min Qty = 1, tiers must be in ascending qty order.
- Price/markup value warnings for non-descending patterns (warn, don't block).

### Purchase Order-level
- Supplier is required.
- Date is required.
- At least one item must have Order Qty > 0.
- On receive: received date required, payment method required, bank account required if non-cash payment.

---

## 11. Confirmation Dialogs

All confirmation dialogs use the existing `ConfirmModal` component.

| Trigger | Title | Message | Cancel | Confirm |
|---------|-------|---------|--------|---------|
| Delete supplier | Delete Supplier | Are you sure you want to delete **{name}**? This action cannot be undone. | Cancel | Delete (danger) |
| Delete rack | Delete Rack | Are you sure you want to delete rack **{name}** ({code})? This action cannot be undone. | Cancel | Delete (danger) |
| Save product-level pricing via [Save Pricing] button (variants exist) | Update All Variant Pricing | This will replace pricing on all variants, including any custom values. Are you sure you want to continue? | Cancel | Update All |
| Change Price Setting (existing pricing data) | Change Price Setting | Changing price setting will reset all pricing data. Are you sure you want to continue? | Cancel | Continue |
| Change supplier on PO (items exist) | Change Supplier | Changing the supplier will reset the product list. Are you sure you want to continue? | Cancel | Continue |
| Delete PO (draft only) | Delete Purchase Order | Are you sure you want to delete **{poNumber}**? This action cannot be undone. | Cancel | Delete (danger) |
| Mark PO as sent | Send Purchase Order | Mark this PO as sent to the supplier? | Cancel | Send |
| Mark PO as completed | Complete Purchase Order | Mark this PO as completed? This action cannot be undone. | Cancel | Complete |
| Cancel PO | Cancel Purchase Order | Are you sure you want to cancel **{poNumber}**? This action cannot be undone. | Cancel | Cancel Order (danger) |

---

## 12. Suggestions & Notes

### Suggestion: Allow Draft → Received (Skip "Sent")

For walk-in purchases or phone orders where goods arrive immediately, consider adding a **"Receive Directly"** button on draft POs that skips the "sent" status. This is common in small retail operations. The flow would be: Draft → Received (directly). This is optional — implement only if desired.

### Suggestion: Partial Receiving

The current design assumes a single receive event per PO. In practice, suppliers sometimes ship in multiple batches. Consider supporting multiple receive events per PO in a future phase. For now, a single receive is sufficient.

### Suggestion: PO Print / Export

Consider adding a "Print" or "Export PDF" button on the PO detail page in a future phase. This is useful for sending POs to suppliers who don't use email.

### Backend Notes for Future Implementation

- **Supplier email/WhatsApp**: "Mark as Sent" should trigger a notification to the supplier with PO details.
- **Purchase price history**: Track cost prices from completed POs per supplier + variant. Use the latest price as the default in new POs.
- **Forecast minimum quantity**: Add a `minStockForecast` field to variants. Use it to auto-suggest order quantities in new POs (default order qty = minStockForecast - currentStock).
- **Stock ledger**: Instead of directly modifying `currentStock`, create stock movement records (type: "purchase_receive", qty, date, poId) for audit trail.
- **Multi-currency**: If dealing with international suppliers, consider adding currency to supplier and PO models.

---

## 13. General Notes

- Follow existing page design, layout, and styling patterns established in previous phases.
- Show toast notifications on all create/update/delete/status-change actions.
- Supplier and Rack forms use **modals** (consistent with Category and User).
- PO pages (create, detail, receive) use **full pages** (consistent with Product form pattern — sticky header with Save/Cancel, back link).
- All data persisted in Zustand stores with mock initial data loaded on app start.
- The PO receive flow is the only mechanism to update variant `currentStock` in this phase.
- Reuse existing components (`MultiSelect`, `Table`, `Tabs`, `ConfirmModal`, `Toggle`, `ImageUpload`, etc.) wherever possible.
- The "Price" tab in the product form uses the same `VariantPricing` component internally (tiered pricing table), just with a different context label.
