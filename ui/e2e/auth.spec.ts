import { expect, test } from '@playwright/test';

test.describe('Authentication', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
  });

  test('should display login page', async ({ page }) => {
    await page.goto('/login');

    await expect(page.locator('h1')).toContainText('Login');
    await expect(page.getByLabel(/email/i)).toBeVisible();
    await expect(page.getByLabel(/password/i)).toBeVisible();
    await expect(page.getByRole('button', { name: /login/i })).toBeVisible();
  });

  test('should display register page', async ({ page }) => {
    await page.goto('/register');

    await expect(page.locator('h1')).toContainText('Register');
    await expect(page.getByLabel(/username/i)).toBeVisible();
    await expect(page.getByLabel(/email/i)).toBeVisible();
    await expect(page.getByLabel('Password', { exact: true })).toBeVisible();
    await expect(page.getByLabel(/confirm password/i)).toBeVisible();
    await expect(page.getByRole('button', { name: /register/i })).toBeVisible();
  });

  test('should show validation errors on empty login form', async ({ page }) => {
    await page.goto('/login');

    await page.getByRole('button', { name: /login/i }).click();

    // Wait for validation messages
    await expect(page.locator('text=/email is required/i')).toBeVisible();
    await expect(page.locator('text=/password is required/i')).toBeVisible();
  });

  test('should show validation errors on empty register form', async ({ page }) => {
    await page.goto('/register');

    await page.getByRole('button', { name: /register/i }).click();

    // Wait for validation messages
    await expect(page.locator('text=/username is required/i')).toBeVisible();
    await expect(page.locator('text=/email is required/i')).toBeVisible();
  });

  test('should register new user and redirect to login', async ({ page }) => {
    await page.goto('/register');

    const timestamp = Date.now();
    const username = `testuser${timestamp}`;
    const email = `test${timestamp}@example.com`;
    const password = 'Password123!';

    await page.getByLabel(/username/i).fill(username);
    await page.getByLabel(/email/i).fill(email);
    await page.getByLabel('Password', { exact: true }).fill(password);
    await page.getByLabel(/confirm password/i).fill(password);

    await page.getByRole('button', { name: /register/i }).click();

    // Should redirect to login page
    await expect(page).toHaveURL('/login');

    // Should show success message
    await expect(page.locator('text=/registration successful/i')).toBeVisible();
  });

  test('should login successfully and redirect to videos page', async ({ page }) => {
    // First register a user
    await page.goto('/register');

    const timestamp = Date.now();
    const email = `test${timestamp}@example.com`;
    const password = 'Password123!';

    await page.getByLabel(/username/i).fill(`user${timestamp}`);
    await page.getByLabel(/email/i).fill(email);
    await page.getByLabel('Password', { exact: true }).fill(password);
    await page.getByLabel(/confirm password/i).fill(password);
    await page.getByRole('button', { name: /register/i }).click();

    // Wait for redirect to login
    await expect(page).toHaveURL('/login');

    // Now login
    await page.getByLabel(/email/i).fill(email);
    await page.getByLabel(/password/i).fill(password);
    await page.getByRole('button', { name: /login/i }).click();

    // Should redirect to videos page
    await expect(page).toHaveURL('/videos', { timeout: 10000 });

    // Should see user navigation
    await expect(page.getByRole('button', { name: /logout/i })).toBeVisible();
  });

  test('should show error on invalid credentials', async ({ page }) => {
    await page.goto('/login');

    await page.getByLabel(/email/i).fill('invalid@example.com');
    await page.getByLabel(/password/i).fill('wrongpassword');
    await page.getByRole('button', { name: /login/i }).click();

    // Should show error message
    await expect(page.locator('text=/invalid credentials/i')).toBeVisible();
  });

  test('should logout successfully', async ({ page }) => {
    // Login first
    await page.goto('/register');

    const timestamp = Date.now();
    const email = `test${timestamp}@example.com`;
    const password = 'Password123!';

    await page.getByLabel(/username/i).fill(`user${timestamp}`);
    await page.getByLabel(/email/i).fill(email);
    await page.getByLabel('Password', { exact: true }).fill(password);
    await page.getByLabel(/confirm password/i).fill(password);
    await page.getByRole('button', { name: /register/i }).click();

    await expect(page).toHaveURL('/login');

    await page.getByLabel(/email/i).fill(email);
    await page.getByLabel(/password/i).fill(password);
    await page.getByRole('button', { name: /login/i }).click();

    await expect(page).toHaveURL('/videos', { timeout: 10000 });

    // Logout
    await page.getByRole('button', { name: /logout/i }).click();

    // Should redirect to login
    await expect(page).toHaveURL('/login');
  });

  test('should redirect to login when accessing protected route without auth', async ({ page }) => {
    await page.goto('/videos');

    // Should redirect to login
    await expect(page).toHaveURL('/login');
  });

  test('should redirect to login when accessing upload page without auth', async ({ page }) => {
    await page.goto('/upload');

    // Should redirect to login
    await expect(page).toHaveURL('/login');
  });
});
