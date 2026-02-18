# Skill: E2E Test

Creates, edits, and fixes Playwright end-to-end tests for the Point of Sale admin panel.

## When to Use
- Creating new e2e tests for a page or feature
- Fixing failing e2e tests
- Editing or modifying existing e2e tests

---

## Workflow: Creating Tests

### Step 1: Read the source code

**Read the page and component source code first.** This is the fastest and most accurate way to discover every element name, label, toast message, and behavior you need to test.

Read these files for the target page:
1. **Page file** (`src/app/{feature}/page.tsx`) — column definitions, button labels, modal titles, toast messages, search placeholder, action handlers
2. **Form/Modal component** (`src/components/{entity}/...`) — field labels, placeholders, required attributes, validation logic, submit handler
3. **Store** (`src/stores/use{Entity}Store.ts`) — API call signatures, data types, field names

From these files, extract:
- Heading text, button labels (Create, Save, Update, Delete, Cancel)
- Modal titles (e.g., `isEdit ? 'Edit User' : 'Create User'`)
- Form field labels and which are required
- Toast messages (exact strings from `addToast(...)` calls)
- Table column headers (from `columns` array)
- Search placeholder text
- Validation logic (what triggers errors)
- Error handling (silent `try/catch` that might swallow errors)

### Step 2: Read existing patterns

1. Read existing helpers in `e2e/helpers/`
2. Read 1-2 existing tests in the same feature area for patterns

### Optional: Use agent-browser for verification

Use **agent-browser** only when you need to verify runtime behavior that isn't obvious from source code:
- Dynamic content that depends on database state
- CSS-based selectors (sort indicator colors, status badges)
- Complex multi-step flows where you're unsure of the intermediate states
- Debugging a failing test to see what the page actually shows

### Step 3: Write the test file

File path: `e2e/tests/{feature}/{page}.spec.ts`

Write tests in this order:
1. Page load & display (heading, buttons, table headers)
2. Data display (rows exist)
3. Pagination info
4. Search (filter, empty state, reset)
5. Create (open modal, validation, successful create, cancel)
6. Edit (prefill check, update, validation)
7. Delete (confirmation, cancel, confirm)
8. Sorting
9. Page size
10. Feature-specific (toggles, status badges, approval flows)

### Step 4: Run and fix

```bash
cd e2e && npx playwright test tests/{feature}/{page}.spec.ts --reporter=list
```

**Max 3 fix-run iterations.** If still failing after 3 rounds, switch to the debugging workflow below.

---

## Workflow: Fixing Failing Tests (Fast Path)

### Step 1: Read error + screenshot + error context

```bash
cd e2e && npx playwright test -g "test name" --reporter=list
```

Read the error output, screenshot, AND `error-context.md` in test-results.

### Step 2: Classify the failure and act

| Failure Type | Signal | Fix |
|---|---|---|
| **Element not found / timeout** | `waiting for getByRole(...)` | Selector mismatch → browse with agent-browser to check actual labels |
| **Strict mode violation** | `resolved to N elements` | Add `{ exact: true }`, `.first()`, or scope to row (see patterns) |
| **Click works but nothing happens** | Modal stays open, no toast, no API call | **This is a JS error, NOT a click problem** → read component source code (see Step 3) |
| **Assertion mismatch** | `expected X, received Y` | Check actual value from screenshot/context |
| **Toast not visible** | `waiting for getByText(...)` after API action | Increase timeout, or check if API is failing |
| **Race condition** | Flaky pass/fail | Add `waitForResponse` or explicit waits |

### Step 3: When form submit doesn't trigger (CRITICAL)

If clicking a submit button works (element found, click succeeds) but the form doesn't submit (no toast, no API call, modal stays open):

**DO NOT** try different click methods (JS click, dispatchEvent, keyboard Enter, viewport resize). The click works fine.

**DO THIS instead:**
1. **Read the component source code** (the form component, not the page)
2. Look for `try/catch` blocks that silently swallow errors
3. Check if any `.trim()`, `.map()`, or property access runs on potentially `null/undefined` values
4. Check browser console for React warnings ("controlled to uncontrolled" = null state from API)
5. The fix is almost always in the **frontend code**, not the test

Common root causes:
- API returns `null` for optional fields → component does `setPhone(user.phone)` → state is `null` → `null.trim()` throws in submit handler
- Silent `catch(error) {}` blocks that swallow TypeErrors
- Fix: `setField(value || '')` in the component's useEffect

### Step 4: Max iterations

- **3 fix-run cycles max** per failure
- If still stuck: read the component source code, check browser console, inspect API responses
- Never try 5+ variations of the same approach (click methods, viewport sizes, scroll strategies)

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

### Unique Test Data
```typescript
const categoryName = `Test Category ${Date.now()}`;
const userEmail = `testuser${Date.now()}@example.com`;
```

