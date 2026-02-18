# Skill: E2E Test

Creates, edits, and fixes Playwright end-to-end tests. Always discovers the live UI with agent-browser before writing any test code.

## When to Use
- Creating new e2e tests for a page or feature
- Fixing failing e2e tests
- Editing or modifying existing e2e tests

## MANDATORY: Use agent-browser First

**Before writing, editing, or fixing ANY test code, you MUST browse the live page with `agent-browser`.** No exceptions.

- **Creating** → browse every flow (page load, add, edit, delete, search, sort, pagination)
- **Editing** → browse to verify the current UI matches what the test expects
- **Fixing** → reproduce the failing interaction to see actual vs expected state

**Do NOT read page source code** (page.tsx, components, stores, API handlers). All element names, labels, text, roles, and behavior MUST come from agent-browser interaction with the live running application.

---

## Discovery Checklist

Use the **agent-browser skill** for all browser interactions. Refer to its documentation for commands (open, snapshot, fill, click, wait, state save/load, close, etc.).

### Login & Navigate
1. Open `http://localhost:3000/login`, snapshot, fill credentials, click login, wait for dashboard
2. Save auth state for reuse across discovery sessions
3. Navigate to the target page route, wait for network idle

### What to Discover

For each page, snapshot and note down:

- [ ] **Page load**: heading text, button labels, search placeholder, table column headers
- [ ] **Add/Create flow**: click Add button → modal title or page heading, form field labels/placeholders, submit button text → submit empty → validation behavior → fill valid data → submit → toast message
- [ ] **Edit flow**: click Edit on a row → modal title, pre-filled values, submit button text → modify → submit → toast message
- [ ] **Delete flow**: click Delete → confirmation modal text, button labels → confirm → toast message
- [ ] **Search**: placeholder text, debounce behavior, empty state text, reset behavior
- [ ] **Sorting**: which columns are sortable, sort indicator behavior (active/inactive states)
- [ ] **Pagination**: "Showing X-Y of Z items" text, page size selector label and options
- [ ] **Special elements**: active/status toggles, dynamic rows (bank accounts), tabs, status badges

Tips:
- Use snapshot with `-C` flag to capture clickable divs (table rows, sort headers, tabs)
- Watch for toasts — snapshot immediately after triggering actions
- For multi-page features, browse each route separately (`/add`, `/edit/[id]`)
- Always close the browser when done

---

## Workflow: Creating Tests

### Step 1: Prerequisites
1. **Browse the page** with agent-browser (see Discovery Checklist above)
2. **Read existing helpers** in `e2e/helpers/` for reusable utilities
3. **Read existing tests** in the same feature folder for patterns and shared setup

### Step 2: Create Test File

File path: `e2e/tests/{feature}/{page}.spec.ts`

Examples: `tests/master/category.spec.ts`, `tests/auth/login.spec.ts`, `tests/transaction/purchase-order.spec.ts`

### Step 3: Write Tests

Use the structure and patterns from the reference sections below. Write tests in this order:

1. Page load & display (heading, buttons, table headers)
2. Data display (rows exist in table)
3. Pagination info
4. Search (filter, empty state, reset)
5. Create (open modal, validation, successful create, cancel)
6. Edit (prefill check, update, validation)
7. Delete (confirmation modal, cancel, confirm delete)
8. Sorting (column click cycles: asc → desc → none)
9. Page size (change items per page)
10. Feature-specific (toggles, dynamic rows, tabs, status badges)

---

## Workflow: Fixing Failing Tests

### Step 1: Run the Failing Test
```bash
cd e2e && npx playwright test tests/{feature}/{page}.spec.ts
# Or single test by name:
cd e2e && npx playwright test -g "test name"
```

Read the error output carefully. Note the error type:
- **Timeout / Element not found** — selector doesn't match the live UI
- **Strict mode violation** — selector matches multiple elements
- **Assertion failure** — expected vs actual mismatch
- **Navigation error** — page didn't reach expected URL

### Step 2: Read the Failing Test File
Understand what the test is trying to do and which selectors/assertions it uses.

### Step 3: Browse with agent-browser
Use the **agent-browser skill** to navigate to the same page and reproduce the failing interaction. Snapshot to compare actual element text/roles against what the test expects.

### Step 4: Identify Root Cause

| Error Type | What to Check |
|---|---|
| Element not found | Snapshot — did the label/text/role change? |
| Timeout | Is the element behind a loading state? Is timeout too low? |
| Strict mode violation | Count matching elements — add `exact: true` or narrow scope |
| Wrong text/value | Get text of the element to see actual content |
| Navigation error | Check if a modal blocked navigation, or URL changed |
| Toast not visible | Snapshot immediately after triggering action |

### Step 5: Fix, Then Verify
Apply the fix, then re-run the test:
```bash
cd e2e && npx playwright test tests/{feature}/{page}.spec.ts
```

If still failing, go back to Step 3 and re-inspect with agent-browser.

### Step 6: Close
Close the browser session when done.

---

## Test Structure

### Admin Pages (require login)

