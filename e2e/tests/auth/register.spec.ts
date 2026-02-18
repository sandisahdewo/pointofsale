import { test, expect } from '@playwright/test';

test.describe('Register Page', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/register');
  });

  test('should display all form elements', async ({ page }) => {
    await expect(page.getByRole('heading', { name: 'Create Account' })).toBeVisible();
    await expect(page.getByText('Register a new account')).toBeVisible();
    await expect(page.getByLabel('Name')).toBeVisible();
    await expect(page.getByLabel('Email')).toBeVisible();
    await expect(page.getByLabel('Password', { exact: true })).toBeVisible();
    await expect(page.getByLabel('Confirm Password')).toBeVisible();
    await expect(page.getByRole('button', { name: 'Register' })).toBeVisible();
    await expect(page.getByRole('link', { name: 'Login' })).toBeVisible();
  });

  test('should show validation errors on empty submit', async ({ page }) => {
    await page.getByRole('button', { name: 'Register' }).click();

    await expect(page.getByText('Name is required')).toBeVisible();
    await expect(page.getByText('Email is required')).toBeVisible();
    await expect(page.getByText('Password is required')).toBeVisible();
    await expect(page.getByText('Please confirm your password')).toBeVisible();
  });

  test('should prevent submission with invalid email via browser validation', async ({ page }) => {
    const emailInput = page.getByLabel('Email');
    await page.getByLabel('Name').fill('Test User');
    await emailInput.fill('invalid-email');
    await page.getByLabel('Password', { exact: true }).fill('Test@1234');
    await page.getByLabel('Confirm Password').fill('Test@1234');
    await page.getByRole('button', { name: 'Register' }).click();

    // Browser's native type="email" validation prevents form submission
    await expect(emailInput).toHaveJSProperty('validity.typeMismatch', true);
  });

  test('should show error for password shorter than 8 characters', async ({ page }) => {
    await page.getByLabel('Name').fill('Test User');
    await page.getByLabel('Email').fill('test@example.com');
    await page.getByLabel('Password', { exact: true }).fill('Ab@1');
    await page.getByLabel('Confirm Password').fill('Ab@1');
    await page.getByRole('button', { name: 'Register' }).click();

    await expect(page.getByText('Password must be at least 8 characters')).toBeVisible();
  });

  test('should show error for password missing complexity requirements', async ({ page }) => {
    await page.getByLabel('Name').fill('Test User');
    await page.getByLabel('Email').fill('test@example.com');
    await page.getByLabel('Password', { exact: true }).fill('alllowercase');
    await page.getByLabel('Confirm Password').fill('alllowercase');
    await page.getByRole('button', { name: 'Register' }).click();

    await expect(page.getByText('Password must contain uppercase, lowercase, digit, and special character')).toBeVisible();
  });

  test('should show error when passwords do not match', async ({ page }) => {
    await page.getByLabel('Name').fill('Test User');
    await page.getByLabel('Email').fill('test@example.com');
    await page.getByLabel('Password', { exact: true }).fill('Test@1234');
    await page.getByLabel('Confirm Password').fill('Different@1234');
    await page.getByRole('button', { name: 'Register' }).click();

    await expect(page.getByText('Passwords do not match')).toBeVisible();
  });

  test('should navigate to login page', async ({ page }) => {
    await page.getByRole('link', { name: 'Login' }).click();
    await expect(page).toHaveURL('/login');
  });
});
