import { test, expect } from '@playwright/test';
import { login } from '@helpers/auth';

test.describe('Master Product', () => {
  test.beforeEach(async ({ page }) => {
    await login(page, 'admin@pointofsale.com', 'Admin@12345');
    await page.goto('/master/product');
    await expect(page.getByRole('heading', { name: 'Master Product' })).toBeVisible({ timeout: 10000 });
  });

  test('should display page heading and key elements', async ({ page }) => {
    await expect(page.getByRole('button', { name: 'Add Product' })).toBeVisible();
    await expect(page.getByPlaceholder('Search products...')).toBeVisible();
    // Table headers
    await expect(page.getByRole('columnheader', { name: /ID/i })).toBeVisible();
    await expect(page.getByRole('columnheader', { name: /Image/i })).toBeVisible();
    await expect(page.getByRole('columnheader', { name: /Name/i })).toBeVisible();
    await expect(page.getByRole('columnheader', { name: /Category/i })).toBeVisible();
    await expect(page.getByRole('columnheader', { name: /Status/i })).toBeVisible();
    await expect(page.getByRole('columnheader', { name: /Actions/i })).toBeVisible();
  });

  test('should load products in the table', async ({ page }) => {
    // Wait for table data to load (no more "Loading products..." text)
    await expect(page.getByText('Loading products...')).not.toBeVisible({ timeout: 10000 });
    const rows = page.getByRole('row');
    await expect(rows).not.toHaveCount(1); // more than just the header
  });

  test('should search products', async ({ page }) => {
    // Wait for data to load
    await expect(page.getByText('Loading products...')).not.toBeVisible({ timeout: 10000 });

    // Get the name of the first product to search for
    const firstRow = page.getByRole('row').nth(1);
    const firstProductName = await firstRow.getByRole('cell').nth(2).textContent();

    await page.getByPlaceholder('Search products...').fill(firstProductName!);
    // Wait for debounce + API response — use first() since product name appears in both image and name cells
    await expect(page.getByRole('cell', { name: firstProductName! }).first()).toBeVisible({ timeout: 5000 });
  });

  test('should show empty state when search has no results', async ({ page }) => {
    await page.getByPlaceholder('Search products...').fill('zzz_nonexistent_product_xyz');
    await expect(page.getByText('No data available')).toBeVisible({ timeout: 5000 });
  });

  test('should navigate to add product page', async ({ page }) => {
    await page.getByRole('button', { name: 'Add Product' }).click();
    await expect(page).toHaveURL(/\/master\/product\/add/);
  });

  test('should navigate to edit product page', async ({ page }) => {
    // Wait for data to load
    await expect(page.getByText('Loading products...')).not.toBeVisible({ timeout: 10000 });

    const firstRow = page.getByRole('row').nth(1);
    await firstRow.getByRole('button', { name: 'Edit' }).click();
    await expect(page).toHaveURL(/\/master\/product\/edit\/\d+/);
  });

  test('should open delete confirmation modal', async ({ page }) => {
    // Wait for data to load
    await expect(page.getByText('Loading products...')).not.toBeVisible({ timeout: 10000 });

    const firstRow = page.getByRole('row').nth(1);
    const productName = await firstRow.getByRole('cell').nth(2).textContent();

    await firstRow.getByRole('button', { name: 'Delete' }).click();

    await expect(page.getByRole('heading', { name: 'Delete Product' })).toBeVisible();
    await expect(page.getByText(/Are you sure you want to delete/)).toBeVisible();
    await expect(page.locator('strong', { hasText: productName! })).toBeVisible();
    await expect(page.getByRole('button', { name: 'Cancel' })).toBeVisible();

    // Cancel and verify modal closes
    await page.getByRole('button', { name: 'Cancel' }).click();
    await expect(page.getByRole('heading', { name: 'Delete Product' })).not.toBeVisible();
  });

  test('should sort by name column', async ({ page }) => {
    // Wait for data to load
    await expect(page.getByText('Loading products...')).not.toBeVisible({ timeout: 10000 });

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

  test('should change page size', async ({ page }) => {
    // Wait for data to load
    await expect(page.getByText('Loading products...')).not.toBeVisible({ timeout: 10000 });

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
