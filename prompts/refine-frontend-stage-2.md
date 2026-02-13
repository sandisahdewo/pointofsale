# Frontend App - Admin Panel (Phase 2)

## Overview

Phase 2 adds table enhancements (sorting, items-per-page) to existing pages and introduces the Master Product page — a complex form with units, variants, and pricing.

> **Backend note**: All data is client-side (mock/Zustand) for now. Design sorting, pagination, and filtering so they can be replaced with backend API calls in the future without changing the UI.

---

## 1. Table Sorting

Enhance the existing `Table` component to support column sorting.

### Requirements

- Add a `sortable` boolean property to the `Column` interface. Only columns with `sortable: true` show sort controls.
- Sortable column headers display a sort icon (ascending ▲, descending ▼, or neutral ⇅ when unsorted).
- Clicking a sortable column header cycles: **unsorted → ascending → descending → unsorted**.
- Only one column can be sorted at a time. Clicking a different column resets the previous one.
- The Table component exposes `onSort(key: string, direction: 'asc' | 'desc' | null)` callback — the parent page handles the actual sorting logic. This keeps the Table component backend-ready (the parent can sort client-side now, server-side later).
- Sorting resets the current page to 1.
- The Actions column must NOT be sortable.

### Apply to Existing Pages

- **Master Category**: make ID, Name, and Description columns sortable.

---

## 2. Items Per Page (Pagination Enhancement)

Enhance the existing `Table` pagination to allow changing items per page.

### Requirements

- Add a dropdown/select above or next to the pagination controls with options: **5, 10, 25, 50, 100**.
- Default: **10** items per page (change the current hardcoded `PAGE_SIZE = 5`).
- The Table component exposes `onPageSizeChange(size: number)` callback — the parent page handles re-slicing the data.
- Display total item count: "Showing 1-10 of 47 items".
- Changing items per page resets the current page to 1.

### Apply to Existing Pages

- **Master Category**: integrate the new items-per-page selector.

---

## 3. Master Product Page (`/master/product`)

Full CRUD page for managing products. This is the most complex page in the system.

### 3.1 Product List View

Similar to Master Category:
- **Table columns**: ID, Image (thumbnail of first product image), Name, Category, Status, Actions
- **Search**: text search across product name
- **Sorting**: sortable on ID, Name, Category columns
- **Pagination**: with items-per-page selector (same as enhanced Table)
- **Add Product** button → navigates to product form page (not a modal — the form is too complex for a modal)
- **Actions per row**: Edit (navigate to form page), Delete (confirmation modal)

### 3.2 Product Form Page (`/master/product/add` and `/master/product/edit/[id]`)

The product form is a **full page** (not a modal) inside AdminLayout. It has two sections:
1. **Top section**: General product fields
2. **Bottom section**: Tabbed area with **Units** and **Variants** tabs

#### 3.2.1 General Product Fields (Top Section)

| Field | Type | Required | Notes |
|-------|------|----------|-------|
| Name | Text input | Yes | Product name |
| Description | Textarea | No | Product description |
| Category | Dropdown/Select | Yes | Select from existing categories (use category store data) |
| Images | Multi-image upload | No | Multiple images with drag-to-reorder. Upload UI only (store as mock data/base64 for now). First image = primary/thumbnail. |
| Price Setting | Radio/Toggle | Yes | **Fixed Price** or **Markup Price** |
| Markup Type | Radio/Toggle | Conditional | Only shown when Price Setting = Markup. Options: **Percentage** or **Fixed Amount** |
| Has Variants | Radio/Toggle | Yes | **No** (default) or **Yes**. Controls which form is shown in the Variants tab. Switching this shows a confirmation modal if variant data already exists: title "Reset Variant Data", message "Changing this will reset variant data. Are you sure you want to continue?", with Cancel and Continue buttons. |
| Status | Toggle/Switch | Yes | Active / Inactive. Default: Active |

#### 3.2.2 Units Tab

Units are **per-product** (not a separate master data entity). Each product defines its own unit structure.

**Data model per unit row:**
- `name`: string — unit name (e.g., "Box", "Dozen", "Pcs")
- `conversionFactor`: number — how many of the referenced unit this equals
- `convertsTo`: string — reference to another unit in this product (by name or id)
- `toBaseUnit`: number — auto-calculated total conversion to base unit

**Base unit rules:**
- The **first unit added** is always the **base unit** (smallest unit).
- The base unit has no conversion (it IS the base, factor = 1).
- The base unit **cannot be deleted**.
- The base unit row shows a "Base Unit" badge/label.
- The base unit **name is editable** — it is not hardcoded. The user types the name when creating it (e.g., "Pcs", "Kg", "Liter", "Meter", "Sheet") and can rename it later via an inline edit (pencil icon or click-to-edit on the name cell). Renaming the base unit automatically updates all conversion display strings that reference it (e.g., "1 = 12 × Pcs" becomes "1 = 12 × Kg").
- All other units must ultimately reference back to the base unit (directly or through a chain).

