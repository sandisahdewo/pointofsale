# Skill: E2E Test

Creates new Playwright end-to-end tests and fixes failing ones, using agent-browser for all live page discovery.

## When to Use
- When the user asks to **add, create, or write** e2e tests, Playwright tests, or integration tests for a page or feature
- When the user asks to **fix, debug, or troubleshoot** failing e2e tests, Playwright test errors, or test failures

## Prerequisites
Before writing or fixing any test, you MUST:
1. **Browse the page with agent-browser** — discover all elements, labels, buttons, text, and behavior by interacting with the live page (see Discovery Workflow below)
2. **Check existing helpers** — read `e2e/helpers/` for reusable utilities (e.g., `login`)
3. **Check existing tests** — read tests in the same feature folder for patterns and shared setup

**IMPORTANT**: Do NOT read the page source code (`page.tsx`), component files, stores, or API calls. All discovery must come from agent-browser interaction with the live application.

---

## Part A: Creating New E2E Tests

### Discovery Workflow (agent-browser)

Use `agent-browser` to discover everything about the page before writing tests. Interact with the live app to find element text, form fields, validation messages, toasts, and behavior.

#### Step 1: Login and Navigate

Given the target page route `{route}` (e.g., `master/category`, `settings/users`, `transaction/purchase`):

```bash
# Login
agent-browser open http://localhost:3000/login
agent-browser snapshot -i
agent-browser fill @e1 "admin@pointofsale.com"
agent-browser fill @e2 "Admin@12345"
agent-browser click @e3                          # Login button
agent-browser wait --url "**/dashboard"
agent-browser state save auth.json               # Save auth for reuse

# Navigate to the target page
agent-browser open http://localhost:3000/{route}
agent-browser wait --load networkidle
```

To reuse saved auth in a later session:
```bash
agent-browser state load auth.json
agent-browser open http://localhost:3000/{route}
```

#### Step 2: Discover Page Elements

```bash
# Snapshot the page — get all interactive elements
agent-browser snapshot -i
# Output shows buttons, inputs, links, headings with refs (@e1, @e2, ...)
# Note down: page heading text, button labels, search placeholder, table headers
```

#### Step 3: Discover Add/Create Flow

```bash
# Click the "Add" button (ref from snapshot)
agent-browser click @eN
agent-browser snapshot -i
# Note: modal title or page heading, form field labels, button text (Save/Add/Submit)

# Test validation — submit empty form
agent-browser click @eN                          # Submit/Save button
agent-browser snapshot -i
# Note: exact validation error messages

# Fill form with valid data
agent-browser fill @eN "Test Value"
# Fill other fields...
agent-browser click @eN                          # Submit
agent-browser snapshot -i
# Note: success toast message text, redirect behavior
```

#### Step 4: Discover Edit Flow

```bash
# Find edit button in table row
agent-browser snapshot -i -C
agent-browser click @eN                          # Edit button on a row
agent-browser snapshot -i
# Note: modal title or page heading, pre-filled field values, button text (Save/Update)

# Modify and save
agent-browser fill @eN "Updated Value"
agent-browser click @eN                          # Save/Update button
agent-browser snapshot -i
# Note: success toast message text
```

#### Step 5: Discover Delete Flow

```bash
agent-browser click @eN                          # Delete button on a row
agent-browser snapshot -i
# Note: confirmation modal text, button labels (Cancel/Delete/Confirm)

agent-browser click @eN                          # Confirm delete
agent-browser snapshot -i
# Note: success toast message text
```

#### Step 6: Discover Search, Sort, Pagination

```bash
# Search
agent-browser fill @eN "search term"
agent-browser snapshot -i
# Note: filtered results, debounce behavior

# Sort — click column headers
agent-browser click @eN                          # Column header
agent-browser snapshot -i

# Pagination — check for page controls
agent-browser snapshot -i
# Note: page numbers, prev/next buttons, items per page selector
```

#### Step 7: Multi-Page Features (Product, Purchase Order)

For features with sub-routes, browse each separately:
```bash
agent-browser open http://localhost:3000/{route}/add
agent-browser snapshot -i
# Discover add page layout, tabs, sections

agent-browser open http://localhost:3000/{route}/1
agent-browser snapshot -i
# Discover detail page

agent-browser open http://localhost:3000/{route}/1/edit
agent-browser snapshot -i
# Discover edit page
```

