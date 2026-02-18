import { test, expect } from '@playwright/test';
import { login } from '@helpers/auth';

test.describe('Settings Roles', () => {
  test.beforeEach(async ({ page }) => {
    await login(page, 'admin@pointofsale.com', 'Admin@12345');
    await page.goto('/settings/roles');
    await expect(page.getByRole('heading', { name: 'Roles & Permissions' })).toBeVisible({ timeout: 10000 });
  });

  // --- Page Load & Display ---

  test('should display page heading and key elements', async ({ page }) => {
    await expect(page.getByRole('button', { name: 'Create Role' })).toBeVisible();
    await expect(page.getByPlaceholder('Search roles...')).toBeVisible();
    // Table headers
    await expect(page.getByRole('columnheader', { name: /ID/i })).toBeVisible();
    await expect(page.getByRole('columnheader', { name: /Name/i })).toBeVisible();
    await expect(page.getByRole('columnheader', { name: /Description/i })).toBeVisible();
    await expect(page.getByRole('columnheader', { name: /Users/i })).toBeVisible();
    await expect(page.getByRole('columnheader', { name: /Actions/i })).toBeVisible();
  });

  test('should display roles in the table', async ({ page }) => {
    const rows = page.getByRole('row');
    await expect(rows).not.toHaveCount(1); // more than just the header
  });

  test('should show pagination info', async ({ page }) => {
    await expect(page.getByText(/Showing \d+-\d+ of \d+ items/)).toBeVisible();
    await expect(page.getByLabel('Items per page:')).toBeVisible();
  });

  // --- Search ---

  test('should filter roles by search query', async ({ page }) => {
    await page.getByPlaceholder('Search roles...').fill('Admin');
    await expect(page.getByRole('cell', { name: /Admin/i }).first()).toBeVisible({ timeout: 5000 });
  });

  test('should show empty state when search has no results', async ({ page }) => {
    await page.getByPlaceholder('Search roles...').fill('zzz_nonexistent_role_xyz');
    await expect(page.getByText('No data available')).toBeVisible({ timeout: 5000 });
  });

  test('should reset results when search is cleared', async ({ page }) => {
    await page.getByPlaceholder('Search roles...').fill('zzz_nonexistent_role_xyz');
    await expect(page.getByText('No data available')).toBeVisible({ timeout: 5000 });

    await page.getByPlaceholder('Search roles...').fill('');
    await expect(page.getByText('No data available')).not.toBeVisible({ timeout: 5000 });
  });

  // --- Create ---

  test('should open create role modal with correct fields', async ({ page }) => {
    await page.getByRole('button', { name: 'Create Role' }).click();
    await expect(page.getByRole('heading', { name: 'Create Role' })).toBeVisible();
    await expect(page.getByLabel('Name')).toBeVisible();
    await expect(page.getByLabel('Description')).toBeVisible();
    await expect(page.getByRole('button', { name: 'Cancel' })).toBeVisible();
    await expect(page.getByRole('button', { name: 'Create', exact: true })).toBeVisible();
  });

  test('should show validation error when name is empty', async ({ page }) => {
    await page.getByRole('button', { name: 'Create Role' }).click();
    await page.getByRole('button', { name: 'Create', exact: true }).click();
    await expect(page.getByText('Name is required')).toBeVisible();
  });

  test('should create a new role', async ({ page }) => {
    const roleName = `Test Role ${Date.now()}`;

    await page.getByRole('button', { name: 'Create Role' }).click();
    await page.getByLabel('Name').fill(roleName);
    await page.getByLabel('Description').fill('A test role for e2e');
    await page.getByRole('button', { name: 'Create', exact: true }).click();

    await expect(page.getByText('created successfully')).toBeVisible({ timeout: 10000 });
    // Modal should close
    await expect(page.getByRole('heading', { name: 'Create Role' })).not.toBeVisible();

    // Verify the role appears in the table
    await page.getByPlaceholder('Search roles...').fill(roleName);
    await expect(page.getByRole('cell', { name: roleName })).toBeVisible({ timeout: 5000 });
  });

  test('should close create modal on cancel', async ({ page }) => {
    await page.getByRole('button', { name: 'Create Role' }).click();
    await expect(page.getByRole('heading', { name: 'Create Role' })).toBeVisible();

    await page.getByRole('button', { name: 'Cancel' }).click();
    await expect(page.getByRole('heading', { name: 'Create Role' })).not.toBeVisible();
  });

  // --- Edit ---

  test('should open edit modal with pre-filled data', async ({ page }) => {
    test.slow();

    // Create a role to edit
    const roleName = `Prefill Me ${Date.now()}`;
    await page.getByRole('button', { name: 'Create Role' }).click();
    await page.getByLabel('Name').fill(roleName);
    await page.getByLabel('Description').fill('Check prefill');
    await page.getByRole('button', { name: 'Create', exact: true }).click();
    await expect(page.getByText('created successfully')).toBeVisible({ timeout: 10000 });

    // Search and open edit
    await page.getByPlaceholder('Search roles...').fill(roleName);
    await expect(page.getByRole('cell', { name: roleName })).toBeVisible({ timeout: 5000 });

    const row = page.getByRole('row').filter({ hasText: roleName });
    await row.getByRole('button', { name: 'Edit' }).click();

    await expect(page.getByRole('heading', { name: 'Edit Role' })).toBeVisible();
    await expect(page.getByLabel('Name')).toHaveValue(roleName);
    await expect(page.getByLabel('Description')).toHaveValue('Check prefill');
  });

  test('should update a role', async ({ page }) => {
    test.slow();

    // Create a role to update
    const roleName = `Edit Me ${Date.now()}`;
    await page.getByRole('button', { name: 'Create Role' }).click();
    await page.getByLabel('Name').fill(roleName);
    await page.getByLabel('Description').fill('Original description');
    await page.getByRole('button', { name: 'Create', exact: true }).click();
    await expect(page.getByText('created successfully')).toBeVisible({ timeout: 10000 });

    // Search and edit
    await page.getByPlaceholder('Search roles...').fill(roleName);
    await expect(page.getByRole('cell', { name: roleName })).toBeVisible({ timeout: 5000 });

    const row = page.getByRole('row').filter({ hasText: roleName });
    await row.getByRole('button', { name: 'Edit' }).click();
    await page.getByLabel('Name').fill(`${roleName} Updated`);
    await page.getByLabel('Description').fill('Updated description');
    await page.getByRole('button', { name: 'Update' }).click();

    await expect(page.getByText('updated successfully')).toBeVisible({ timeout: 10000 });
  });

  test('should show validation error when editing name to empty', async ({ page }) => {
    test.slow();

    // Create a role to edit
    const roleName = `Validate Me ${Date.now()}`;
    await page.getByRole('button', { name: 'Create Role' }).click();
    await page.getByLabel('Name').fill(roleName);
    await page.getByRole('button', { name: 'Create', exact: true }).click();
    await expect(page.getByText('created successfully')).toBeVisible({ timeout: 10000 });

    // Search and open edit
    await page.getByPlaceholder('Search roles...').fill(roleName);
    await expect(page.getByRole('cell', { name: roleName })).toBeVisible({ timeout: 5000 });

    const row = page.getByRole('row').filter({ hasText: roleName });
    await row.getByRole('button', { name: 'Edit' }).click();
    await page.getByLabel('Name').fill('');
    await page.getByRole('button', { name: 'Update' }).click();
    await expect(page.getByText('Name is required')).toBeVisible();
  });

  // --- Delete ---

  test('should open delete confirmation modal', async ({ page }) => {
    test.slow();

    // Create a role to delete
    const roleName = `Delete Modal ${Date.now()}`;
    await page.getByRole('button', { name: 'Create Role' }).click();
    await page.getByLabel('Name').fill(roleName);
    await page.getByRole('button', { name: 'Create', exact: true }).click();
    await expect(page.getByText('created successfully')).toBeVisible({ timeout: 10000 });

    // Search and open delete modal
    await page.getByPlaceholder('Search roles...').fill(roleName);
    await expect(page.getByRole('cell', { name: roleName })).toBeVisible({ timeout: 5000 });

    const row = page.getByRole('row').filter({ hasText: roleName });
    await row.getByRole('button', { name: 'Delete' }).click();

    await expect(page.getByText('Delete Role')).toBeVisible();
    await expect(page.getByText(/Are you sure you want to delete the role/)).toBeVisible();
  });

  test('should cancel delete and keep the role', async ({ page }) => {
    test.slow();

    // Create a role to test cancel
    const roleName = `Cancel Delete ${Date.now()}`;
    await page.getByRole('button', { name: 'Create Role' }).click();
    await page.getByLabel('Name').fill(roleName);
    await page.getByRole('button', { name: 'Create', exact: true }).click();
    await expect(page.getByText('created successfully')).toBeVisible({ timeout: 10000 });

    // Search, open delete, cancel
    await page.getByPlaceholder('Search roles...').fill(roleName);
    await expect(page.getByRole('cell', { name: roleName })).toBeVisible({ timeout: 5000 });

    const row = page.getByRole('row').filter({ hasText: roleName });
    await row.getByRole('button', { name: 'Delete' }).click();
    await expect(page.getByText(/Are you sure you want to delete the role/)).toBeVisible();

    await page.getByRole('button', { name: 'Cancel' }).click();
    // Role should still be visible
    await expect(page.getByRole('cell', { name: roleName })).toBeVisible();
  });

  test('should delete a role', async ({ page }) => {
    test.slow();

    // Create a role to delete
    const roleName = `Delete Me ${Date.now()}`;
    await page.getByRole('button', { name: 'Create Role' }).click();
    await page.getByLabel('Name').fill(roleName);
    await page.getByRole('button', { name: 'Create', exact: true }).click();
    await expect(page.getByText('created successfully')).toBeVisible({ timeout: 10000 });

    // Search and delete
    await page.getByPlaceholder('Search roles...').fill(roleName);
    await expect(page.getByRole('cell', { name: roleName })).toBeVisible({ timeout: 5000 });

    const row = page.getByRole('row').filter({ hasText: roleName });
    await row.getByRole('button', { name: 'Delete' }).click();
    await expect(page.getByText(/Are you sure you want to delete the role/)).toBeVisible();

    // Click Delete in the confirmation modal
    await page.getByRole('button', { name: 'Delete' }).nth(1).click();
    await expect(page.getByText(/has been deleted/)).toBeVisible({ timeout: 10000 });
  });

  // --- Sorting ---

  test('should sort by name column', async ({ page }) => {
    // Click Name header to sort ascending
    await page.getByRole('columnheader', { name: /Name/i }).click();
    await expect(page.getByRole('columnheader', { name: /Name/i }).locator('span.text-blue-600')).toBeVisible();

    // Click again for descending
    await page.getByRole('columnheader', { name: /Name/i }).click();
    await expect(page.getByRole('columnheader', { name: /Name/i }).locator('span.text-blue-600')).toBeVisible();

    // Click again to clear sort
    await page.getByRole('columnheader', { name: /Name/i }).click();
    await expect(page.getByRole('columnheader', { name: /Name/i }).locator('span.text-gray-400')).toBeVisible();
  });

  // --- Page Size ---

  test('should change page size', async ({ page }) => {
    const pageSizeSelect = page.getByLabel('Items per page:');
    await expect(pageSizeSelect).toHaveValue('10');

    const responsePromise = page.waitForResponse(resp => resp.url().includes('/roles') && resp.status() === 200);
    await pageSizeSelect.selectOption('5');
    await responsePromise;
    await expect(pageSizeSelect).toHaveValue('5');
  });

  // --- System Role Protection ---

  test('should not have edit/delete buttons for system roles', async ({ page }) => {
    // Search for "Super Admin" which is a system role
    await page.getByPlaceholder('Search roles...').fill('Super Admin');
    await expect(page.getByRole('cell', { name: 'Super Admin', exact: true }).first()).toBeVisible({ timeout: 5000 });

    const row = page.getByRole('row').filter({ hasText: 'Super Admin' });
    // System roles render Edit/Delete as disabled spans, not buttons
    // The Permissions button should still be present
    await expect(row.getByRole('button', { name: 'Permissions' })).toBeVisible();
    // Edit and Delete should NOT be buttons (they are <span> elements)
    await expect(row.getByRole('button', { name: 'Edit' })).not.toBeVisible();
    await expect(row.getByRole('button', { name: 'Delete' })).not.toBeVisible();
  });

  // --- Permissions Navigation ---

  test('should navigate to permissions page', async ({ page }) => {
    test.slow();

    // Create a role to check permissions navigation
    const roleName = `Perms Nav ${Date.now()}`;
    await page.getByRole('button', { name: 'Create Role' }).click();
    await page.getByLabel('Name').fill(roleName);
    await page.getByRole('button', { name: 'Create', exact: true }).click();
    await expect(page.getByText('created successfully')).toBeVisible({ timeout: 10000 });

    // Search and click Permissions
    await page.getByPlaceholder('Search roles...').fill(roleName);
    await expect(page.getByRole('cell', { name: roleName })).toBeVisible({ timeout: 5000 });

    const row = page.getByRole('row').filter({ hasText: roleName });
    await row.getByRole('button', { name: 'Permissions' }).click();

    // Should navigate to permissions page
    await expect(page.getByRole('heading', { name: new RegExp(`Permissions — ${roleName}`) })).toBeVisible({ timeout: 10000 });
    await expect(page.getByText('Back to Roles')).toBeVisible();
    // Should have Module / Feature column header
    await expect(page.getByRole('columnheader', { name: /Module \/ Feature/i })).toBeVisible();
  });
});