**Adding the base unit:**
- When no units exist yet, the Units tab shows a prompt: **"Define base unit"** with a single text input for the unit name and an **[Add Base Unit]** button.
- This is a simplified flow — no conversion fields needed since the base unit always has factor = 1.

**Adding subsequent units:**
- User clicks **[+ Add Unit]**
- Form/inline row appears with:
  - Unit name: text input
  - Conversion: `1 [unit name] = [number input] × [dropdown of existing units]`
- The dropdown shows all existing units for this product.
- After adding, the table shows the auto-calculated "= X base unit" value.

**Display as a table:**

```
| Unit   | Conversion          | = Base Unit  | Actions     |
|--------|---------------------|--------------|-------------|
| Pcs    | Base Unit           | 1            | [✏️]        |
| Dozen  | 1 = 12 × Pcs       | 12           | [✏️] [🗑]  |
| Box    | 1 = 12 × Dozen     | 144          | [✏️] [🗑]  |
| Bag    | 1 = 50 × Pcs       | 50           | [✏️] [🗑]  |
| Karung | 1 = 2 × Box        | 288          | [✏️] [🗑]  |
```

- The ✏️ (edit) action on the base unit row only allows renaming.
- The ✏️ (edit) action on other units allows renaming and changing the conversion factor / reference unit.

**Supports both linear chains and branching** — a unit can reference any existing unit, not just the previous one. This allows structures like:
- Linear: Box → Dozen → Pcs
- Branching: Both "Box = 12 Dozen" and "Bag = 50 Pcs" referencing different points in the tree

**Validation:**
- Unit name must be unique within the product.
- Conversion factor must be a positive number greater than 0.
- A unit cannot reference itself (no circular reference).
- Prevent circular references: if A → B → C, then C cannot reference A.

**Stock locking (future-ready):**
- When stock exists (to be enforced by backend later), the entire unit table becomes **read-only**. Show a lock icon with tooltip: "Units cannot be modified while stock exists."
- For now (no backend), build the UI with the lock mechanism but leave it always unlocked.

**Product must have at least one unit (the base unit) before it can be saved.**

#### 3.2.3 Variants Tab

The Variants tab content changes based on the **Has Variants** toggle in the general fields above.

> **Backend schema note**: Both modes use the same underlying data structure — a product always has a `variants[]` array. "No variants" simply means the array has exactly one entry. This ensures a single database schema works for both cases when the backend is built later.

---

**Mode A: Has Variants = No (Simple Form)**

When the user selects "No" for Has Variants, show a **single-row form** directly — no attributes, no combinations, just one set of fields:

```
┌─────────────────────────────────────────────────────────────┐
│  SKU:         [________________________]                    │
│  Barcode:     [________________________]                    │
│  Price Type:  ○ Retail  ○ Wholesale                         │
│                                                             │
│  (Pricing fields based on Price Setting + Price Type,       │
│   see section 3.2.4)                                        │
│                                                             │
│  Images:      [📷 Drop images here / Browse]  (optional)   │
│               [img1] [img2]                                 │
└─────────────────────────────────────────────────────────────┘
```

This is the default mode. Internally this creates a single variant with no attribute values.

---

**Mode B: Has Variants = Yes (Full Variant Form)**

When the user selects "Yes" for Has Variants, show the full variant management system:

**Step 1: Define Variant Attributes**

User adds attribute types and their values directly on the product form:

```
[+ Add Attribute]

| Attribute Name | Values                      | Actions |
|----------------|-----------------------------|---------|
| Color          | Red, Blue, Green  [+ Add]   | [🗑]   |
| Size           | S, M, L, XL      [+ Add]   | [🗑]   |
```

- Attribute name: text input
- Values: tag-style input (type and press Enter to add, click × to remove)
- User can add multiple attributes (Color, Size, Material, etc.)

**Step 2: Auto-Generate Variant Combinations**

After defining attributes and values, click **[Generate Variants]** to auto-create all combinations:

```
| # | SKU        | Barcode       | Color | Size | Price Type | Pricing Config      | Actions |
|---|------------|---------------|-------|------|------------|---------------------|---------|
| 1 | [input]    | [input]       | Red   | S    | [select]   | [expand ▼]          | [🗑]   |
| 2 | [input]    | [input]       | Red   | M    | [select]   | [expand ▼]          | [🗑]   |
| 3 | [input]    | [input]       | Red   | L    | [select]   | [expand ▼]          | [🗑]   |
| 4 | [input]    | [input]       | Blue  | S    | [select]   | [expand ▼]          | [🗑]   |
| ...                                                                                         |
```