```typescript
import { test, expect } from '@playwright/test';
import { login } from '@helpers/auth';

test.describe('Page Name', () => {
  test.beforeEach(async ({ page }) => {
    await login(page, 'admin@pointofsale.com', 'Admin@12345');
    await page.goto('/route/to/page');
    await expect(page.getByRole('heading', { name: 'Page Title' })).toBeVisible({ timeout: 10000 });
  });

  // Tests...
});
```

### Auth Pages (no login needed)

```typescript
import { test, expect } from '@playwright/test';

test.describe('Page Name', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/login');
  });

  // Tests...
});
```

---

## Patterns Reference

All patterns below are extracted from the real test files. Use these as the authoritative reference.

### Unique Test Data
Use `Date.now()` to generate unique names so tests don't collide:
```typescript
const categoryName = `Test Category ${Date.now()}`;
const rackCode = `TR-${Date.now()}`;
```

### HTML5 Native Validation
Check browser-native required field validation (not custom error messages):
```typescript
await page.getByRole('button', { name: 'Create' }).click();
const nameInput = page.getByLabel('Name');
expect(await nameInput.evaluate((el: HTMLInputElement) => el.validity.valueMissing)).toBe(true);
```

Verify validation clears after filling:
```typescript
await nameInput.fill('Test');
expect(await nameInput.evaluate((el: HTMLInputElement) => el.validity.valid)).toBe(true);
```

For textareas (e.g., Address):
```typescript
const addressTextarea = page.getByPlaceholder('Supplier address');
expect(await addressTextarea.evaluate((el: HTMLTextAreaElement) => el.validity.valueMissing)).toBe(true);
```

### Custom Validation Errors (Auth Pages)
Auth pages use custom error messages, not native validation:
```typescript
await page.getByRole('button', { name: 'Login' }).click();
await expect(page.getByText('Email is required')).toBeVisible();
await expect(page.getByText('Password is required')).toBeVisible();
```

### Browser Email Validation
```typescript
await emailInput.fill('not-an-email');
await page.getByRole('button', { name: 'Login' }).click();
await expect(emailInput).toHaveJSProperty('validity.typeMismatch', true);
```

### Create-Search-Act Flow
For edit/delete tests, create fresh data first, search for it, then act. This avoids depending on pre-existing data:
```typescript
test('should update an entity', async ({ page }) => {
  // 1. Create
  const name = `Edit Me ${Date.now()}`;
  await page.getByRole('button', { name: 'Add Entity' }).click();
  await page.getByLabel('Name').fill(name);
  await page.getByRole('button', { name: 'Create' }).click();
  await expect(page.getByText('Entity created successfully')).toBeVisible({ timeout: 10000 });

  // 2. Search
  await page.getByPlaceholder('Search entities...').fill(name);
  await expect(page.getByRole('cell', { name })).toBeVisible({ timeout: 5000 });

  // 3. Act (edit)
  await page.getByRole('button', { name: 'Edit' }).click();
  await page.getByLabel('Name').fill(`${name} Updated`);
  await page.getByRole('button', { name: 'Update' }).click();
  await expect(page.getByText('Entity updated successfully')).toBeVisible({ timeout: 10000 });
});
```

### Delete Confirmation
The delete confirmation modal shows the entity name in bold. Use `.nth(1)` for the modal's Delete button since the row's Delete button is `.nth(0)`:
```typescript
await page.getByRole('button', { name: 'Delete' }).click();
await expect(page.getByText(/Are you sure you want to delete/)).toBeVisible();
await expect(page.locator('strong', { hasText: entityName })).toBeVisible();

// Confirm delete (second Delete button = modal's)
await page.getByRole('button', { name: 'Delete' }).nth(1).click();
await expect(page.getByText('Entity deleted successfully')).toBeVisible({ timeout: 10000 });
```

### Search with Debounce
Search has a ~300ms debounce. Use `{ timeout: 5000 }`:
```typescript
await page.getByPlaceholder('Search entities...').fill('search term');
await expect(page.getByRole('cell', { name: 'Expected Match' })).toBeVisible({ timeout: 5000 });
```

Empty state:
```typescript
await page.getByPlaceholder('Search entities...').fill('zzz_nonexistent_xyz');
await expect(page.getByText('No data available')).toBeVisible({ timeout: 5000 });
```

Reset:
```typescript
await page.getByPlaceholder('Search entities...').fill('');
await expect(page.getByText('No data available')).not.toBeVisible({ timeout: 5000 });
```

### Column Sorting
Sort cycles: asc → desc → none. Active sort shows blue, inactive shows gray:
```typescript
// Click to sort ascending
await page.getByRole('columnheader', { name: /Name/i }).click();
await expect(page.getByRole('columnheader', { name: /Name/i }).locator('span.text-blue-600')).toBeVisible();

// Click again for descending
await page.getByRole('columnheader', { name: /Name/i }).click();
await expect(page.getByRole('columnheader', { name: /Name/i }).locator('span.text-blue-600')).toBeVisible();

// Click again to clear
await page.getByRole('columnheader', { name: /Name/i }).click();
await expect(page.getByRole('columnheader', { name: /Name/i }).locator('span.text-gray-400')).toBeVisible();
```