#### Step 8: Close

```bash
agent-browser close
```

#### Discovery Tips
- Always re-snapshot (`agent-browser snapshot -i`) after every click, navigation, or form submission — refs are invalidated on DOM changes
- Use `agent-browser snapshot -i -C` to capture clickable divs (table rows, sort headers, tabs)
- Use `agent-browser get text @eN` to get exact text content of specific elements
- Browse every distinct route the feature uses — list, add, detail, edit pages may all need separate tests
- Watch for toasts — they appear briefly, so snapshot immediately after actions
- Test edge cases: empty states, validation errors, permission-restricted elements

### File to Create

Given feature `{feature}` and page `{page}`:

```
e2e/tests/{feature}/{page}.spec.ts
```

Examples:
- `e2e/tests/master/category.spec.ts`
- `e2e/tests/master/product.spec.ts`
- `e2e/tests/settings/users.spec.ts`
- `e2e/tests/settings/roles.spec.ts`
- `e2e/tests/transaction/purchase-order.spec.ts`

### Test Structure

```typescript
import { test, expect } from '@playwright/test';
import { login } from '@helpers/auth';

test.describe('Page Name', () => {
  test.beforeEach(async ({ page }) => {
    // Login as admin (required for all admin pages)
    await login(page, 'admin@pointofsale.com', 'Admin@12345');
    await page.goto('/route/to/page');
  });

  // Tests go here...
});
```

### What to Test (in order)

#### 1. Page Load & Display
- Heading/title is visible
- Key UI elements render (buttons, search bar, table/list)
- Data loads and displays (at least one row/item visible)

```typescript
test('should display page heading and elements', async ({ page }) => {
  await expect(page.getByRole('heading', { name: 'Page Title' })).toBeVisible();
  await expect(page.getByRole('button', { name: 'Add Entity' })).toBeVisible();
  await expect(page.getByPlaceholder('Search...')).toBeVisible();
});
```

#### 2. Search & Filter
- Type in search box, verify table filters
- Clear search, verify table resets

```typescript
test('should filter items by search query', async ({ page }) => {
  await page.getByPlaceholder('Search...').fill('some term');
  // Wait for debounce (300ms) + render
  await expect(page.getByRole('cell', { name: 'Expected Match' })).toBeVisible();
  await expect(page.getByRole('cell', { name: 'Should Not Match' })).not.toBeVisible();
});
```

#### 3. Create (Add New)
- Open add modal/navigate to add page
- Submit empty form → verify validation errors
- Fill form with valid data → submit
- Verify success toast and new item appears

```typescript
test('should show validation errors on empty submit', async ({ page }) => {
  await page.getByRole('button', { name: 'Add Entity' }).click();
  // For modal-based forms:
  await page.getByRole('button', { name: 'Save' }).click();
  await expect(page.getByText('Name is required')).toBeVisible();
});

test('should create a new entity', async ({ page }) => {
  await page.getByRole('button', { name: 'Add Entity' }).click();
  await page.getByLabel('Name').fill('New Entity');
  // Fill other required fields...
  await page.getByRole('button', { name: 'Save' }).click();
  await expect(page.getByText(/created successfully/i)).toBeVisible({ timeout: 10000 });
});
```

#### 4. Read (View/Detail)
- For pages with detail view: click item → verify detail page loads
- Verify key data fields are displayed

#### 5. Update (Edit)
- Click edit button on existing item
- Verify form pre-fills with current data
- Modify a field → submit
- Verify success toast and updated data displays

```typescript
test('should edit an existing entity', async ({ page }) => {
  // Click edit on first row
  await page.getByRole('row').nth(1).getByRole('button', { name: 'Edit' }).click();
  // Verify pre-filled data
  await expect(page.getByLabel('Name')).not.toBeEmpty();
  // Modify and save
  await page.getByLabel('Name').fill('Updated Name');
  await page.getByRole('button', { name: 'Save' }).click();
  await expect(page.getByText(/updated successfully/i)).toBeVisible({ timeout: 10000 });
});
```

#### 6. Delete
- Click delete button → verify confirmation modal
- Cancel → verify item still exists
- Confirm → verify success toast and item removed