Clicking **[expand ▼]** on a variant row reveals the pricing detail form inline (see section 3.2.4).

- Regenerating variants should show a confirmation modal: title "Regenerate Variants", message "This will reset existing variant data. Are you sure you want to continue?", with Cancel and Regenerate buttons.
- User can manually delete individual variants they don't need.
- User can also add a variant manually without using the generator.

**Per-variant fields (applies to both modes):**

| Field | Type | Required | Notes |
|-------|------|----------|-------|
| SKU | Text input | No | Stock Keeping Unit code, unique across all variants |
| Barcode | Text input | No | Barcode number, unique across all variants |
| Price Type | Select | Yes | **Retail** or **Wholesale** |

> **No cost price in the product master.** Cost price comes from purchase orders (dynamic, changes per transaction). The product master only stores pricing **rules** — the actual sell price is resolved at the cashier/POS at transaction time.

**Variant images (Mode B only):**
- Each variant can optionally have its own images (click variant row to expand/show image upload area).
- If no variant images are set, the product-level images are used as fallback.

#### 3.2.4 Variant Pricing

The product master only stores **pricing rules**. The actual sell price is resolved at the cashier/POS at transaction time.

Pricing depends on the **product-level Price Setting** (fixed vs. markup) and the **variant-level Price Type** (retail vs. wholesale).

**A) Fixed Price + Retail:**
- Simple number input for selling price.

```
| Sell Price |
|------------|
| 1000       |
```

**B) Fixed Price + Wholesale (tiered pricing):**
- Table of quantity tiers with fixed sell prices.

```
| Min Qty | Sell Price | Actions |
|---------|------------|---------|
| 1       | 1000       | [🗑]   |
| 10      | 900        | [🗑]   |
| 100     | 800        | [🗑]   |
[+ Add Tier]
```

Rules: Min Qty tier 1 must always be 1 (the default/base price). Tiers must be in ascending order. Sell price should decrease as quantity increases (show warning if not, but don't block — user may have a reason).

**C) Markup Percentage + Retail:**
- Input for markup percentage only. No sell price preview (cost is unknown until transaction time).

```
| Markup % |
|----------|
| 25%      |
```

**D) Markup Percentage + Wholesale (tiered):**
- Table of quantity tiers with markup percentages.

```
| Min Qty | Markup % | Actions |
|---------|----------|---------|
| 1       | 25%      | [🗑]   |
| 10      | 15%      | [🗑]   |
| 100     | 5%       | [🗑]   |
[+ Add Tier]
```

**E) Markup Fixed Amount + Retail:**
- Input for markup amount only.

```
| Markup Amount |
|---------------|
| 200           |
```

**F) Markup Fixed Amount + Wholesale (tiered):**
- Table of quantity tiers with markup amounts.

```
| Min Qty | Markup Amount | Actions |
|---------|---------------|---------|
| 1       | 200           | [🗑]   |
| 10      | 100           | [🗑]   |
| 100     | 50            | [🗑]   |
[+ Add Tier]
```

**How pricing is resolved at transaction time (future — for context only):**

| Price Setting | At the cashier (POS) |
|---|---|
| Fixed Price | Use the stored sell price directly. System shows a recommendation/warning based on latest purchase cost (e.g., "Latest cost: Rp 800, margin: 25%") so the user knows if the fixed price is still profitable. |
| Markup | System looks up the latest cost from purchase order history or current stock, applies the stored markup, and displays the calculated sell price. e.g., latest cost Rp 800 + 25% markup = sell price Rp 1,000. |

**Important pricing notes:**
- All prices/markups are in the **base unit**. When the product is sold in a larger unit, the transaction page (future) will calculate: `unit price = base unit price × conversion factor`.
- Changing the product-level Price Setting (fixed ↔ markup) should show a confirmation modal: title "Change Price Setting", message "Changing price setting will reset variant pricing data. Are you sure you want to continue?", with Cancel and Continue buttons.
- For markup mode, the form shows a helper text: "Sell price will be calculated from purchase cost at transaction time."

### 3.3 Product Form Layout Summary

