import { Page } from '@playwright/test';

/**
 * Log in as a user via the login page.
 * Reuse across tests that require an authenticated session.
 */
export async function login(page: Page, email: string, password: string) {
  await page.goto('/login');
  await page.getByLabel('Email').fill(email);
  await page.getByLabel('Password').fill(password);
  await page.getByRole('button', { name: /login|sign in/i }).click();
  await page.waitForURL('/dashboard');
}
