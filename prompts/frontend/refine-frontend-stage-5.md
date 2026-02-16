# Stage 5 — Sales Transaction Page (`/transaction/sales`)

> All data is mock/in-memory via Zustand. No backend connection.

---

## Overview

Build a multi-session sales page where users sell product variants to buyers. Multiple sessions can run simultaneously, each with its own independent cart. The page lives at `frontend/src/app/transaction/sales/page.tsx`, uses `'use client'`, and wraps in `<AdminLayout>`.

---

## 1. Update Mock Data — Add Stock

Before building the sales page, update `frontend/src/data/products.ts` to give variants realistic `currentStock` values so the page is testable. Set a mix of stock levels (e.g., some at 50, some at 100, some at 5, keep one or two at 0 to test the out-of-stock UI).

---

## 2. Sales Session Management

### 2.1 Session Tabs

- Display sessions as **horizontal tabs** at the top of the page.
- Each tab shows the session name.
- A **"+" button** at the end of the tabs creates a new session.
- Default session is created automatically when the page loads (so user never sees an empty state).
- **Maximum 10 sessions** allowed. Disable the "+" button when limit is reached.

### 2.2 Session Naming

- Sessions are **auto-numbered**: "Session 1", "Session 2", etc.
- Numbering increments from the highest number ever created in the current page lifecycle (so closing Session 2 then creating a new one gives "Session 3", not "Session 2").

### 2.3 Closing a Session

- Each tab has a **close (X) button**.
- Use a **two-click confirmation pattern** (seamless, no modal):
  1. **First click**: The X icon changes to a **checkmark (confirm) icon** to indicate "click again to confirm".
  2. **Second click** on the confirm icon: The session closes and is removed.
  3. If the user clicks elsewhere or switches tabs, the confirm icon **reverts back to the X icon** (cancel the close).
- If only one session remains, the close button should still work — closing it creates a new default session automatically (user always has at least one session).
- Closing a session discards its cart data.

### 2.4 Switching Sessions

- Clicking a tab switches to that session instantly.
- Each session maintains its own **independent state**: search query, search results, cart items, and any in-progress form state.
- Switching between sessions preserves all state — no data is lost.

---

## 3. Product Search

### 3.1 Search Input

- A search input field with a **search button** (icon or labeled "Search").
- Placeholder text: "Search by product name, SKU, or barcode..."
- **Minimum 3 characters** required before search executes.
- **Enter key** triggers search (same as clicking the search button).
- If fewer than 3 characters, show an inline hint: "Type at least 3 characters".

### 3.2 Search Logic

- Search against these fields (case-insensitive):
  - **Product name** (`product.name`)
  - **Variant SKU** (`variant.sku`)
  - **Variant barcode** (`variant.barcode`)
- **Exclude inactive products** (`status: 'inactive'` should not appear).
- **Limit results to 10 products** max.
- If no products match, display a message: "No results found".

### 3.3 Search Results Dropdown

- Results appear in a **dropdown below the search input**.
- Dropdown has a **fixed max-height** and **scrolls internally** when content overflows.
- Dropdown stays open until the user explicitly closes it via a **close button** (X or "Close" at the top/bottom of the dropdown). Do NOT auto-close when a variant is selected.
- Dropdown layout:

#### For products WITH variants (`hasVariants: true`):

```
[Product Image] Product Name
  [Variant Image] | SKU | Attribute values (e.g. "Red, S") | Stock: N | [Select Button]
  [Variant Image] | SKU | Attribute values (e.g. "Red, M") | Stock: N | [Select Button]
  [Variant Image] | SKU | Attribute values (e.g. "Blue, L") | Stock: N | [Select Button]
```

#### For products WITHOUT variants (`hasVariants: false`):

```
[Product Image] | Product Name | SKU | Stock: N | [Select Button]
```

Show as a single selectable row — no nested variant list.

### 3.4 Image Handling

- Use the first image from `product.images` or `variant.images` array.
- If no images exist (empty array), show a **placeholder icon** (e.g., a generic product/image icon).

### 3.5 Out-of-Stock Variants

- If `currentStock === 0`:
  - Variant row has a **soft red background** (e.g., `bg-red-50`).
  - The **Select button is disabled**.
  - The variant **cannot be added to the cart**.

---

## 4. Cart

### 4.1 Adding Items to Cart

