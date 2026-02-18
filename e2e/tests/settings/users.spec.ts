import { test, expect } from '@playwright/test';
import { login } from '@helpers/auth';

test.describe('Settings Users', () => {
  test.beforeEach(async ({ page }) => {
    await login(page, 'admin@pointofsale.com', 'Admin@12345');
    await page.goto('/settings/users');
    await expect(page.getByRole('heading', { name: 'Users' })).toBeVisible({ timeout: 10000 });
  });

  // --- Page Load & Display ---

  test('should display page heading and key elements', async ({ page }) => {
    await expect(page.getByRole('button', { name: 'Create User' })).toBeVisible();
    await expect(page.getByPlaceholder('Search users...')).toBeVisible();
    // Table headers
    await expect(page.getByRole('columnheader', { name: /ID/i })).toBeVisible();
    await expect(page.getByRole('columnheader', { name: /Profile/i })).toBeVisible();
    await expect(page.getByRole('columnheader', { name: /Name/i })).toBeVisible();
    await expect(page.getByRole('columnheader', { name: /Email/i })).toBeVisible();
    await expect(page.getByRole('columnheader', { name: /Phone/i })).toBeVisible();
    await expect(page.getByRole('columnheader', { name: /Roles/i })).toBeVisible();
    await expect(page.getByRole('columnheader', { name: /Status/i })).toBeVisible();
    await expect(page.getByRole('columnheader', { name: /Actions/i })).toBeVisible();
  });

  test('should display users in the table', async ({ page }) => {
    const rows = page.getByRole('row');
    await expect(rows).not.toHaveCount(1); // more than just the header
  });

  test('should show pagination info', async ({ page }) => {
    await expect(page.getByText(/Showing \d+-\d+ of \d+ items/)).toBeVisible();
    await expect(page.getByLabel('Items per page:')).toBeVisible();
  });

  // --- Search ---

  test('should filter users by search query', async ({ page }) => {
    // Use the known admin user
    await page.getByPlaceholder('Search users...').fill('admin');
    await expect(page.getByRole('cell', { name: /admin/i }).first()).toBeVisible({ timeout: 5000 });
  });

  test('should show empty state when search has no results', async ({ page }) => {
    await page.getByPlaceholder('Search users...').fill('zzz_nonexistent_user_xyz');
    await expect(page.getByText('No data available')).toBeVisible({ timeout: 5000 });
  });

  test('should reset results when search is cleared', async ({ page }) => {
    await page.getByPlaceholder('Search users...').fill('zzz_nonexistent_user_xyz');
    await expect(page.getByText('No data available')).toBeVisible({ timeout: 5000 });

    await page.getByPlaceholder('Search users...').fill('');
    await expect(page.getByText('No data available')).not.toBeVisible({ timeout: 5000 });
  });

  // --- Create ---

  test('should open create user modal with roles loaded', async ({ page }) => {
    await page.getByRole('button', { name: 'Create User' }).click();
    await expect(page.getByRole('heading', { name: 'Create User' })).toBeVisible();
    await expect(page.getByLabel('Name')).toBeVisible();
    await expect(page.getByLabel('Email')).toBeVisible();
    await expect(page.getByLabel('Phone')).toBeVisible();
    await expect(page.getByLabel('Address')).toBeVisible();
    await expect(page.getByRole('button', { name: 'Select roles...' })).toBeVisible();
    // Status should NOT be visible in create mode
    await expect(page.getByLabel('Status')).not.toBeVisible();
    await expect(page.getByRole('button', { name: 'Cancel' })).toBeVisible();
    await expect(page.getByRole('button', { name: 'Save' })).toBeVisible();

    // Verify roles are loaded in the dropdown
    await page.getByRole('button', { name: 'Select roles...' }).click();
    // Should NOT show "No options available" — roles must be fetched
    await expect(page.getByText('No options available')).not.toBeVisible();
    // At least one role checkbox should be visible (e.g., Admin)
    await expect(page.getByRole('checkbox').first()).toBeVisible({ timeout: 5000 });
  });

  test('should show validation error when submitting empty required fields', async ({ page }) => {
    await page.getByRole('button', { name: 'Create User' }).click();
    await page.getByRole('button', { name: 'Save' }).click();
    // HTML5 native required validation prevents form submission
    const nameInput = page.getByLabel('Name');
    expect(await nameInput.evaluate((el: HTMLInputElement) => el.validity.valueMissing)).toBe(true);
  });

  test('should show validation error for invalid email', async ({ page }) => {
    await page.getByRole('button', { name: 'Create User' }).click();
    await page.getByLabel('Name').fill('Test User');
    // Use email that passes HTML5 type=email but fails custom regex (requires dot in domain)
    await page.getByLabel('Email').fill('a@b');
    await page.getByRole('button', { name: 'Save' }).click();
    await expect(page.getByText('Please enter a valid email address')).toBeVisible();
  });

  test('should create a new user with a role selected', async ({ page }) => {
    const userName = `Test User ${Date.now()}`;
    const userEmail = `testuser${Date.now()}@example.com`;

    await page.getByRole('button', { name: 'Create User' }).click();
    await page.getByLabel('Name').fill(userName);
    await page.getByLabel('Email').fill(userEmail);
    await page.getByLabel('Phone').fill('08123456789');

    // Open roles dropdown and wait for options to load
    await page.getByRole('button', { name: 'Select roles...' }).click();
    await expect(page.getByRole('checkbox').first()).toBeVisible({ timeout: 5000 });

    // Select the "Cashier" role via JS click (dropdown may extend beyond viewport in modal)
    await page.getByLabel('Cashier').evaluate((el: HTMLElement) => el.click());
    // Verify selected role shows on the trigger button
    await expect(page.getByRole('button', { name: 'Cashier' })).toBeVisible();

    // Close dropdown by clicking outside it
    await page.getByLabel('Name').click();

    // Save via JS click (button may be below viewport in modal)
    await page.getByRole('button', { name: 'Save' }).evaluate((el: HTMLElement) => el.click());

    await expect(page.getByText('User created successfully')).toBeVisible({ timeout: 10000 });
    // Modal should close
    await expect(page.getByRole('heading', { name: 'Create User' })).not.toBeVisible();

    // Verify the role shows in the table
    await page.getByPlaceholder('Search users...').fill(userName);
    await expect(page.getByRole('cell', { name: userName })).toBeVisible({ timeout: 5000 });
    const row = page.getByRole('row').filter({ hasText: userName });
    await expect(row.getByText('Cashier')).toBeVisible();
  });

  test('should create a new user with multiple roles selected', async ({ page }) => {
    const userName = `Multi Role ${Date.now()}`;
    const userEmail = `multirole${Date.now()}@example.com`;

    await page.getByRole('button', { name: 'Create User' }).click();
    await page.getByLabel('Name').fill(userName);
    await page.getByLabel('Email').fill(userEmail);

    // Open roles dropdown and wait for options
    await page.getByRole('button', { name: 'Select roles...' }).click();
    await expect(page.getByRole('checkbox').first()).toBeVisible({ timeout: 5000 });

    // Select Manager and Cashier (use JS click — dropdown may extend beyond viewport)
    await page.getByLabel('Manager').evaluate((el: HTMLElement) => el.click());
    await page.getByLabel('Cashier').evaluate((el: HTMLElement) => el.click());

    // Trigger button should show both selected roles
    const triggerButton = page.locator('button', { hasText: /Manager/ }).filter({ hasText: /Cashier/ });
    await expect(triggerButton).toBeVisible();

    // Close dropdown
    await page.getByLabel('Name').click();

    // Save
    await page.getByRole('button', { name: 'Save' }).evaluate((el: HTMLElement) => el.click());
    await expect(page.getByText('User created successfully')).toBeVisible({ timeout: 10000 });
    await expect(page.getByRole('heading', { name: 'Create User' })).not.toBeVisible();

    // Verify both roles show in the table
    await page.getByPlaceholder('Search users...').fill(userName);
    await expect(page.getByRole('cell', { name: userName })).toBeVisible({ timeout: 5000 });
    const row = page.getByRole('row').filter({ hasText: userName });
    await expect(row.getByText('Manager')).toBeVisible();
    await expect(row.getByText('Cashier')).toBeVisible();
  });

  test('should edit a user to change roles', async ({ page }) => {
    test.slow();

    // Create a user with Cashier role first
    const userName = `Role Edit ${Date.now()}`;
    const userEmail = `roleedit${Date.now()}@example.com`;

    await page.getByRole('button', { name: 'Create User' }).click();
    await page.getByLabel('Name').fill(userName);
    await page.getByLabel('Email').fill(userEmail);

    // Select Cashier role
    await page.getByRole('button', { name: 'Select roles...' }).click();
    await expect(page.getByRole('checkbox').first()).toBeVisible({ timeout: 5000 });
    await page.getByLabel('Cashier').evaluate((el: HTMLElement) => el.click());
    await page.getByLabel('Name').click();
    await page.getByRole('button', { name: 'Save' }).evaluate((el: HTMLElement) => el.click());
    await expect(page.getByText('User created successfully')).toBeVisible({ timeout: 10000 });

    // Search and open edit modal
    await page.getByPlaceholder('Search users...').fill(userName);
    await expect(page.getByRole('cell', { name: userName })).toBeVisible({ timeout: 5000 });
    const row = page.getByRole('row').filter({ hasText: userName });
    await row.getByRole('button', { name: 'Edit' }).click();
    await expect(page.getByRole('heading', { name: 'Edit User' })).toBeVisible();

    // Open roles dropdown — Cashier should already be checked
    await page.locator('button', { hasText: 'Cashier' }).evaluate((el: HTMLElement) => el.click());
    await expect(page.getByRole('checkbox').first()).toBeVisible({ timeout: 5000 });
    await expect(page.getByLabel('Cashier')).toBeChecked();

    // Add Manager role and deselect Cashier
    await page.getByLabel('Manager').evaluate((el: HTMLElement) => el.click());
    await page.getByLabel('Cashier').evaluate((el: HTMLElement) => el.click());

    // Close dropdown
    await page.getByLabel('Name').click();

    // Save
    await page.getByRole('button', { name: 'Save' }).evaluate((el: HTMLElement) => el.click());
    await expect(page.getByText('User updated successfully')).toBeVisible({ timeout: 10000 });

    // Verify updated roles in the table
    await page.getByPlaceholder('Search users...').fill(userName);
    await expect(page.getByRole('cell', { name: userName })).toBeVisible({ timeout: 5000 });
    const updatedRow = page.getByRole('row').filter({ hasText: userName });
    await expect(updatedRow.getByText('Manager')).toBeVisible();
    // Cashier should no longer be listed
    await expect(updatedRow.getByText('Cashier')).not.toBeVisible();
  });

  test('should close create modal on cancel', async ({ page }) => {
    await page.getByRole('button', { name: 'Create User' }).click();
    await expect(page.getByRole('heading', { name: 'Create User' })).toBeVisible();

    await page.getByRole('button', { name: 'Cancel' }).click();
    await expect(page.getByRole('heading', { name: 'Create User' })).not.toBeVisible();
  });

  // --- Edit ---

  test('should open edit modal with pre-filled data', async ({ page }) => {
    test.slow();

    // Create a user to edit
    const userName = `Prefill Me ${Date.now()}`;
    const userEmail = `prefill${Date.now()}@example.com`;
    await page.getByRole('button', { name: 'Create User' }).click();
    await page.getByLabel('Name').fill(userName);
    await page.getByLabel('Email').fill(userEmail);
    await page.getByRole('button', { name: 'Save' }).click();
    await expect(page.getByText('User created successfully')).toBeVisible({ timeout: 10000 });

    // Search and open edit
    await page.getByPlaceholder('Search users...').fill(userName);
    await expect(page.getByRole('cell', { name: userName })).toBeVisible({ timeout: 5000 });

    const row = page.getByRole('row').filter({ hasText: userName });
    await row.getByRole('button', { name: 'Edit' }).click();

    await expect(page.getByRole('heading', { name: 'Edit User' })).toBeVisible();
    await expect(page.getByLabel('Name')).toHaveValue(userName);
    await expect(page.getByLabel('Email')).toHaveValue(userEmail);
    // Status should be visible in edit mode
    await expect(page.getByLabel('Status')).toBeVisible();

    // Verify roles are loaded in the edit modal dropdown
    await page.getByRole('button', { name: /Select roles|roles/i }).evaluate((el: HTMLElement) => el.click());
    await expect(page.getByRole('checkbox').first()).toBeVisible({ timeout: 5000 });
    // Should have role options, not empty
    await expect(page.getByText('No options available')).not.toBeVisible();
  });

  test('should update a user', async ({ page }) => {
    test.slow();

    // Create a user to update
    const userName = `Edit Me ${Date.now()}`;
    const userEmail = `editme${Date.now()}@example.com`;
    await page.getByRole('button', { name: 'Create User' }).click();
    await page.getByLabel('Name').fill(userName);
    await page.getByLabel('Email').fill(userEmail);
    await page.getByRole('button', { name: 'Save' }).click();
    await expect(page.getByText('User created successfully')).toBeVisible({ timeout: 10000 });

    // Search and edit
    await page.getByPlaceholder('Search users...').fill(userName);
    await expect(page.getByRole('cell', { name: userName })).toBeVisible({ timeout: 5000 });

    const row = page.getByRole('row').filter({ hasText: userName });
    await row.getByRole('button', { name: 'Edit' }).click();
    await page.getByLabel('Name').fill(`${userName} Updated`);
    // Save button may be outside viewport in edit modal, use JS click
    await page.getByRole('button', { name: 'Save' }).evaluate(el => (el as HTMLElement).click());

    await expect(page.getByText('User updated successfully')).toBeVisible({ timeout: 10000 });
  });

  test('should show validation error when editing name to empty', async ({ page }) => {
    // Use the first non-superadmin row
    await page.getByRole('row').nth(1).getByRole('button', { name: 'Edit' }).click();
    await expect(page.getByRole('heading', { name: 'Edit User' })).toBeVisible();

    await page.getByLabel('Name').fill('');
    // Submit via Enter key since Save button may be outside viewport in edit modal
    await page.getByLabel('Email').press('Enter');
    // HTML5 native required validation prevents form submission
    const nameInput = page.getByLabel('Name');
    expect(await nameInput.evaluate((el: HTMLInputElement) => el.validity.valueMissing)).toBe(true);
  });

  // --- Delete ---

  test('should open delete confirmation modal', async ({ page }) => {
    test.slow();

    // Create a user to delete
    const userName = `Delete Modal ${Date.now()}`;
    const userEmail = `deletemodal${Date.now()}@example.com`;
    await page.getByRole('button', { name: 'Create User' }).click();
    await page.getByLabel('Name').fill(userName);
    await page.getByLabel('Email').fill(userEmail);
    await page.getByRole('button', { name: 'Save' }).click();
    await expect(page.getByText('User created successfully')).toBeVisible({ timeout: 10000 });

    // Search and open delete modal
    await page.getByPlaceholder('Search users...').fill(userName);
    await expect(page.getByRole('cell', { name: userName })).toBeVisible({ timeout: 5000 });

    const row = page.getByRole('row').filter({ hasText: userName });
    await row.getByRole('button', { name: 'Delete' }).click();

    await expect(page.getByText('Delete User')).toBeVisible();
    await expect(page.getByText(/Are you sure you want to delete/)).toBeVisible();
  });

  test('should cancel delete and keep the user', async ({ page }) => {
    test.slow();

    // Create a user to test cancel
    const userName = `Cancel Delete ${Date.now()}`;
    const userEmail = `canceldelete${Date.now()}@example.com`;
    await page.getByRole('button', { name: 'Create User' }).click();
    await page.getByLabel('Name').fill(userName);
    await page.getByLabel('Email').fill(userEmail);
    await page.getByRole('button', { name: 'Save' }).click();
    await expect(page.getByText('User created successfully')).toBeVisible({ timeout: 10000 });

    // Search, open delete, cancel
    await page.getByPlaceholder('Search users...').fill(userName);
    await expect(page.getByRole('cell', { name: userName })).toBeVisible({ timeout: 5000 });

    const row = page.getByRole('row').filter({ hasText: userName });
    await row.getByRole('button', { name: 'Delete' }).click();
    await expect(page.getByText(/Are you sure you want to delete/)).toBeVisible();

    await page.getByRole('button', { name: 'Cancel' }).click();
    // User should still be visible
    await expect(page.getByRole('cell', { name: userName })).toBeVisible();
  });

  test('should delete a user', async ({ page }) => {
    test.slow();

    // Create a user to delete
    const userName = `Delete Me ${Date.now()}`;
    const userEmail = `deleteme${Date.now()}@example.com`;
    await page.getByRole('button', { name: 'Create User' }).click();
    await page.getByLabel('Name').fill(userName);
    await page.getByLabel('Email').fill(userEmail);
    await page.getByRole('button', { name: 'Save' }).click();
    await expect(page.getByText('User created successfully')).toBeVisible({ timeout: 10000 });

    // Search and delete
    await page.getByPlaceholder('Search users...').fill(userName);
    await expect(page.getByRole('cell', { name: userName })).toBeVisible({ timeout: 5000 });

    const row = page.getByRole('row').filter({ hasText: userName });
    await row.getByRole('button', { name: 'Delete' }).click();
    await expect(page.getByText(/Are you sure you want to delete/)).toBeVisible();

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

    const responsePromise = page.waitForResponse(resp => resp.url().includes('/users') && resp.status() === 200);
    await pageSizeSelect.selectOption('5');
    await responsePromise;
    await expect(pageSizeSelect).toHaveValue('5');
    const rows = page.getByRole('row');
    await expect(rows).toHaveCount(6, { timeout: 5000 }); // header + 5
  });

  // --- Feature-Specific: Approve & Reject ---

  test('should approve a pending user', async ({ page }) => {
    test.slow();

    // Create prerequisite state: register a new user
    await page.goto('/register');
    const userName = `Approve Me ${Date.now()}`;
    const userEmail = `approve${Date.now()}@example.com`;
    await page.getByLabel('Name').fill(userName);
    await page.getByLabel('Email').fill(userEmail);
    await page.getByLabel('Password', { exact: true }).fill('Test@12345');
    await page.getByLabel('Confirm Password').fill('Test@12345');
    await page.getByRole('button', { name: 'Register' }).click();
    await expect(page.getByText(/registration successful/i)).toBeVisible({ timeout: 10000 });

    // Re-login as admin and navigate to users
    await login(page, 'admin@pointofsale.com', 'Admin@12345');
    await page.goto('/settings/users');
    await expect(page.getByRole('heading', { name: 'Users' })).toBeVisible({ timeout: 10000 });

    // Search for the pending user
    await page.getByPlaceholder('Search users...').fill(userName);
    await expect(page.getByRole('cell', { name: userName })).toBeVisible({ timeout: 5000 });

    const row = page.getByRole('row').filter({ hasText: userName });
    await row.getByRole('button', { name: 'Approve' }).click();
    await expect(page.getByText(/has been approved/)).toBeVisible({ timeout: 10000 });
  });

  test('should reject a pending user', async ({ page }) => {
    test.slow();

    // Create prerequisite state: register a new user
    await page.goto('/register');
    const userName = `Reject Me ${Date.now()}`;
    const userEmail = `reject${Date.now()}@example.com`;
    await page.getByLabel('Name').fill(userName);
    await page.getByLabel('Email').fill(userEmail);
    await page.getByLabel('Password', { exact: true }).fill('Test@12345');
    await page.getByLabel('Confirm Password').fill('Test@12345');
    await page.getByRole('button', { name: 'Register' }).click();
    await expect(page.getByText(/registration successful/i)).toBeVisible({ timeout: 10000 });

    // Re-login as admin
    await login(page, 'admin@pointofsale.com', 'Admin@12345');
    await page.goto('/settings/users');
    await expect(page.getByRole('heading', { name: 'Users' })).toBeVisible({ timeout: 10000 });

    // Search for the pending user
    await page.getByPlaceholder('Search users...').fill(userName);
    await expect(page.getByRole('cell', { name: userName })).toBeVisible({ timeout: 5000 });

    const row = page.getByRole('row').filter({ hasText: userName });
    await row.getByRole('button', { name: 'Reject' }).click();
    await expect(page.getByText(/Are you sure you want to reject/)).toBeVisible();

    // Confirm rejection
    await page.getByRole('button', { name: 'Reject' }).nth(1).click();
    await expect(page.getByText(/has been rejected/)).toBeVisible({ timeout: 10000 });
  });

  // --- Super Admin Protection ---

  test('should disable delete button for super admin', async ({ page }) => {
    // Search for the admin user
    await page.getByPlaceholder('Search users...').fill('admin@pointofsale.com');
    await expect(page.getByRole('cell', { name: 'admin@pointofsale.com' })).toBeVisible({ timeout: 5000 });

    const row = page.getByRole('row').filter({ hasText: 'admin@pointofsale.com' });
    const deleteButton = row.getByRole('button', { name: 'Delete' });
    await expect(deleteButton).toBeDisabled();
  });
});
