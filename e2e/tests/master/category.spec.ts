import { test, expect } from '@playwright/test';
import { login } from '@helpers/auth';

test.describe('Master Category', () => {
  test.beforeEach(async ({ page }) => {
    await login(page, 'admin@pointofsale.com', 'Admin@12345');
    await page.goto('/master/category');
    await expect(page.getByRole('heading', { name: 'Master Category' })).toBeVisible({ timeout: 10000 });
  });

  test('should display page heading and key elements', async ({ page }) => {
    await expect(page.getByRole('button', { name: 'Add Category' })).toBeVisible();
    await expect(page.getByPlaceholder('Search categories...')).toBeVisible();
    // Table headers
    await expect(page.getByRole('columnheader', { name: /ID/i })).toBeVisible();
    await expect(page.getByRole('columnheader', { name: /Name/i })).toBeVisible();
    await expect(page.getByRole('columnheader', { name: /Description/i })).toBeVisible();
    await expect(page.getByRole('columnheader', { name: /Actions/i })).toBeVisible();
  });

  test('should display categories in the table', async ({ page }) => {
    // At least one data row should exist (beyond the header row)
    const rows = page.getByRole('row');
    await expect(rows).not.toHaveCount(1); // more than just the header
  });

  test('should show pagination info', async ({ page }) => {
    await expect(page.getByText(/Showing \d+-\d+ of \d+ items/)).toBeVisible();
    await expect(page.getByLabel('Items per page:')).toBeVisible();
  });

  // --- Search ---

  test('should filter categories by search query', async ({ page }) => {
    // Get the name of the first category to search for
    const firstRow = page.getByRole('row').nth(1);
    const firstCellText = await firstRow.getByRole('cell').nth(1).textContent();

    await page.getByPlaceholder('Search categories...').fill(firstCellText!);
    // Wait for debounce (300ms) + API response
    await expect(page.getByRole('cell', { name: firstCellText! })).toBeVisible({ timeout: 5000 });
  });

  test('should show empty state when search has no results', async ({ page }) => {
    await page.getByPlaceholder('Search categories...').fill('zzz_nonexistent_category_xyz');
    await expect(page.getByText('No data available')).toBeVisible({ timeout: 5000 });
  });

  test('should reset results when search is cleared', async ({ page }) => {
    await page.getByPlaceholder('Search categories...').fill('zzz_nonexistent_xyz');
    await expect(page.getByText('No data available')).toBeVisible({ timeout: 5000 });

    await page.getByPlaceholder('Search categories...').fill('');
    // Data should reappear
    await expect(page.getByText('No data available')).not.toBeVisible({ timeout: 5000 });
  });

  // --- Create ---

  test('should open add category modal', async ({ page }) => {
    await page.getByRole('button', { name: 'Add Category' }).click();
    await expect(page.getByRole('heading', { name: 'Add Category' })).toBeVisible();
    await expect(page.getByLabel('Name')).toBeVisible();
    await expect(page.getByLabel('Description')).toBeVisible();
    await expect(page.getByRole('button', { name: 'Cancel' })).toBeVisible();
    await expect(page.getByRole('button', { name: 'Add', exact: true })).toBeVisible();
  });

  test('should show validation error when submitting empty name', async ({ page }) => {
    await page.getByRole('button', { name: 'Add Category' }).click();
    await page.getByRole('button', { name: 'Add', exact: true }).click();
    // Native HTML5 required validation prevents form submission
    const nameInput = page.getByLabel('Name');
    expect(await nameInput.evaluate((el: HTMLInputElement) => el.validity.valueMissing)).toBe(true);
  });

  test('should clear validation error when typing in name field', async ({ page }) => {
    await page.getByRole('button', { name: 'Add Category' }).click();
    await page.getByRole('button', { name: 'Add', exact: true }).click();
    const nameInput = page.getByLabel('Name');
    expect(await nameInput.evaluate((el: HTMLInputElement) => el.validity.valueMissing)).toBe(true);

    await nameInput.fill('Test');
    expect(await nameInput.evaluate((el: HTMLInputElement) => el.validity.valid)).toBe(true);
  });

  test('should create a new category', async ({ page }) => {
    const categoryName = `Test Category ${Date.now()}`;

    await page.getByRole('button', { name: 'Add Category' }).click();
    await page.getByLabel('Name').fill(categoryName);
    await page.getByLabel('Description').fill('Test description');
    await page.getByRole('button', { name: 'Add', exact: true }).click();

    await expect(page.getByText('Category created successfully')).toBeVisible({ timeout: 10000 });
    // Modal should close
    await expect(page.getByLabel('Name')).not.toBeVisible();
  });

  test('should close add modal on cancel', async ({ page }) => {
    await page.getByRole('button', { name: 'Add Category' }).click();
    await expect(page.getByLabel('Name')).toBeVisible();

    await page.getByRole('button', { name: 'Cancel' }).click();
    await expect(page.getByLabel('Name')).not.toBeVisible();
  });

  // --- Edit ---

  test('should open edit modal with pre-filled data', async ({ page }) => {
    const firstRow = page.getByRole('row').nth(1);
    const categoryName = await firstRow.getByRole('cell').nth(1).textContent();

    await firstRow.getByRole('button', { name: 'Edit' }).click();

    await expect(page.getByRole('heading', { name: 'Edit Category' })).toBeVisible();
    await expect(page.getByLabel('Name')).toHaveValue(categoryName!);
    await expect(page.getByRole('button', { name: 'Update' })).toBeVisible();
  });

  test('should update a category', async ({ page }) => {
    // First create a category to edit
    const categoryName = `Edit Me ${Date.now()}`;
    await page.getByRole('button', { name: 'Add Category' }).click();
    await page.getByLabel('Name').fill(categoryName);
    await page.getByRole('button', { name: 'Add', exact: true }).click();
    await expect(page.getByText('Category created successfully')).toBeVisible({ timeout: 10000 });

    // Search for it and edit
    await page.getByPlaceholder('Search categories...').fill(categoryName);
    await expect(page.getByRole('cell', { name: categoryName })).toBeVisible({ timeout: 5000 });

    await page.getByRole('button', { name: 'Edit' }).click();
    await page.getByLabel('Name').fill(`${categoryName} Updated`);
    await page.getByRole('button', { name: 'Update' }).click();

    await expect(page.getByText('Category updated successfully')).toBeVisible({ timeout: 10000 });
  });

  test('should show validation error when editing name to empty', async ({ page }) => {
    await page.getByRole('row').nth(1).getByRole('button', { name: 'Edit' }).click();
    await page.getByLabel('Name').fill('');
    await page.getByRole('button', { name: 'Update' }).click();

    // Native HTML5 required validation prevents form submission
    const nameInput = page.getByLabel('Name');
    expect(await nameInput.evaluate((el: HTMLInputElement) => el.validity.valueMissing)).toBe(true);
  });

  // --- Delete ---

  test('should open delete confirmation modal', async ({ page }) => {
    const firstRow = page.getByRole('row').nth(1);
    const categoryName = await firstRow.getByRole('cell').nth(1).textContent();

    await firstRow.getByRole('button', { name: 'Delete' }).click();

    await expect(page.getByText('Delete Category')).toBeVisible();
    await expect(page.getByText(/Are you sure you want to delete/)).toBeVisible();
    await expect(page.locator('strong', { hasText: categoryName! })).toBeVisible();
  });

  test('should cancel delete and keep the category', async ({ page }) => {
    const firstRow = page.getByRole('row').nth(1);
    const categoryName = await firstRow.getByRole('cell').nth(1).textContent();

    await firstRow.getByRole('button', { name: 'Delete' }).click();
    await page.getByRole('button', { name: 'Cancel' }).click();

    // Category should still be in the table
    await expect(page.getByRole('cell', { name: categoryName! })).toBeVisible();
  });

  test('should delete a category', async ({ page }) => {
    // Create a category to delete
    const categoryName = `Delete Me ${Date.now()}`;
    await page.getByRole('button', { name: 'Add Category' }).click();
    await page.getByLabel('Name').fill(categoryName);
    await page.getByRole('button', { name: 'Add', exact: true }).click();
    await expect(page.getByText('Category created successfully')).toBeVisible({ timeout: 10000 });

    // Search for it and delete
    await page.getByPlaceholder('Search categories...').fill(categoryName);
    await expect(page.getByRole('cell', { name: categoryName })).toBeVisible({ timeout: 5000 });

    await page.getByRole('button', { name: 'Delete' }).click();
    await expect(page.getByText(/Are you sure you want to delete/)).toBeVisible();

    // Click the Delete button in the confirmation modal
    await page.getByRole('button', { name: 'Delete' }).nth(1).click();

    await expect(page.getByText('Category deleted successfully')).toBeVisible({ timeout: 10000 });
  });

  // --- Sorting ---

  test('should sort by name column', async ({ page }) => {
    // Click Name header to sort ascending
    await page.getByRole('columnheader', { name: /Name/i }).click();
    // Sort icon should change (▲ for asc)
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

    await pageSizeSelect.selectOption('5');
    await expect(pageSizeSelect).toHaveValue('5');
    // Table should refresh with new page size
    const rows = page.getByRole('row');
    // Header + max 5 data rows
    const count = await rows.count();
    expect(count).toBeLessThanOrEqual(6);
  });
});