- When a variant is selected from search results, it is added to the **active session's cart** with **quantity = 1** and **unit = base unit** (the unit where `isBase: true`).
- If the **same variant already exists** in the cart, **increment its quantity by 1** instead of adding a duplicate row.
- After adding, keep the search dropdown open (user may want to add more items).

### 4.2 Cart Table Layout

Each cart row displays:

```
Row 1: [Variant Image] | SKU | Name | Attributes | Quantity Input | Unit Selector | Price | Total | Actions (remove)
Row 2:                  | Stock: N | Description
```

- **Variant Image**: Same placeholder logic as search results if no image.
- **SKU**: Variant SKU.
- **Name**: Product name.
- **Attributes**: Variant attribute key-value pairs (e.g., "Color: Red, Size: S"). Empty for non-variant products.
- **Quantity Input**: Editable number input. **Minimum value is 1**. Do not allow 0 or negative.
- **Unit Selector**: Dropdown of available units from the product's `units` array. Default is `base_unit` (where `isBase: true`).
- **Price**: Per-unit price (considering tiered pricing and selected unit).
- **Total**: `quantity × price`.
- **Actions**: Remove button to delete the item from the cart.
- **Stock**: Show `currentStock` value (always in base unit).
- **Description**: Product description.

### 4.3 Unit Change Behavior

When the user changes the unit in the cart:
- **Quantity stays the same** (the number does not convert).
- **Price changes** to reflect the new unit: `base_price × unit.toBaseUnit`.
- Example: If base price is Rp 75.000/Pcs, and user selects Dozen (toBaseUnit: 12), price becomes Rp 900.000/Dozen. If quantity is 2, total = 2 × Rp 900.000 = Rp 1.800.000.

### 4.4 Tiered Pricing

- Pricing tiers are defined on the variant as `pricingTiers: [{ minQty, value }]`.
- To determine the active tier, **convert the cart quantity to base-unit quantity**:
  - `baseQty = quantity × selectedUnit.toBaseUnit`
- Find the tier where `baseQty >= tier.minQty` with the **highest matching minQty**.
- The tier `value` is the **per-base-unit price**. Multiply by `selectedUnit.toBaseUnit` for the displayed per-unit price.
- **Visually indicate** when a tiered price is active (e.g., show the original price struck-through, or a small badge like "Tier price").
- Example: T-Shirt Red S has tiers `[{minQty:1, value:75000}, {minQty:12, value:70000}]`. If user sets qty=1 Dozen (baseQty=12), tier 2 applies → price = 70000 × 12 = Rp 840.000/Dozen.

### 4.5 Stock Validation

- Stock (`currentStock`) is always in **base units**.
- When quantity or unit changes, calculate `baseQty = quantity × selectedUnit.toBaseUnit`.
- If `baseQty > currentStock`, show an **error message** on that cart row (e.g., "Insufficient stock. Available: {currentStock} {baseUnitName}").
- **Do not block the checkout button** based on stock — just show the warning (the user may know stock will be replenished). OR: block checkout if any item exceeds stock. **Decision: block checkout if any row has a stock error.**

### 4.6 Empty Cart

- If the cart is empty, show a message: "Cart is empty. Search and add products to get started."

---

## 5. Cart Summary

Below the cart items, show a summary section:

```
Total Items: {count of cart rows}
Subtotal: Rp {sum of all row totals}
Grand Total: Rp {same as subtotal — no discount/tax for now}
```

---

## 6. Payment Method

Below the cart summary, show a **payment method selector**:

- Three options: **Cash**, **Card**, **QRIS**
- Use radio buttons or selectable cards/chips.
- Default: none selected (user must choose before checkout).
- **Simple selection only** — no extra input fields (no amount tendered, no card details, no QR).

---

## 7. Checkout

- A **"Checkout" button** below the payment method section.
- Button is **disabled** if:
  - Cart is empty.
  - No payment method is selected.
  - Any cart row has a stock error (quantity exceeds stock).
- On click:
  - Show a **success message** (toast notification).
  - Deduct stock from the product variants in the product store (`currentStock -= baseQty` for each item).
  - Show the **receipt** (see below).

---

## 8. Receipt

After checkout, display a receipt overlay/modal:

### Receipt Content (Basic)