```
┌─────────────────────────────────────────────────────┐
│  ← Back to Product List          [Save] [Cancel]    │
├─────────────────────────────────────────────────────┤
│                                                     │
│  Product Name:    [________________________]        │
│  Description:     [________________________]        │
│                   [________________________]        │
│  Category:        [Select category ▼      ]        │
│  Images:          [📷 Drop images here / Browse]   │
│                   [img1] [img2] [img3]              │
│  Price Setting:   ○ Fixed Price  ○ Markup Price     │
│  Markup Type:     ○ Percentage   ○ Fixed Amount     │
│  Has Variants:    ○ No  ○ Yes                       │
│  Status:          [🔘 Active]                       │
│                                                     │
├─────────────────────────────────────────────────────┤
│  [ Units ]  [ Variants ]                            │
├─────────────────────────────────────────────────────┤
│                                                     │
│  (Tab content renders here)                         │
│                                                     │
└─────────────────────────────────────────────────────┘
```

### 3.4 State Management

Create a new Zustand store `useProductStore` with:

- Product CRUD operations (add, update, delete)
- Mock data: 3-5 sample products with varying configurations (one with variants, one without, one with wholesale pricing, different unit structures)
- Product type interface covering all fields described above

### 3.5 Mock Data

Create sample products that exercise different scenarios:

1. **Product with linear units + variants + wholesale pricing** (e.g., "T-Shirt" with base unit **Pcs**, Box → Dozen → Pcs, Color/Size variants, wholesale tiers)
2. **Product with branching units + retail pricing** (e.g., "Rice" with base unit **Kg**, Karung = 50 × Kg, Bag = 25 × Kg, fixed retail price)
3. **Product without variants + markup pricing** (e.g., "Notebook" with base unit **Pcs**, Carton = 48 × Pcs, single variant, markup percentage)
4. **Simple product** (e.g., "Cooking Oil" with base unit **Liter**, no other units, no variants, fixed retail price)

---

## 4. New Reusable Components

Build these new components as needed for the product page:

- **Tabs** — tab switcher component (Units | Variants)
- **Select/Dropdown** — form select input with options (for category, unit reference, price type)
- **TagInput** — input that creates tags/chips on Enter (for variant attribute values)
- **ImageUpload** — multi-image upload area with preview and drag-to-reorder
- **Toggle/Switch** — on/off toggle (for product status)
- **Textarea** — multi-line text input (for product description)

---

## 5. Validation Summary

### Product-level
- Name is required
- Category is required
- At least one unit (base unit) must be defined
- Price Setting must be selected

### Unit-level
- Unit name is required and unique within the product
- Conversion factor must be > 0
- No circular references allowed

### Variant-level
- SKU must be unique across all variants in this product (if filled)
- Barcode must be unique across all variants in this product (if filled)
- If Has Variants = No: exactly one variant exists (auto-managed)
- If Has Variants = Yes: at least one variant must exist after generation
- Fixed Price mode: sell price is required for every variant
- Markup mode: markup value (percentage or amount) is required for every variant
- Wholesale tiers: first tier must have Min Qty = 1, tiers must be in ascending qty order

---

## 6. General Notes

- The product form is a **full page**, not a modal. Use Next.js routing: `/master/product/add` and `/master/product/edit/[id]`.
- Show toast notifications on save/delete actions.
- Add a "Back to Product List" link/button at the top of the form page.
- Save and Cancel buttons at the top-right of the form (sticky/visible without scrolling).
- Form should handle unsaved changes warning — if the user clicks Cancel or navigates away with unsaved changes, show a confirmation modal: title "Unsaved Changes", message "You have unsaved changes. Are you sure you want to leave?", with Stay and Leave buttons. (The `beforeunload` browser event should also be used as a fallback for browser tab/close.)

---

## 7. Confirmation Dialogs

**All confirmation dialogs must use the existing `Modal` component** (not `window.confirm()`). This ensures a consistent UI experience throughout the application.

Each confirmation modal should have:
- A clear **title** describing the action
- A **message** explaining the consequence
- A **Cancel** button (secondary/outline style) on the left
- A **Confirm/action** button (primary or danger style depending on context) on the right
- The confirm button label should match the action (e.g., "Continue", "Leave", "Regenerate", "Delete")

Confirmation modals used in the product form:

| Trigger | Title | Message | Cancel | Confirm |
|---------|-------|---------|--------|---------|
| Change Price Setting | Change Price Setting | Changing price setting will reset variant pricing data. Are you sure you want to continue? | Cancel | Continue |
| Toggle Has Variants | Reset Variant Data | Changing this will reset variant data. Are you sure you want to continue? | Cancel | Continue |
| Regenerate variants | Regenerate Variants | This will reset existing variant data. Are you sure you want to continue? | Cancel | Regenerate |
| Cancel / navigate away with unsaved changes | Unsaved Changes | You have unsaved changes. Are you sure you want to leave? | Stay | Leave |
| Delete product (list page) | Delete Product | Are you sure you want to delete **{product name}**? This action cannot be undone. | Cancel | Delete (danger) |
