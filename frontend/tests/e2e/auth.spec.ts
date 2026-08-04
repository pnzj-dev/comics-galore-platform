import { test, expect } from '@playwright/test';

test.describe('Auth pages', () => {
	test('login page loads and shows form', async ({ page }) => {
		await page.goto('/login');
		await expect(page.locator('h2')).toContainText('Login');
		await expect(page.getByLabel('Email')).toBeVisible();
		await expect(page.getByLabel('Password')).toBeVisible();
		await expect(page.getByRole('button', { name: 'Sign in' })).toBeVisible();
	});

	test('register page loads and shows form', async ({ page }) => {
		await page.goto('/register');
		await expect(page.locator('h2')).toContainText('Register');
		await expect(page.getByLabel('Email')).toBeVisible();
		await expect(page.getByLabel('Password')).toBeVisible();
		await expect(page.getByLabel('Confirm Password')).toBeVisible();
		await expect(page.getByRole('button', { name: 'Create account' })).toBeVisible();
	});
});

test.describe('Unauthenticated user', () => {
	test('sees Login and Register in nav', async ({ page }) => {
		await page.goto('/');
		await expect(page.getByRole('link', { name: 'Login' })).toBeVisible();
		await expect(page.getByRole('link', { name: 'Register' })).toBeVisible();
		await expect(page.getByRole('link', { name: 'Get Started' })).toBeVisible();
		await expect(page.getByRole('link', { name: 'Sign In' })).toBeVisible();
	});

	test('is redirected from /upload', async ({ page }) => {
		await page.goto('/upload');
		await expect(page).not.toHaveURL('/upload');
	});

	test('dark mode toggle works', async ({ page }) => {
		await page.goto('/');

		const toggle = page.getByLabel('Toggle theme');
		await expect(toggle).toBeVisible();

		const wasDark = (await page.locator('html').getAttribute('class'))?.includes('dark');

		await toggle.click();

		if (wasDark) {
			await expect(page.locator('html')).not.toHaveClass(/dark/);
		} else {
			await expect(page.locator('html')).toHaveClass(/dark/);
		}
	});
});
