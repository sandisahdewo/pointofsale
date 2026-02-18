import { test, expect, Page } from '@playwright/test';
import { login } from '@helpers/auth';

/**
 * Navigate to a role's permissions page via the roles list.
 */
async function goToPermissions(page: Page, roleName: string) {
  await page.goto('/settings/roles');
  await expect(page.getByRole('heading', { name: 'Roles & Permissions' })).toBeVisible({ timeout: 10000 });
  await page.getByPlaceholder('Search roles...').fill(roleName);
  await expect(page.getByRole('cell', { name: roleName, exact: true }).first()).toBeVisible({ timeout: 5000 });
  const row = page.getByRole('row').filter({ hasText: roleName });
  await row.getByRole('button', { name: 'Permissions' }).click();
  await expect(page.getByRole('heading', { name: new RegExp(`Permissions — ${roleName}`) })).toBeVisible({ timeout: 10000 });
}

/**
 * Create a fresh role and navigate to its permissions page.
 * Returns the role name for assertions.
 */
async function createRoleAndGoToPermissions(page: Page): Promise<string> {
  const roleName = `Perm Test ${Date.now()}`;
  await page.goto('/settings/roles');
  await expect(page.getByRole('heading', { name: 'Roles & Permissions' })).toBeVisible({ timeout: 10000 });
  await page.getByRole('button', { name: 'Create Role' }).click();
  await page.getByLabel('Name').fill(roleName);
  await page.getByRole('button', { name: 'Create', exact: true }).click();
  await expect(page.getByText('created successfully')).toBeVisible({ timeout: 10000 });

  await page.getByPlaceholder('Search roles...').fill(roleName);
  await expect(page.getByRole('cell', { name: roleName })).toBeVisible({ timeout: 5000 });
  const row = page.getByRole('row').filter({ hasText: roleName });
  await row.getByRole('button', { name: 'Permissions' }).click();
  await expect(page.getByRole('heading', { name: new RegExp(`Permissions — ${roleName}`) })).toBeVisible({ timeout: 10000 });
  return roleName;
}

