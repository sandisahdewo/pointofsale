import { test, expect } from '@playwright/test';

test.describe('Login Page', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/login');
  });

  test('should display all form elements', async ({ page }) => {
    await expect(page.getByRole('heading', { name: 'Point of Sale' })).toBeVisible();
    await expect(page.getByText('Sign in to your account')).toBeVisible();
    await expect(page.getByLabel('Email')).toBeVisible();
    await expect(page.getByLabel('Password')).toBeVisible();
    await expect(page.getByLabel('Remember me')).toBeVisible();
    await expect(page.getByRole('button', { name: 'Login' })).toBeVisible();
    await expect(page.getByRole('link', { name: 'Forgot your password?' })).toBeVisible();
    await expect(page.getByRole('link', { name: 'Register' })).toBeVisible();
  });

  test('should show validation errors on empty submit', async ({ page }) => {
    await page.getByRole('button', { name: 'Login' }).click();

    await expect(page.getByText('Email is required')).toBeVisible();
    await expect(page.getByText('Password is required')).toBeVisible();
  });

  test('should prevent submission with invalid email via browser validation', async ({ page }) => {
    const emailInput = page.getByLabel('Email');
    await emailInput.fill('not-an-email');
    await page.getByLabel('Password').fill('somepassword');
    await page.getByRole('button', { name: 'Login' }).click();

    // Browser's native type="email" validation prevents form submission
    await expect(emailInput).toHaveJSProperty('validity.typeMismatch', true);
  });

  test('should show only email error when password is provided', async ({ page }) => {
    await page.getByLabel('Password').fill('somepassword');
    await page.getByRole('button', { name: 'Login' }).click();

    await expect(page.getByText('Email is required')).toBeVisible();
    await expect(page.getByText('Password is required')).not.toBeVisible();
  });

  test('should show only password error when email is provided', async ({ page }) => {
    await page.getByLabel('Email').fill('test@example.com');
    await page.getByRole('button', { name: 'Login' }).click();

    await expect(page.getByText('Password is required')).toBeVisible();
    await expect(page.getByText('Email is required')).not.toBeVisible();
  });

  test('should toggle remember me checkbox', async ({ page }) => {
    const checkbox = page.getByLabel('Remember me');
    await expect(checkbox).not.toBeChecked();
    await checkbox.check();
    await expect(checkbox).toBeChecked();
    await checkbox.uncheck();
    await expect(checkbox).not.toBeChecked();
  });

  test('should login successfully with valid credentials and redirect to dashboard', async ({ page }) => {
    await page.getByLabel('Email').fill('admin@pointofsale.com');
    await page.getByLabel('Password').fill('Admin@12345');
    await page.getByRole('button', { name: 'Login' }).click();

    // Should show success toast and redirect to dashboard
    await expect(page.getByText('Login successful')).toBeVisible({ timeout: 10000 });
    await expect(page).toHaveURL('/dashboard', { timeout: 10000 });
  });

  test('should show error for invalid credentials', async ({ page }) => {
    await page.getByLabel('Email').fill('admin@pointofsale.com');
    await page.getByLabel('Password').fill('WrongPassword@123');
    await page.getByRole('button', { name: 'Login' }).click();

    await expect(page.getByText('Invalid email or password')).toBeVisible({ timeout: 10000 });
    // Should stay on login page
    await expect(page).toHaveURL('/login');
  });

  test('should navigate to register page', async ({ page }) => {
    await page.getByRole('link', { name: 'Register' }).click();
    await expect(page).toHaveURL('/register');
  });

  test('should navigate to reset password page', async ({ page }) => {
    await page.getByRole('link', { name: 'Forgot your password?' }).click();
    await expect(page).toHaveURL('/reset-password');
  });
});
