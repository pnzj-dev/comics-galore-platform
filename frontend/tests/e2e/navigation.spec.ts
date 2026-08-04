import { test, expect } from '@playwright/test';

test.describe('Navigation links', () => {
	test('Browse link navigates to /comics', async ({ page }) => {
		await page.goto('/');
		const browseLink = page.locator('nav a[href="/comics"]');
		await expect(browseLink).toBeVisible();
		await browseLink.click();
		await expect(page).toHaveURL('/comics');
		await expect(page.locator('h1')).toContainText('Browse Comics');
	});

	test('Pricing link navigates to /pricing', async ({ page }) => {
		await page.goto('/');
		const pricingLink = page.locator('nav a[href="/pricing"]');
		await expect(pricingLink).toBeVisible();
		await pricingLink.click();
		await expect(page).toHaveURL('/pricing');
		await expect(page.locator('h1')).toContainText('Plans & Pricing');
	});
});

test.describe('Footer links', () => {
	test('Terms link navigates to /legal/terms', async ({ page }) => {
		await page.goto('/');
		const termsLink = page.locator('footer a[href="/legal/terms"]');
		await expect(termsLink).toBeVisible();
		await termsLink.click();
		await expect(page).toHaveURL('/legal/terms');
		await expect(page.locator('h1')).toContainText('Terms of Service');
	});

	test('Privacy link navigates to /legal/privacy', async ({ page }) => {
		await page.goto('/');
		const privacyLink = page.locator('footer a[href="/legal/privacy"]');
		await expect(privacyLink).toBeVisible();
		await privacyLink.click();
		await expect(page).toHaveURL('/legal/privacy');
		await expect(page.locator('h1')).toContainText('Privacy Policy');
	});

	test('DMCA link navigates to /legal/dmca', async ({ page }) => {
		await page.goto('/');
		const dmcaLink = page.locator('footer a[href="/legal/dmca"]');
		await expect(dmcaLink).toBeVisible();
		await dmcaLink.click();
		await expect(page).toHaveURL('/legal/dmca');
		await expect(page.locator('h1')).toContainText('DMCA');
	});
});

test.describe('Theme toggle', () => {
	test('theme toggle is present in nav', async ({ page }) => {
		await page.goto('/');
		await expect(page.getByLabel('Toggle theme')).toBeVisible();
	});
});