test.describe('Settings Permissions', () => {
  test.beforeEach(async ({ page }) => {
    await login(page, 'admin@pointofsale.com', 'Admin@12345');
  });

  // --- Page Load & Display ---

  test('should display page elements for a non-system role', async ({ page }) => {
    await goToPermissions(page, 'Cashier');

    // Back link
    await expect(page.getByRole('link', { name: /Back to Roles/ })).toBeVisible();
    // Save/Cancel buttons visible for non-system role
    await expect(page.getByRole('button', { name: 'Save' })).toBeVisible();
    await expect(page.getByRole('button', { name: 'Cancel' })).toBeVisible();
    // Table header
    await expect(page.getByRole('columnheader', { name: /Module \/ Feature/i })).toBeVisible();
    // Action column headers
    await expect(page.getByRole('columnheader', { name: 'Create' })).toBeVisible();
    await expect(page.getByRole('columnheader', { name: 'Read' })).toBeVisible();
    await expect(page.getByRole('columnheader', { name: 'Update' })).toBeVisible();
    await expect(page.getByRole('columnheader', { name: 'Delete' })).toBeVisible();
  });

  test('should display module groups and features', async ({ page }) => {
    await goToPermissions(page, 'Cashier');

    // Module group names (scoped to table to avoid sidebar matches)
    const table = page.getByRole('table');
    await expect(table.getByRole('button', { name: 'Master Data' })).toBeVisible();
    await expect(table.getByRole('button', { name: 'Transaction' })).toBeVisible();
    await expect(table.getByRole('button', { name: 'Settings' })).toBeVisible();

    // Feature checkboxes (via label)
    await expect(page.getByLabel('Category')).toBeVisible();
    await expect(page.getByLabel('Supplier')).toBeVisible();
    await expect(page.getByLabel('Rack')).toBeVisible();
    await expect(page.getByLabel('Product')).toBeVisible();
    await expect(page.getByLabel('Purchase Order')).toBeVisible();
    await expect(page.getByLabel('Sale')).toBeVisible();
    await expect(page.getByLabel('Stock Adjustment')).toBeVisible();
    await expect(page.getByLabel('Users')).toBeVisible();
    await expect(page.getByLabel('Roles & Permissions')).toBeVisible();
  });

  // --- Existing Permission State ---

  test('should display Cashier existing permissions as checked', async ({ page }) => {
    await goToPermissions(page, 'Cashier');

    // Cashier seed data: Product:read, Sale:create+read
    // Action column order: Create, Read, Update, Delete, Send, Receive
    // Each feature row has checkboxes: [0]=feature, [1]=Create, [2]=Read, [3]=Update, [4]=Delete
    // (Send/Receive only appear as checkboxes for features that support them)

    // Product row — only Read should be checked
    const productRow = page.locator('tr').filter({ has: page.getByLabel('Product') });
    await expect(productRow.getByRole('checkbox').nth(1)).not.toBeChecked(); // Create
    await expect(productRow.getByRole('checkbox').nth(2)).toBeChecked();     // Read
    await expect(productRow.getByRole('checkbox').nth(3)).not.toBeChecked(); // Update
    await expect(productRow.getByRole('checkbox').nth(4)).not.toBeChecked(); // Delete

    // Sale row — Create and Read should be checked
    const saleRow = page.locator('tr').filter({ has: page.getByLabel('Sale') });
    await expect(saleRow.getByRole('checkbox').nth(1)).toBeChecked();        // Create
    await expect(saleRow.getByRole('checkbox').nth(2)).toBeChecked();        // Read
    await expect(saleRow.getByRole('checkbox').nth(3)).not.toBeChecked();    // Update
    await expect(saleRow.getByRole('checkbox').nth(4)).not.toBeChecked();    // Delete

    // Category — all unchecked
    const categoryRow = page.locator('tr').filter({ has: page.getByLabel('Category') });
    await expect(categoryRow.getByRole('checkbox').nth(1)).not.toBeChecked(); // Create
    await expect(categoryRow.getByRole('checkbox').nth(2)).not.toBeChecked(); // Read
    await expect(categoryRow.getByRole('checkbox').nth(3)).not.toBeChecked(); // Update
    await expect(categoryRow.getByRole('checkbox').nth(4)).not.toBeChecked(); // Delete
  });

  // --- Super Admin ---

  test('should show read-only notice for Super Admin', async ({ page }) => {
    await goToPermissions(page, 'Super Admin');

    await expect(
      page.getByText('Super Admin has full access to all features. Permissions cannot be modified.')
    ).toBeVisible();
    // No Save/Cancel buttons for Super Admin
    await expect(page.getByRole('button', { name: 'Save' })).not.toBeVisible();
    await expect(page.getByRole('button', { name: 'Cancel' })).not.toBeVisible();
  });

  test('should have disabled checkboxes for Super Admin', async ({ page }) => {
    await goToPermissions(page, 'Super Admin');

    // Feature checkboxes should be checked and disabled
    await expect(page.getByLabel('Category')).toBeChecked();
    await expect(page.getByLabel('Category')).toBeDisabled();
    await expect(page.getByLabel('Product')).toBeChecked();
    await expect(page.getByLabel('Product')).toBeDisabled();
    await expect(page.getByLabel('Users')).toBeChecked();
    await expect(page.getByLabel('Users')).toBeDisabled();
  });

  // --- Module Collapse/Expand ---

  test('should collapse and expand a module', async ({ page }) => {
    await goToPermissions(page, 'Cashier');

    // Features under Master Data should be visible initially
    await expect(page.getByLabel('Category')).toBeVisible();
    await expect(page.getByLabel('Supplier')).toBeVisible();

    // Collapse Master Data module (scope to table to avoid sidebar)
    const table = page.getByRole('table');
    await table.getByRole('button', { name: 'Master Data' }).click();
    await expect(page.getByLabel('Category')).not.toBeVisible();
    await expect(page.getByLabel('Supplier')).not.toBeVisible();

    // Expand again
    await table.getByRole('button', { name: 'Master Data' }).click();
    await expect(page.getByLabel('Category')).toBeVisible();
    await expect(page.getByLabel('Supplier')).toBeVisible();
  });

  // --- Toggle & Save ---

  test('should toggle a feature checkbox and save permissions', async ({ page }) => {
    test.slow();
    const roleName = await createRoleAndGoToPermissions(page);

    // Fresh role has no permissions — Category should be unchecked
    await expect(page.getByLabel('Category')).not.toBeChecked();

    // Toggle Category feature (checks all actions for Category)
    await page.getByLabel('Category').click();
    await expect(page.getByLabel('Category')).toBeChecked();

    // Save
    await page.getByRole('button', { name: 'Save' }).click();
    await expect(page.getByText('Permissions updated')).toBeVisible({ timeout: 10000 });

    // Should redirect to roles page
    await expect(page.getByRole('heading', { name: 'Roles & Permissions' })).toBeVisible({ timeout: 10000 });
  });

  test('should persist saved permissions after navigating back', async ({ page }) => {
    test.slow();
    const roleName = await createRoleAndGoToPermissions(page);

    // Toggle Category and Supplier
    await page.getByLabel('Category').click();
    await page.getByLabel('Supplier').click();

    // Save
    await page.getByRole('button', { name: 'Save' }).click();
    await expect(page.getByText('Permissions updated')).toBeVisible({ timeout: 10000 });
    await expect(page.getByRole('heading', { name: 'Roles & Permissions' })).toBeVisible({ timeout: 10000 });

    // Navigate back to the same role's permissions
    await goToPermissions(page, roleName);

    // Category and Supplier should still be checked
    await expect(page.getByLabel('Category')).toBeChecked();
    await expect(page.getByLabel('Supplier')).toBeChecked();
    // Rack should still be unchecked
    await expect(page.getByLabel('Rack')).not.toBeChecked();
  });

  // --- Cancel ---

  test('should navigate back on cancel without changes', async ({ page }) => {
    await goToPermissions(page, 'Cashier');

    await page.getByRole('button', { name: 'Cancel' }).click();

    // Should redirect to roles page immediately (no confirm modal)
    await expect(page.getByRole('heading', { name: 'Roles & Permissions' })).toBeVisible({ timeout: 10000 });
  });

  test('should show unsaved changes modal and stay on page', async ({ page }) => {
    test.slow();
    const roleName = await createRoleAndGoToPermissions(page);

    // Make a change
    await page.getByLabel('Category').click();

    // Cancel should show confirm modal
    await page.getByRole('button', { name: 'Cancel' }).click();
    await expect(page.getByRole('heading', { name: 'Unsaved Changes' })).toBeVisible();
    await expect(page.getByText('You have unsaved changes. Are you sure you want to leave?')).toBeVisible();

    // Click Stay — remains on the page
    await page.getByRole('button', { name: 'Stay' }).click();
    await expect(page.getByRole('heading', { name: new RegExp(`Permissions — ${roleName}`) })).toBeVisible();
    // Change should still be applied
    await expect(page.getByLabel('Category')).toBeChecked();
  });

  test('should discard changes and leave on confirm', async ({ page }) => {
    test.slow();
    const roleName = await createRoleAndGoToPermissions(page);

    // Make a change
    await page.getByLabel('Category').click();

    // Cancel → Leave
    await page.getByRole('button', { name: 'Cancel' }).click();
    await expect(page.getByRole('heading', { name: 'Unsaved Changes' })).toBeVisible();
    await page.getByRole('button', { name: 'Leave' }).click();

    // Should redirect to roles page
    await expect(page.getByRole('heading', { name: 'Roles & Permissions' })).toBeVisible({ timeout: 10000 });
  });

  // --- Back to Roles Link ---

  test('should navigate back via Back to Roles link', async ({ page }) => {
    await goToPermissions(page, 'Cashier');

    await page.getByRole('link', { name: /Back to Roles/ }).click();
    await expect(page.getByRole('heading', { name: 'Roles & Permissions' })).toBeVisible({ timeout: 10000 });
  });

  test('should show unsaved changes modal when clicking Back to Roles with changes', async ({ page }) => {
    test.slow();
    const roleName = await createRoleAndGoToPermissions(page);

    // Make a change
    await page.getByLabel('Category').click();

    // Click Back to Roles link — should show confirm modal
    await page.getByRole('link', { name: /Back to Roles/ }).click();
    await expect(page.getByRole('heading', { name: 'Unsaved Changes' })).toBeVisible();
    await expect(page.getByText('You have unsaved changes. Are you sure you want to leave?')).toBeVisible();
  });

  // --- Role Not Found ---

  test('should show not found for invalid role ID', async ({ page }) => {
    await page.goto('/settings/roles/999999/permissions');
    await expect(page.getByRole('heading', { name: 'Role not found' })).toBeVisible({ timeout: 10000 });
    await expect(page.getByText('The role you are looking for does not exist or has been deleted.')).toBeVisible();
    await expect(page.getByRole('link', { name: /Back to Roles/ })).toBeVisible();
  });
});
