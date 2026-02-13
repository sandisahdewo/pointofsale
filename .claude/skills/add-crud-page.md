# Skill: Add CRUD Page

Creates a new simple CRUD page (like Category, Rack, Supplier) with all supporting files.

## When to Use
When the user asks to add a new master data page or simple CRUD entity.

## Files to Create

Given entity name `{Entity}` (e.g., `Brand`) and route `{route}` (e.g., `master/brand`):

### 1. Mock Data — `frontend/src/data/{entities}.ts`
- Import the Entity type from the store
- Export `initial{Entities}` array with 8-13 seed records
- Each record has `id: number` + entity-specific fields

### 2. Zustand Store — `frontend/src/stores/use{Entity}Store.ts`
- `'use client'` directive at top
- Import `create` from `zustand` and initial data from `@/data/{entities}`
- Export interface `{Entity}` with `id: number` + fields
- Export interface `{Entity}State` with: items array, add/update/delete actions
- ID generation: `state.items.reduce((max, i) => Math.max(max, i.id), 0) + 1`
- Pattern: follow `useCategoryStore.ts` exactly

### 3. Page — `frontend/src/app/{route}/page.tsx`
- `'use client'` directive
- Wrap in `<AdminLayout>`
- Include: search, sort, pagination (DEFAULT_PAGE_SIZE = 10)
- Add/Edit Modal with form validation
- Delete confirmation Modal
- Table with columns + action buttons (Edit, Delete)
- Use `useToastStore` for success messages
- Pattern: follow `master/category/page.tsx` exactly

### 4. Update Sidebar — `frontend/src/components/layout/Sidebar.tsx`
- Add new menu item to appropriate section in `menuItems` array

## Key Conventions
- All files start with `'use client'`
- Table generic requires `T extends { id: number }`
- Use Tailwind default palette (blue-600 primary, gray tones)
- Form validation: track `formErrors` as `Record<string, string>`
- Toast messages: `addToast('Entity created successfully', 'success')`
- Sorting uses the three-state cycle: null → asc → desc → null