```typescript
test('should delete an entity with confirmation', async ({ page }) => {
  await page.getByRole('row').nth(1).getByRole('button', { name: 'Delete' }).click();
  // Confirmation modal appears
  await expect(page.getByText(/are you sure/i)).toBeVisible();
  await page.getByRole('button', { name: 'Delete' }).click();
  await expect(page.getByText(/deleted successfully/i)).toBeVisible({ timeout: 10000 });
});
```

#### 7. Sorting (if table has sortable columns)
- Click column header → verify sort direction changes
- Verify data reorders

#### 8. Pagination (if page has pagination)
- Verify pagination controls show (page numbers, prev/next)
- Navigate to next page → verify different data shows
- Change page size → verify rows per page changes

#### 9. Feature-Specific Tests
- **Status badges**: verify correct color/text for each status
- **Status transitions** (transactions): test action buttons per status
- **Nested forms** (supplier bank accounts, PO line items): add/remove items
- **Tabs** (product form): switch tabs, verify content persists
- **Conditional rendering**: elements that show/hide based on state
- **Permission-based UI**: buttons disabled/hidden for non-authorized users

### Page Type Templates

#### Simple CRUD (Category, Rack, Supplier)
All operations happen on a single page with modals:
- Page load with table
- Search/filter
- Add via modal → validate → save
- Edit via modal → verify prefill → update
- Delete via confirmation modal
- Sort columns
- Pagination

#### Complex CRUD (Product)
Multi-page flow with separate routes:
- List page: table, search, sort, pagination, delete
- Add page (`/add`): form with tabs/sections → validate → save → redirect to list
- Edit page (`/edit/[id]`): form prefilled → update → redirect to list
- Test tab switching and complex form interactions

#### Transaction (Purchase Order)
Card-based list with status workflow:
- List page: status tabs, search, cards with actions
- Add page: supplier select, date, line items (add/remove), save as draft
- Detail page: view PO info, items, status-based action buttons
- Status transitions: draft → sent → received → completed
- Receive page: enter received quantities
- Delete (draft only)

---

## Part B: Fixing Failing E2E Tests

### Step 1: Run the Failing Test

First, run the specific failing test to get the actual error output. Do NOT skip this step — you need the real error log to understand what's broken.

```bash
# Run a specific test file
cd e2e && npx playwright test tests/{feature}/{page}.spec.ts

# Or run a single test by name
cd e2e && npx playwright test -g "test name"
```

Read the error output carefully. Common error types:
- **Timeout / Element not found** — selector doesn't match the live UI
- **Unexpected value** — text, attribute, or count differs from expectation
- **Navigation error** — page didn't reach expected URL
- **Strict mode violation** — selector matches multiple elements
- **Assertion failure** — expected vs actual mismatch

Also read the failing test file to understand what it's trying to do and which selectors/assertions it uses.

### Step 2: Browse the Live Page with agent-browser

Use agent-browser to inspect the actual page state. Navigate to the same page the test targets and snapshot to see current elements.

```bash
# Login (or load saved state)
agent-browser open http://localhost:3000/login
agent-browser snapshot -i
agent-browser fill @e1 "admin@pointofsale.com"
agent-browser fill @e2 "Admin@12345"
agent-browser click @e3
agent-browser wait --url "**/dashboard"
agent-browser state save auth.json

# Navigate to the page under test
agent-browser open http://localhost:3000/{route}
agent-browser wait --load networkidle
agent-browser snapshot -i
```

### Step 3: Reproduce the Failing Interaction

Walk through the same steps the test performs using agent-browser. Focus on the specific step that failed in the error log:

```bash
# Example: if the test clicks a button and expects a modal
agent-browser click @eN              # Click the same button
agent-browser snapshot -i            # See what actually appeared
# Compare with what the test expects
```

Key discovery commands:
- `agent-browser snapshot -i` — get all interactive elements with refs
- `agent-browser snapshot -i -C` — include clickable divs (table rows, tabs, sort headers)
- `agent-browser get text @eN` — get exact text content of a specific element
- `agent-browser screenshot` — capture visual state for layout issues

After every click, navigation, or form submission, **always re-snapshot** — refs are invalidated on DOM changes.

### Step 4: Identify Root Cause

Compare the error output from Step 1 with what agent-browser reveals in Steps 2-3:

