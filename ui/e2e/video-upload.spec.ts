import { expect, test } from "@playwright/test";
import path from "path";

test.describe("Video Upload", () => {
	test.beforeEach(async ({ page }) => {
		// Register and login
		const timestamp = Date.now();
		const email = `test${timestamp}@example.com`;
		const password = "Password123!";

		await page.goto("/register");
		await page.getByLabel(/username/i).fill(`user${timestamp}`);
		await page.getByLabel(/email/i).fill(email);
		await page.getByLabel("Password", { exact: true }).fill(password);
		await page.getByLabel(/confirm password/i).fill(password);
		await page.getByRole("button", { name: /register/i }).click();

		await expect(page).toHaveURL("/login");

		await page.getByLabel(/email/i).fill(email);
		await page.getByLabel(/password/i).fill(password);
		await page.getByRole("button", { name: /login/i }).click();

		await expect(page).toHaveURL("/videos", { timeout: 10000 });
	});

	test("should display upload page", async ({ page }) => {
		await page.goto("/upload");

		await expect(page.locator("h1")).toContainText("Upload Video");
		await expect(page.locator("text=/drag and drop/i")).toBeVisible();
	});

	test("should navigate to upload page from videos page", async ({ page }) => {
		await page.getByRole("link", { name: /upload/i }).click();

		await expect(page).toHaveURL("/upload");
		await expect(page.locator("h1")).toContainText("Upload Video");
	});

	test("should show upload zone", async ({ page }) => {
		await page.goto("/upload");

		// Should see drag and drop zone
		await expect(page.locator("text=/drag.*drop/i")).toBeVisible();
		await expect(page.locator("text=/click to select/i")).toBeVisible();
	});

	test("should display empty state on videos page", async ({ page }) => {
		await page.goto("/videos");

		// New user should see empty state
		await expect(page.locator("text=/no videos/i")).toBeVisible();
		await expect(page.getByRole("link", { name: /upload/i })).toBeVisible();
	});

	test("should show navigation header", async ({ page }) => {
		await expect(page.getByRole("link", { name: /videos/i })).toBeVisible();
		await expect(page.getByRole("link", { name: /upload/i })).toBeVisible();
		await expect(page.getByRole("button", { name: /logout/i })).toBeVisible();
	});
});
