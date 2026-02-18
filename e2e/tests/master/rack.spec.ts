import { test, expect } from '@playwright/test';
import { login } from '@helpers/auth';

test.describe('Master Rack', () => {
  test.beforeEach(async ({ page }) => {
    await login(page, 'admin@pointofsale.com', 'Admin@12345');
    await page.goto('/master/rack');
    await expect(page.getByRole('heading', { name: 'Master Rack' })).toBeVisible({ timeout: 10000 });
  });

  test('should display page heading and key elements', async ({ page }) => {
    await expect(page.getByRole('button', { name: 'Add Rack' })).toBeVisible();
    await expect(page.getByPlaceholder('Search racks...')).toBeVisible();
    // Table headers
    await expect(page.getByRole('columnheader', { name: /ID/i })).toBeVisible();
    await expect(page.getByRole('columnheader', { name: /Name/i })).toBeVisible();
    await expect(page.getByRole('columnheader', { name: /Code/i })).toBeVisible();
    await expect(page.getByRole('columnheader', { name: /Location/i })).toBeVisible();
    await expect(page.getByRole('columnheader', { name: /Capacity/i })).toBeVisible();
    await expect(page.getByRole('columnheader', { name: /Status/i })).toBeVisible();
    await expect(page.getByRole('columnheader', { name: /Actions/i })).toBeVisible();
  });

  test('should display racks in the table', async ({ page }) => {
    const rows = page.getByRole('row');
    await expect(rows).not.toHaveCount(1); // more than just the header
  });

  test('should show pagination info', async ({ page }) => {
    await expect(page.getByText(/Showing \d+-\d+ of \d+ items/)).toBeVisible();
    await expect(page.getByLabel('Items per page:')).toBeVisible();
  });

  // --- Search ---

  test('should filter racks by search query', async ({ page }) => {
    const firstRow = page.getByRole('row').nth(1);
    const firstCellText = await firstRow.getByRole('cell').nth(1).textContent();

    await page.getByPlaceholder('Search racks...').fill(firstCellText!);
    await expect(page.getByRole('cell', { name: firstCellText! })).toBeVisible({ timeout: 5000 });
  });

  test('should show empty state when search has no results', async ({ page }) => {
    await page.getByPlaceholder('Search racks...').fill('zzz_nonexistent_rack_xyz');
    await expect(page.getByText('No data available')).toBeVisible({ timeout: 5000 });
  });

  test('should reset results when search is cleared', async ({ page }) => {
    await page.getByPlaceholder('Search racks...').fill('zzz_nonexistent_xyz');
    await expect(page.getByText('No data available')).toBeVisible({ timeout: 5000 });

    await page.getByPlaceholder('Search racks...').fill('');
    await expect(page.getByText('No data available')).not.toBeVisible({ timeout: 5000 });
  });

  // --- Create ---

  test('should open add rack modal', async ({ page }) => {
    await page.getByRole('button', { name: 'Add Rack' }).click();
    await expect(page.getByRole('heading', { name: 'Create Rack' })).toBeVisible();
    await expect(page.getByLabel('Name')).toBeVisible();
    await expect(page.getByLabel('Code')).toBeVisible();
    await expect(page.getByLabel('Location')).toBeVisible();
    await expect(page.getByLabel('Capacity')).toBeVisible();
    await expect(page.getByPlaceholder('Optional description')).toBeVisible();
    await expect(page.getByRole('button', { name: 'Cancel' })).toBeVisible();
    await expect(page.getByRole('button', { name: 'Create' })).toBeVisible();
  });

  test('should not show active toggle on create modal', async ({ page }) => {
    await page.getByRole('button', { name: 'Add Rack' }).click();
    await expect(page.getByRole('heading', { name: 'Create Rack' })).toBeVisible();
    await expect(page.getByRole('switch')).not.toBeVisible();
  });

  test('should show validation error when submitting empty name', async ({ page }) => {
    await page.getByRole('button', { name: 'Add Rack' }).click();
    await page.getByRole('button', { name: 'Create' }).click();

    const nameInput = page.getByLabel('Name');
    expect(await nameInput.evaluate((el: HTMLInputElement) => el.validity.valueMissing)).toBe(true);
  });

  test('should create a new rack', async ({ page }) => {
    const rackName = `Test Rack ${Date.now()}`;
    const rackCode = `TR-${Date.now()}`;

    await page.getByRole('button', { name: 'Add Rack' }).click();
    await page.getByLabel('Name').fill(rackName);
    await page.getByLabel('Code').fill(rackCode);
    await page.getByLabel('Location').fill('Warehouse A');
    await page.getByLabel('Capacity').fill('50');
    await page.getByRole('button', { name: 'Create' }).click();

    await expect(page.getByText('Rack created successfully')).toBeVisible({ timeout: 10000 });
    // Modal should close
    await expect(page.getByRole('heading', { name: 'Create Rack' })).not.toBeVisible();
  });

  test('should close add modal on cancel', async ({ page }) => {
    await page.getByRole('button', { name: 'Add Rack' }).click();
    await expect(page.getByRole('heading', { name: 'Create Rack' })).toBeVisible();

    await page.getByRole('button', { name: 'Cancel' }).click();
    await expect(page.getByRole('heading', { name: 'Create Rack' })).not.toBeVisible();
  });

  // --- Edit ---

  test('should open edit modal with pre-filled data', async ({ page }) => {
    const firstRow = page.getByRole('row').nth(1);
    const rackName = await firstRow.getByRole('cell').nth(1).textContent();

    await firstRow.getByRole('button', { name: 'Edit' }).click();

    await expect(page.getByRole('heading', { name: 'Edit Rack' })).toBeVisible();
    await expect(page.getByLabel('Name')).toHaveValue(rackName!);
    await expect(page.getByRole('button', { name: 'Update' })).toBeVisible();
  });

  test('should show active toggle on edit modal', async ({ page }) => {
    await page.getByRole('row').nth(1).getByRole('button', { name: 'Edit' }).click();
    await expect(page.getByRole('heading', { name: 'Edit Rack' })).toBeVisible();
    await expect(page.getByRole('switch')).toBeVisible();
  });

  test('should update a rack', async ({ page }) => {
    // Create a rack to edit
    const rackName = `Edit Me ${Date.now()}`;
    const rackCode = `EM-${Date.now()}`;
    await page.getByRole('button', { name: 'Add Rack' }).click();
    await page.getByLabel('Name').fill(rackName);
    await page.getByLabel('Code').fill(rackCode);
    await page.getByLabel('Location').fill('Edit Location');
    await page.getByLabel('Capacity').fill('20');
    await page.getByRole('button', { name: 'Create' }).click();
    await expect(page.getByText('Rack created successfully')).toBeVisible({ timeout: 10000 });

    // Search for it and edit
    await page.getByPlaceholder('Search racks...').fill(rackName);
    await expect(page.getByRole('cell', { name: rackName })).toBeVisible({ timeout: 5000 });

    await page.getByRole('button', { name: 'Edit' }).click();
    await page.getByLabel('Name').fill(`${rackName} Updated`);
    await page.getByRole('button', { name: 'Update' }).click();

    await expect(page.getByText('Rack updated successfully')).toBeVisible({ timeout: 10000 });
  });

  test('should show validation error when editing name to empty', async ({ page }) => {
    await page.getByRole('row').nth(1).getByRole('button', { name: 'Edit' }).click();
    await page.getByLabel('Name').fill('');
    await page.getByRole('button', { name: 'Update' }).click();

    const nameInput = page.getByLabel('Name');
    expect(await nameInput.evaluate((el: HTMLInputElement) => el.validity.valueMissing)).toBe(true);
  });

  // --- Delete ---

  test('should open delete confirmation modal', async ({ page }) => {
    const firstRow = page.getByRole('row').nth(1);

    await firstRow.getByRole('button', { name: 'Delete' }).click();

    await expect(page.getByRole('heading', { name: 'Delete Rack' })).toBeVisible();
    await expect(page.getByText(/Are you sure you want to delete rack/)).toBeVisible();
  });

  test('should cancel delete and keep the rack', async ({ page }) => {
    const firstRow = page.getByRole('row').nth(1);
    const rackName = await firstRow.getByRole('cell').nth(1).textContent();

    await firstRow.getByRole('button', { name: 'Delete' }).click();
    await page.getByRole('button', { name: 'Cancel' }).click();

    // Rack should still be in the table
    await expect(page.getByRole('cell', { name: rackName! })).toBeVisible();
  });

  test('should delete a rack', async ({ page }) => {
    // Create a rack to delete
    const rackName = `Delete Me ${Date.now()}`;
    const rackCode = `DM-${Date.now()}`;
    await page.getByRole('button', { name: 'Add Rack' }).click();
    await page.getByLabel('Name').fill(rackName);
    await page.getByLabel('Code').fill(rackCode);
    await page.getByLabel('Location').fill('Delete Location');
    await page.getByLabel('Capacity').fill('5');
    await page.getByRole('button', { name: 'Create' }).click();
    await expect(page.getByText('Rack created successfully')).toBeVisible({ timeout: 10000 });

    // Search for it and delete
    await page.getByPlaceholder('Search racks...').fill(rackName);
    await expect(page.getByRole('cell', { name: rackName })).toBeVisible({ timeout: 5000 });

    await page.getByRole('button', { name: 'Delete' }).click();
    await expect(page.getByText(/Are you sure you want to delete rack/)).toBeVisible();

    // Click the Delete button in the confirmation modal
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

  test('should sort by status column', async ({ page }) => {
    await page.getByRole('columnheader', { name: /Status/i }).click();
    await expect(page.getByRole('columnheader', { name: /Status/i }).locator('span.text-blue-600')).toBeVisible();
  });

  // --- Page Size ---

  test('should change page size', async ({ page }) => {
    const pageSizeSelect = page.getByLabel('Items per page:');
    await expect(pageSizeSelect).toHaveValue('10');

    await pageSizeSelect.selectOption('5');
    await expect(pageSizeSelect).toHaveValue('5');
    // Table should refresh with new page size
    const rows = page.getByRole('row');
    // Header + max 5 data rows
    const count = await rows.count();
    expect(count).toBeLessThanOrEqual(6);
  });
});