### HTML5 Native Validation
```typescript
await page.getByRole('button', { name: 'Save' }).click();
const nameInput = page.getByLabel('Name');
expect(await nameInput.evaluate((el: HTMLInputElement) => el.validity.valueMissing)).toBe(true);
```

### Custom Validation Errors (Auth Pages)
```typescript
await page.getByRole('button', { name: 'Login' }).click();
await expect(page.getByText('Email is required')).toBeVisible();
```

### Create-Search-Act Flow
For edit/delete, create fresh data first to avoid depending on pre-existing state:
```typescript
test('should update an entity', async ({ page }) => {
  test.slow(); // multi-step tests need extra time

  // 1. Create
  const name = `Edit Me ${Date.now()}`;
  await page.getByRole('button', { name: 'Create' }).click();
  await page.getByLabel('Name').fill(name);
  await page.getByRole('button', { name: 'Save' }).click();
  await expect(page.getByText('created successfully')).toBeVisible({ timeout: 10000 });

  // 2. Search
  await page.getByPlaceholder('Search...').fill(name);
  await expect(page.getByRole('cell', { name })).toBeVisible({ timeout: 5000 });

  // 3. Act
  await page.getByRole('button', { name: 'Edit' }).click();
  await page.getByLabel('Name').fill(`${name} Updated`);
  await page.getByRole('button', { name: 'Save' }).click();
  await expect(page.getByText('updated successfully')).toBeVisible({ timeout: 10000 });
});
```

### Self-Contained State Tests
When testing features that require specific state (e.g., pending users), create that state within the test:
```typescript
test('should approve a pending user', async ({ page }) => {
  test.slow();

  // Create prerequisite state: register a new user
  await page.goto('/register');
  const userName = `Approve Me ${Date.now()}`;
  await page.getByLabel('Name').fill(userName);
  await page.getByLabel('Email').fill(`approve${Date.now()}@example.com`);
  await page.getByLabel('Password', { exact: true }).fill('Test@12345');
  await page.getByLabel('Confirm Password').fill('Test@12345');
  await page.getByRole('button', { name: 'Register' }).click();
  await expect(page.getByText(/registration successful/i)).toBeVisible({ timeout: 10000 });

  // Re-login as admin and act
  await login(page, 'admin@pointofsale.com', 'Admin@12345');
  await page.goto('/settings/users');
  await page.getByPlaceholder('Search users...').fill(userName);
  await expect(page.getByRole('cell', { name: userName })).toBeVisible({ timeout: 5000 });

  const row = page.getByRole('row').filter({ hasText: userName });
  await row.getByRole('button', { name: 'Approve' }).click();
  await expect(page.getByText(/has been approved/)).toBeVisible({ timeout: 10000 });
});
```

### Delete Confirmation
Use `.nth(1)` for the modal's Delete button (row's Delete is `.nth(0)`):
```typescript
await page.getByRole('button', { name: 'Delete' }).click();
await expect(page.getByText(/Are you sure you want to delete/)).toBeVisible();
await page.getByRole('button', { name: 'Delete' }).nth(1).click();
await expect(page.getByText(/deleted successfully/)).toBeVisible({ timeout: 10000 });
```

### Search with Debounce
```typescript
// Filter
await page.getByPlaceholder('Search...').fill('term');
await expect(page.getByRole('cell', { name: 'Match' })).toBeVisible({ timeout: 5000 });

// Empty state
await page.getByPlaceholder('Search...').fill('zzz_nonexistent_xyz');
await expect(page.getByText('No data available')).toBeVisible({ timeout: 5000 });

// Reset
await page.getByPlaceholder('Search...').fill('');
await expect(page.getByText('No data available')).not.toBeVisible({ timeout: 5000 });
```

### Column Sorting
```typescript
await page.getByRole('columnheader', { name: /Name/i }).click();
await expect(page.getByRole('columnheader', { name: /Name/i }).locator('span.text-blue-600')).toBeVisible();

await page.getByRole('columnheader', { name: /Name/i }).click(); // desc
await expect(page.getByRole('columnheader', { name: /Name/i }).locator('span.text-blue-600')).toBeVisible();

await page.getByRole('columnheader', { name: /Name/i }).click(); // clear
await expect(page.getByRole('columnheader', { name: /Name/i }).locator('span.text-gray-400')).toBeVisible();
```

### Page Size with waitForResponse
**Always wait for the API response** after changing page size to avoid race conditions:
```typescript
const pageSizeSelect = page.getByLabel('Items per page:');
await expect(pageSizeSelect).toHaveValue('10');

const responsePromise = page.waitForResponse(resp => resp.url().includes('/entity') && resp.status() === 200);
await pageSizeSelect.selectOption('5');
await responsePromise;
await expect(pageSizeSelect).toHaveValue('5');
await expect(page.getByRole('row')).toHaveCount(6, { timeout: 5000 }); // header + 5
```