```
================================
         SALES RECEIPT
================================
Date: {date and time of transaction}
Transaction: #{auto-incremented ID}
--------------------------------
Item Name         Qty  Price  Total
  SKU | Attributes
Item Name         Qty  Price  Total
  SKU | Attributes
--------------------------------
Total Items:    {N}
Subtotal:       Rp {amount}
Grand Total:    Rp {amount}
Payment:        {Cash/Card/QRIS}
================================
```

### Receipt Actions

- **Print**: Trigger browser `window.print()` with a print-friendly stylesheet.
- **Save as PDF**: Use browser print-to-PDF (same `window.print()` flow), or generate a downloadable PDF.
- **Close**: Close the receipt and reset the session (clear cart, clear search, ready for next transaction). The session tab stays open.

---

## 9. Currency Formatting

- Use **Indonesian Rupiah (IDR)** format: `Rp {amount}`.
- Use dot as thousands separator: `Rp 75.000`, `Rp 1.800.000`.
- No decimal places (IDR doesn't use decimals in daily transactions).
- Create a reusable `formatCurrency(amount: number): string` utility function.

---

## 10. Zustand Store — `useSalesStore`

Create a new store at `frontend/src/stores/useSalesStore.ts` to manage:

### State Shape

```typescript
interface CartItem {
  productId: number;
  variantId: string;
  quantity: number;
  selectedUnitId: string; // from product.units
}

interface SalesSession {
  id: number;
  name: string;              // "Session 1", "Session 2", etc.
  cart: CartItem[];
  paymentMethod: 'cash' | 'card' | 'qris' | null;
}

interface SalesState {
  sessions: SalesSession[];
  activeSessionId: number;
  nextSessionNumber: number;  // always increments, never reuses
  transactionCounter: number; // for receipt transaction ID

  // Session actions
  createSession: () => void;
  closeSession: (id: number) => void;
  setActiveSession: (id: number) => void;

  // Cart actions
  addToCart: (sessionId: number, productId: number, variantId: string) => void;
  updateCartItemQuantity: (sessionId: number, variantId: string, quantity: number) => void;
  updateCartItemUnit: (sessionId: number, variantId: string, unitId: string) => void;
  removeFromCart: (sessionId: number, variantId: string) => void;

  // Payment & checkout
  setPaymentMethod: (sessionId: number, method: 'cash' | 'card' | 'qris') => void;
  checkout: (sessionId: number) => void; // returns transaction data for receipt
  resetSession: (sessionId: number) => void; // clear cart + payment after checkout
}
```

The store should reference `useProductStore` for product/variant data when needed (for display, pricing calculation, stock checks). Cart items only store IDs and quantity/unit — derive everything else from the product store.

---

## 11. File Structure

```
frontend/src/
├── app/transaction/sales/
│   └── page.tsx                    # Main sales page
├── components/sales/
│   ├── SessionTabs.tsx             # Tab bar with session management
│   ├── ProductSearch.tsx           # Search input + dropdown
│   ├── SearchResultsDropdown.tsx   # Search results display
│   ├── Cart.tsx                    # Cart table
│   ├── CartSummary.tsx             # Totals display
│   ├── PaymentMethodSelector.tsx   # Cash/Card/QRIS selection
│   └── Receipt.tsx                 # Receipt display with print/PDF
├── stores/
│   └── useSalesStore.ts            # Sales session & cart state
├── utils/
│   └── currency.ts                 # formatCurrency utility
└── data/
    └── products.ts                 # (update existing: add stock values)
```

---

## 12. Summary of Key Decisions

| Topic | Decision |
|-------|----------|
| Session display | Horizontal tabs at top |
| Max sessions | 10 |
| Session naming | Auto-numbered (never reuses numbers) |
| Currency | IDR — `Rp 75.000` |
| Unit change behavior | Quantity stays, price changes |
| Tiered pricing threshold | Convert to base unit before comparing |
| Duplicate cart item | Increment quantity |
| Inactive products | Hidden from search |
| No-variant products | Show as single selectable row |
| Payment method | Simple selection only (Cash/Card/QRIS) |
| Post-checkout | Session stays open, cart resets |
| Receipt | Basic: date, items, totals, payment method |
| Stock error | Blocks checkout |
| Search dropdown | Fixed max-height with scroll |
| Search dropdown close | Manual close button only |
| Search scope | Product name, variant SKU, variant barcode |
