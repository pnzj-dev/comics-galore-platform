import { test, expect } from '@playwright/test';

test.describe('Public pages', () => {
	test('home page loads and shows welcome heading', async ({ page }) => {
		await page.goto('/');
		await expect(page.locator('h1')).toContainText('Welcome to Comics Galore');
	});

	test('browse page loads and shows heading', async ({ page }) => {
		await page.goto('/comics');
		await expect(page.locator('h1')).toContainText('Browse Comics');
	});

	test('pricing page loads and shows plan grid', async ({ page }) => {
		await page.goto('/pricing');
		await expect(page.locator('h1')).toContainText('Plans & Pricing');
	});

	test('terms page loads', async ({ page }) => {
		await page.goto('/legal/terms');
		await expect(page.locator('h1')).toContainText('Terms of Service');
	});

	test('privacy page loads', async ({ page }) => {
		await page.goto('/legal/privacy');
		await expect(page.locator('h1')).toContainText('Privacy Policy');
	});

	test('dmca page loads', async ({ page }) => {
		await page.goto('/legal/dmca');
		await expect(page.locator('h1')).toContainText('DMCA');
	});
});
