import { test, expect } from '@playwright/test';
import { login } from '@helpers/auth';

test.describe('Master Supplier', () => {
  test.beforeEach(async ({ page }) => {
    await login(page, 'admin@pointofsale.com', 'Admin@12345');
    await page.goto('/master/supplier');
    await expect(page.getByRole('heading', { name: 'Master Supplier' })).toBeVisible({ timeout: 10000 });
  });

  test('should display page heading and key elements', async ({ page }) => {
    await expect(page.getByRole('button', { name: 'Add Supplier' })).toBeVisible();
    await expect(page.getByPlaceholder('Search suppliers...')).toBeVisible();
    // Table headers
    await expect(page.getByRole('columnheader', { name: /ID/i })).toBeVisible();
    await expect(page.getByRole('columnheader', { name: /Name/i })).toBeVisible();
    await expect(page.getByRole('columnheader', { name: /Address/i })).toBeVisible();
    await expect(page.getByRole('columnheader', { name: /Phone/i })).toBeVisible();
    await expect(page.getByRole('columnheader', { name: /Email/i })).toBeVisible();
    await expect(page.getByRole('columnheader', { name: /Status/i })).toBeVisible();
    await expect(page.getByRole('columnheader', { name: /Actions/i })).toBeVisible();
  });

  test('should display suppliers in the table', async ({ page }) => {
    const rows = page.getByRole('row');
    await expect(rows).not.toHaveCount(1); // more than just the header
  });

  test('should show pagination info', async ({ page }) => {
    await expect(page.getByText(/Showing \d+-\d+ of \d+ items/)).toBeVisible();
    await expect(page.getByLabel('Items per page:')).toBeVisible();
  });

  // --- Search ---

  test('should filter suppliers by search query', async ({ page }) => {
    const firstRow = page.getByRole('row').nth(1);
    const firstCellText = await firstRow.getByRole('cell').nth(1).textContent();

    await page.getByPlaceholder('Search suppliers...').fill(firstCellText!);
    await expect(page.getByRole('cell', { name: firstCellText! })).toBeVisible({ timeout: 5000 });
  });

  test('should show empty state when search has no results', async ({ page }) => {
    await page.getByPlaceholder('Search suppliers...').fill('zzz_nonexistent_supplier_xyz');
    await expect(page.getByText('No data available')).toBeVisible({ timeout: 5000 });
  });

  test('should reset results when search is cleared', async ({ page }) => {
    await page.getByPlaceholder('Search suppliers...').fill('zzz_nonexistent_xyz');
    await expect(page.getByText('No data available')).toBeVisible({ timeout: 5000 });

    await page.getByPlaceholder('Search suppliers...').fill('');
    await expect(page.getByText('No data available')).not.toBeVisible({ timeout: 5000 });
  });

  // --- Create ---

  test('should open add supplier modal', async ({ page }) => {
    await page.getByRole('button', { name: 'Add Supplier' }).click();
    await expect(page.getByRole('heading', { name: 'Create Supplier' })).toBeVisible();
    await expect(page.getByLabel('Name')).toBeVisible();
    await expect(page.getByLabel('Address')).toBeVisible();
    await expect(page.getByLabel('Phone')).toBeVisible();
    await expect(page.getByLabel('Email')).toBeVisible();
    await expect(page.getByLabel('Website')).toBeVisible();
    await expect(page.getByText('Bank Accounts (optional)')).toBeVisible();
    await expect(page.getByRole('button', { name: 'Cancel' })).toBeVisible();
    await expect(page.getByRole('button', { name: 'Create' })).toBeVisible();
  });

  test('should not show active toggle on create modal', async ({ page }) => {
    await page.getByRole('button', { name: 'Add Supplier' }).click();
    await expect(page.getByRole('heading', { name: 'Create Supplier' })).toBeVisible();
    // Active toggle (switch) should only appear on edit, not create
    await expect(page.getByRole('switch')).not.toBeVisible();
  });

  test('should show validation error when submitting empty name', async ({ page }) => {
    await page.getByRole('button', { name: 'Add Supplier' }).click();
    // Fill address to isolate name validation
    await page.getByPlaceholder('Supplier address').fill('Some address');
    await page.getByRole('button', { name: 'Create' }).click();

    const nameInput = page.getByLabel('Name');
    expect(await nameInput.evaluate((el: HTMLInputElement) => el.validity.valueMissing)).toBe(true);
  });

  test('should show validation error when submitting empty address', async ({ page }) => {
    await page.getByRole('button', { name: 'Add Supplier' }).click();
    await page.getByLabel('Name').fill('Test Supplier');
    // Clear address and submit
    await page.getByPlaceholder('Supplier address').fill('');
    await page.getByRole('button', { name: 'Create' }).click();

    const addressTextarea = page.getByPlaceholder('Supplier address');
    expect(await addressTextarea.evaluate((el: HTMLTextAreaElement) => el.validity.valueMissing)).toBe(true);
  });

  test('should create a new supplier', async ({ page }) => {
    const supplierName = `Test Supplier ${Date.now()}`;

    await page.getByRole('button', { name: 'Add Supplier' }).click();
    await page.getByLabel('Name').fill(supplierName);
    await page.getByPlaceholder('Supplier address').fill('123 Test Street');
    await page.getByLabel('Phone').fill('08123456789');
    await page.getByLabel('Email').fill('test@supplier.com');
    await page.getByRole('button', { name: 'Create' }).click();

    await expect(page.getByText('Supplier created successfully')).toBeVisible({ timeout: 10000 });
    // Modal should close
    await expect(page.getByRole('heading', { name: 'Create Supplier' })).not.toBeVisible();
  });

  test('should create a supplier with bank accounts', async ({ page }) => {
    const supplierName = `Bank Supplier ${Date.now()}`;

    await page.getByRole('button', { name: 'Add Supplier' }).click();
    await page.getByLabel('Name').fill(supplierName);
    await page.getByPlaceholder('Supplier address').fill('456 Bank Street');

    // Add a bank account
    await page.getByRole('button', { name: '+ Add Bank Account' }).click();
    await page.getByPlaceholder('Account Name').fill('BCA');
    await page.getByPlaceholder('Account Number').fill('1234567890');

    await page.getByRole('button', { name: 'Create' }).click();
    await expect(page.getByText('Supplier created successfully')).toBeVisible({ timeout: 10000 });
  });

  test('should add and remove bank account rows', async ({ page }) => {
    await page.getByRole('button', { name: 'Add Supplier' }).click();

    // Add two bank accounts
    await page.getByRole('button', { name: '+ Add Bank Account' }).click();
    await expect(page.getByPlaceholder('Account Name')).toHaveCount(1);

    await page.getByRole('button', { name: '+ Add Bank Account' }).click();
    await expect(page.getByPlaceholder('Account Name')).toHaveCount(2);

    // Remove the first one
    await page.getByRole('button', { name: 'Remove' }).first().click();
    await expect(page.getByPlaceholder('Account Name')).toHaveCount(1);
  });

  test('should close add modal on cancel', async ({ page }) => {
    await page.getByRole('button', { name: 'Add Supplier' }).click();
    await expect(page.getByRole('heading', { name: 'Create Supplier' })).toBeVisible();

    await page.getByRole('button', { name: 'Cancel' }).click();
    await expect(page.getByRole('heading', { name: 'Create Supplier' })).not.toBeVisible();
  });

  // --- Edit ---

  test('should open edit modal with pre-filled data', async ({ page }) => {
    const firstRow = page.getByRole('row').nth(1);
    const supplierName = await firstRow.getByRole('cell').nth(1).textContent();

    await firstRow.getByRole('button', { name: 'Edit' }).click();

    await expect(page.getByRole('heading', { name: 'Edit Supplier' })).toBeVisible();
    await expect(page.getByLabel('Name')).toHaveValue(supplierName!);
    await expect(page.getByRole('button', { name: 'Update' })).toBeVisible();
  });

  test('should show active toggle on edit modal', async ({ page }) => {
    await page.getByRole('row').nth(1).getByRole('button', { name: 'Edit' }).click();
    await expect(page.getByRole('heading', { name: 'Edit Supplier' })).toBeVisible();
    // Active toggle renders as a switch element
    await expect(page.getByRole('switch')).toBeVisible();
  });

  test('should update a supplier', async ({ page }) => {
    // Create a supplier to edit
    const supplierName = `Edit Me ${Date.now()}`;
    await page.getByRole('button', { name: 'Add Supplier' }).click();
    await page.getByLabel('Name').fill(supplierName);
    await page.getByPlaceholder('Supplier address').fill('Edit Street');
    await page.getByRole('button', { name: 'Create' }).click();
    await expect(page.getByText('Supplier created successfully')).toBeVisible({ timeout: 10000 });

    // Search for it and edit
    await page.getByPlaceholder('Search suppliers...').fill(supplierName);
    await expect(page.getByRole('cell', { name: supplierName })).toBeVisible({ timeout: 5000 });

    await page.getByRole('button', { name: 'Edit' }).click();
    await page.getByLabel('Name').fill(`${supplierName} Updated`);
    await page.getByRole('button', { name: 'Update' }).click();

    await expect(page.getByText('Supplier updated successfully')).toBeVisible({ timeout: 10000 });
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
    const supplierName = await firstRow.getByRole('cell').nth(1).textContent();

    await firstRow.getByRole('button', { name: 'Delete' }).click();

    await expect(page.getByText('Delete Supplier')).toBeVisible();
    await expect(page.getByText(/Are you sure you want to delete/)).toBeVisible();
    await expect(page.locator('strong', { hasText: supplierName! })).toBeVisible();
  });

  test('should cancel delete and keep the supplier', async ({ page }) => {
    const firstRow = page.getByRole('row').nth(1);
    const supplierName = await firstRow.getByRole('cell').nth(1).textContent();

    await firstRow.getByRole('button', { name: 'Delete' }).click();
    await page.getByRole('button', { name: 'Cancel' }).click();

    // Supplier should still be in the table
    await expect(page.getByRole('cell', { name: supplierName! })).toBeVisible();
  });

  test('should delete a supplier', async ({ page }) => {
    // Create a supplier to delete
    const supplierName = `Delete Me ${Date.now()}`;
    await page.getByRole('button', { name: 'Add Supplier' }).click();
    await page.getByLabel('Name').fill(supplierName);
    await page.getByPlaceholder('Supplier address').fill('Delete Street');
    await page.getByRole('button', { name: 'Create' }).click();
    await expect(page.getByText('Supplier created successfully')).toBeVisible({ timeout: 10000 });

    // Search for it and delete
    await page.getByPlaceholder('Search suppliers...').fill(supplierName);
    await expect(page.getByRole('cell', { name: supplierName })).toBeVisible({ timeout: 5000 });

    await page.getByRole('button', { name: 'Delete' }).click();
    await expect(page.getByText(/Are you sure you want to delete/)).toBeVisible();

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