### Pagination & Page Size
```typescript
// Pagination info
await expect(page.getByText(/Showing \d+-\d+ of \d+ items/)).toBeVisible();
await expect(page.getByLabel('Items per page:')).toBeVisible();

// Change page size
const pageSizeSelect = page.getByLabel('Items per page:');
await expect(pageSizeSelect).toHaveValue('10');
await pageSizeSelect.selectOption('5');
await expect(pageSizeSelect).toHaveValue('5');
const rows = page.getByRole('row');
const count = await rows.count();
expect(count).toBeLessThanOrEqual(6); // header + max 5
```

### Data Row Count
Verify table has data (more than just the header row):
```typescript
const rows = page.getByRole('row');
await expect(rows).not.toHaveCount(1);
```

### Active Toggle (Edit Only)
Some entities show an active toggle only on edit, not on create:
```typescript
// Create modal — no toggle
await page.getByRole('button', { name: 'Add Entity' }).click();
await expect(page.getByRole('switch')).not.toBeVisible();

// Edit modal — has toggle
await page.getByRole('row').nth(1).getByRole('button', { name: 'Edit' }).click();
await expect(page.getByRole('switch')).toBeVisible();
```

### Dynamic Rows (e.g., Bank Accounts)
```typescript
// Add rows
await page.getByRole('button', { name: '+ Add Bank Account' }).click();
await expect(page.getByPlaceholder('Account Name')).toHaveCount(1);
await page.getByRole('button', { name: '+ Add Bank Account' }).click();
await expect(page.getByPlaceholder('Account Name')).toHaveCount(2);

// Remove first row
await page.getByRole('button', { name: 'Remove' }).first().click();
await expect(page.getByPlaceholder('Account Name')).toHaveCount(1);
```

### Toast Messages
Toast assertions always use `{ timeout: 10000 }`:
```typescript
await expect(page.getByText('Category created successfully')).toBeVisible({ timeout: 10000 });
await expect(page.getByText('Supplier updated successfully')).toBeVisible({ timeout: 10000 });
await expect(page.getByText(/has been deleted/)).toBeVisible({ timeout: 10000 });
```

### Modal Close Verification
After modal actions, verify it closed:
```typescript
// After successful submit
await expect(page.getByLabel('Name')).not.toBeVisible();
// OR verify by heading
await expect(page.getByRole('heading', { name: 'Create Supplier' })).not.toBeVisible();
```

### Navigation Assertions (Auth Pages)
```typescript
await expect(page).toHaveURL('/dashboard', { timeout: 10000 });
await expect(page).toHaveURL('/login'); // no timeout for current page
```

---

## Conventions

### Timeouts
| Context | Timeout |
|---|---|
| Page load / heading visible | `{ timeout: 10000 }` |
| API operations (create, update, delete toasts) | `{ timeout: 10000 }` |
| Redirects (login → dashboard) | `{ timeout: 10000 }` |
| Search debounce (filter, empty state, reset) | `{ timeout: 5000 }` |
| Immediate UI (modals, buttons, form elements) | No explicit timeout needed |

### Selector Priority (most reliable first)
1. `getByRole('button', { name: '...' })` — buttons, links, headings
2. `getByLabel('...')` — form inputs with labels
3. `getByPlaceholder('...')` — search inputs, textarea placeholders
4. `getByText('...')` — visible text content
5. `getByRole('row')` / `getByRole('cell', { name: '...' })` — table elements
6. `getByRole('columnheader', { name: /regex/i })` — column headers
7. `locator('span.text-blue-600')` / `locator('strong', ...)` — CSS-based (last resort)

### Naming
- Test files: `{page}.spec.ts` (e.g., `category.spec.ts`, `purchase-order.spec.ts`)
- Test data: `Test Category ${Date.now()}`, `Edit Me ${Date.now()}`, `Delete Me ${Date.now()}`
- Describe blocks: page/feature name (e.g., `'Master Category'`, `'Login Page'`)
- Test names: `'should ...'` format

### Test Independence
- Each test must be independent — no relying on state from other tests
- For edit/delete tests, create fresh data within the test itself
- Use `test.describe.serial()` only when order genuinely matters

### Other
- Use `{ exact: true }` on selectors when labels overlap (e.g., `'Password'` vs `'Confirm Password'`)
- Use `.nth(1)` to skip the header row when accessing table data rows
- Use `getByRole('row').nth(1).getByRole('cell').nth(1)` to read first data cell
- Avoid hardcoding IDs — always find elements by visible text/role

---

## Running Tests
```bash
# Docker (recommended)
docker compose --profile test run --rm e2e

# Locally
cd e2e && npm test                               # all tests headless
cd e2e && npx playwright test tests/{feature}/{page}.spec.ts  # single file
cd e2e && npx playwright test -g "test name"     # single test by name
cd e2e && npm run test:headed                    # with visible browser
cd e2e && npm run test:debug                     # step-by-step debugger
cd e2e && npm run report                         # view HTML report
```

## Existing Helpers
- `@helpers/auth` → `login(page, email, password)` — logs in and waits for dashboard redirect

When a setup step is used in 3+ test files, extract it to `e2e/helpers/{name}.ts`.