### Toast Messages
Always `{ timeout: 10000 }`:
```typescript
await expect(page.getByText('created successfully')).toBeVisible({ timeout: 10000 });
await expect(page.getByText('updated successfully')).toBeVisible({ timeout: 10000 });
await expect(page.getByText(/has been deleted/)).toBeVisible({ timeout: 10000 });
```

### Data Row Count
```typescript
await expect(page.getByRole('row')).not.toHaveCount(1); // more than header
```

### Active Toggle / Status Field (Edit Only)
```typescript
// Create modal — no status
await page.getByRole('button', { name: 'Create' }).click();
await expect(page.getByLabel('Status')).not.toBeVisible();
await page.getByRole('button', { name: 'Cancel' }).click();

// Edit modal — has status
await page.getByRole('row').nth(1).getByRole('button', { name: 'Edit' }).click();
await expect(page.getByLabel('Status')).toBeVisible();
```

---

## Preventing Common Failures

### Strict Mode Violations
When a selector matches multiple elements:

```typescript
// BAD — "Admin" might match multiple cells
await expect(page.getByRole('cell', { name: cellText })).toBeVisible();

// GOOD — use exact + first
await expect(page.getByRole('cell', { name: cellText, exact: true }).first()).toBeVisible();
```

Scope actions to a specific row:
```typescript
// BAD — "Edit" button exists in every row
await page.getByRole('button', { name: 'Edit' }).click();

// GOOD — scope to the row containing our data
const row = page.getByRole('row').filter({ hasText: userName });
await row.getByRole('button', { name: 'Edit' }).click();
```

### Multi-Step Tests
Use `test.slow()` for tests that create data before acting on it:
```typescript
test('should update a user', async ({ page }) => {
  test.slow(); // 3x default timeout
  // create → search → edit → assert
});
```

### Overlapping Labels
Use `{ exact: true }` when labels are substrings of each other:
```typescript
// BAD — matches both "Password" and "Confirm Password"
await page.getByLabel('Password').fill('test');

// GOOD
await page.getByLabel('Password', { exact: true }).fill('test');
await page.getByLabel('Confirm Password').fill('test');
```

### Reading Cell Data (column index varies by page)
```typescript
// ID at index 0, Name at index 1 (category), Name at index 2 (users — after ID and Profile)
const firstRow = page.getByRole('row').nth(1);
const name = await firstRow.getByRole('cell').nth(1).textContent(); // adjust index per page
```

---

## Conventions

### Timeouts
| Context | Timeout |
|---|---|
| Page load / heading visible | `{ timeout: 10000 }` |
| API operations (create/update/delete toasts) | `{ timeout: 10000 }` |
| Redirects (login → dashboard) | `{ timeout: 10000 }` |
| Search debounce | `{ timeout: 5000 }` |
| Page size row count | `{ timeout: 5000 }` |
| Immediate UI (modals, buttons) | No explicit timeout |

### Selector Priority
1. `getByRole('button', { name })` — buttons, links, headings
2. `getByLabel('...')` — form inputs with labels
3. `getByPlaceholder('...')` — search inputs
4. `getByText('...')` — visible text
5. `getByRole('row')` / `getByRole('cell', { name })` — table elements
6. `getByRole('columnheader', { name: /regex/i })` — column headers
7. `page.getByRole('row').filter({ hasText })` — scope to specific row
8. `locator('span.text-blue-600')` — CSS-based (last resort)

### Naming
- Files: `{page}.spec.ts` (e.g., `category.spec.ts`, `users.spec.ts`)
- Data: `Test Entity ${Date.now()}`, `Edit Me ${Date.now()}`, `Delete Me ${Date.now()}`
- Describe: page name (e.g., `'Master Category'`, `'Settings Users'`)
- Tests: `'should ...'` format

### Test Independence
- Each test is independent — no relying on state from other tests
- For edit/delete: create fresh data within the test (Create-Search-Act)
- For state-dependent features: create prerequisite state within the test

---

## Running Tests
```bash
cd e2e && npm test                                                    # all
cd e2e && npx playwright test tests/{feature}/{page}.spec.ts          # single file
cd e2e && npx playwright test -g "test name"                          # single test
cd e2e && npm run test:headed                                         # visible browser
cd e2e && npm run test:debug                                          # step debugger
docker compose --profile test run --rm e2e                            # Docker
```

## Existing Helpers
- `@helpers/auth` → `login(page, email, password)` — logs in and waits for dashboard redirect

When a setup step is used in 3+ test files, extract it to `e2e/helpers/{name}.ts`.