| Error Type | What to Check with agent-browser |
|---|---|
| Element not found | Snapshot the page — did the label/text/role change? |
| Timeout waiting for selector | Is the element behind a loading state? Does it appear after an API call? |
| Strict mode violation | Snapshot and count matching elements — add `exact: true` or narrow the scope |
| Wrong text/value | Use `get text @eN` to see the actual text |
| Navigation didn't happen | Check if a modal blocked navigation, or URL pattern changed |
| Toast not visible | Toasts are brief — snapshot immediately after the triggering action |

### Step 5: Fix the Test

Apply the fix based on what you discovered. Common fixes:

**Wrong selector** — update to match the actual element:
```typescript
// Before (wrong label)
await page.getByLabel('Username').fill('admin');
// After (discovered correct label via snapshot)
await page.getByLabel('Email').fill('admin@pointofsale.com');
```

**Timing issue** — add proper waits:
```typescript
// Before (race condition)
await page.getByRole('button', { name: 'Save' }).click();
await expect(page.getByText('Saved')).toBeVisible();
// After (wait for API response)
await page.getByRole('button', { name: 'Save' }).click();
await expect(page.getByText('Saved')).toBeVisible({ timeout: 10000 });
```

**Strict mode** — narrow the selector scope:
```typescript
// Before (matches multiple "Delete" buttons)
await page.getByRole('button', { name: 'Delete' }).click();
// After (scope to specific row)
await page.getByRole('row').nth(1).getByRole('button', { name: 'Delete' }).click();
```

**Changed UI structure** — update selectors and assertions to match:
```typescript
// Before (old heading text)
await expect(page.getByRole('heading', { name: 'Users' })).toBeVisible();
// After (heading changed)
await expect(page.getByRole('heading', { name: 'User Management' })).toBeVisible();
```

### Step 6: Verify the Fix

Re-run the same test from Step 1 to confirm the fix works:
```bash
cd e2e && npx playwright test tests/{feature}/{page}.spec.ts
```

Or run a single test by name:
```bash
cd e2e && npx playwright test -g "test name"
```

If the test still fails, go back to Step 2 and re-inspect with agent-browser.

### Step 7: Close

```bash
agent-browser close
```

---

## Reference

### Helpers

#### When to Create a New Helper
If a setup step is used in 3+ test files, extract it to `e2e/helpers/{name}.ts`:
- Navigation helpers (e.g., `navigateToCategory`)
- Data creation helpers (e.g., `createCategory`)
- Common assertion helpers

#### Existing Helpers
- `@helpers/auth` → `login(page, email, password)` — logs in and waits for dashboard

### Selector Priorities (most reliable first)
1. `page.getByRole('button', { name: '...' })` — buttons, links, headings
2. `page.getByLabel('...')` — form inputs with labels
3. `page.getByPlaceholder('...')` — search inputs
4. `page.getByText('...')` — visible text content
5. `page.getByRole('row')` — table rows
6. `page.getByRole('cell', { name: '...' })` — table cells
7. `page.locator('[data-testid="..."]')` — last resort, requires adding data-testid

### Key Conventions
- All admin page tests must login first via `login()` helper in `beforeEach`
- Use `{ timeout: 10000 }` for assertions that wait on API calls (toasts, redirects, data loading)
- Use `{ exact: true }` on `getByLabel` when labels overlap (e.g., `'Password'` vs `'Confirm Password'`)
- Test file names match the page name: `category.spec.ts`, `users.spec.ts`, `purchase-order.spec.ts`
- Group related tests in `test.describe()` blocks (one per page, nested for sub-features if needed)
- Each test should be independent — don't rely on state from previous tests
- Use `test.describe.serial()` only when test order genuinely matters (e.g., create → edit → delete flow that shares server state)
- Avoid hardcoding IDs — find elements by visible text/role
- For date assertions, use partial matching or regex to avoid timezone issues
- For currency assertions, match the formatted value (e.g., `'Rp 1.000.000'` for IDR)

### Running Tests
```bash
# Docker (recommended)
docker compose --profile test run --rm e2e

# Locally
cd e2e && npm test                               # all tests headless
cd e2e && npx playwright test path/to.spec.ts    # single file
cd e2e && npx playwright test -g "test name"     # single test by name
cd e2e && npm run test:headed                    # with visible browser
cd e2e && npm run test:debug                     # step-by-step debugger
cd e2e && npm run report                         # view HTML report
```
