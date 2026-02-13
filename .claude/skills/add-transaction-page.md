# Skill: Add Transaction Page

Creates a new transaction page (like Purchase Order) with list, add, detail, and edit views.

## When to Use
When the user asks to add a new transaction type (e.g., Sales Order, Stock Transfer).

## Files to Create

Given entity `{Entity}` (e.g., `SalesOrder`) and route `transaction/{route}`:

### 1. Mock Data — `frontend/src/data/{entityName}s.ts`
- Export `initial{Entities}` with sample records
- Include status field (e.g., `'draft' | 'confirmed' | 'completed' | 'cancelled'`)
- Include line items as nested array
- Include metadata: createdAt, updatedAt dates

### 2. Zustand Store — `frontend/src/stores/use{Entity}Store.ts`
- Standard CRUD actions (add, update, delete)
- `getItem(id)` getter using `get().items.find()`
- Status-specific actions if needed (e.g., `confirmOrder`, `cancelOrder`)
- Pattern: follow `usePurchaseOrderStore.ts`

### 3. List Page — `frontend/src/app/transaction/{route}/page.tsx`
- `<AdminLayout>` wrapper
- Table with status badges (`<StatusBadge>` or `<Badge>`)
- Search, sort, pagination
- "Add New" button linking to `transaction/{route}/add`
- Row actions: View, Edit, Delete
- Pattern: follow `transaction/purchase/page.tsx`

### 4. Add Page — `frontend/src/app/transaction/{route}/add/page.tsx`
- Thin wrapper that renders the shared Form component in "add" mode

### 5. Form Component — `frontend/src/components/{entity}/{Entity}Form.tsx`
- Shared between add and edit
- Props: `mode: 'add' | 'edit'`, optional `initialData`
- Multi-section form with line items table
- Line item add/remove/edit
- Totals calculation
- Form validation
- Navigation back to list on save via `router.push()`
- Pattern: follow `components/purchase/PurchaseOrderForm.tsx`

### 6. Detail Page — `frontend/src/app/transaction/{route}/[id]/page.tsx`
- Read-only view of the transaction
- Status display with badge
- Line items table (non-editable)
- Action buttons based on status

### 7. Edit Page — `frontend/src/app/transaction/{route}/[id]/edit/page.tsx`
- Thin wrapper rendering Form component in "edit" mode
- Loads existing data from store

### 8. Update Sidebar
- Add menu item under "Transaction" section in `components/layout/Sidebar.tsx`

## Key Conventions
- Use `'use client'` on all pages
- Use `next/navigation` for `useRouter` and `useParams`
- Transaction IDs displayed with prefix (e.g., "PO-001", "SO-001")
- Status flow typically: draft → confirmed → completed (with cancelled as alternative)
- Line items have: product reference, quantity, unit price, subtotal
- Store uses `Date.now().toString()` or similar for generated IDs in line items
