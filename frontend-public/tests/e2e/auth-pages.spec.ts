import { test, expect } from '@playwright/test';

test.describe('Verify Email page', () => {
	test('shows loading state without token', async ({ page }) => {
		await page.goto('/auth/verify-email');
		await expect(page.getByText('Email verified successfully')).not.toBeVisible();
	});

	test('shows error with invalid token', async ({ page }) => {
		await page.goto('/auth/verify-email?token=invalid-token');
		await expect(page.getByText('Invalid or expired verification token')).toBeVisible();
	});
});

test.describe('Reset Password page', () => {
	test('loads and shows email form', async ({ page }) => {
		await page.goto('/auth/reset-password');
		await expect(page.getByText('Reset Password')).toBeVisible();
		await expect(page.getByLabel('Email')).toBeVisible();
		await expect(page.getByRole('button', { name: 'Send Reset Link' })).toBeVisible();
	});

	test('shows confirmation after submit', async ({ page }) => {
		await page.goto('/auth/reset-password');
		await page.getByLabel('Email').fill('test@example.com');
		await page.getByRole('button', { name: 'Send Reset Link' }).click();
		await expect(page.getByText('Check your email')).toBeVisible();
	});
});

test.describe('Reset Password Confirm page', () => {
	test('shows error without token', async ({ page }) => {
		await page.goto('/auth/reset-password/confirm');
		await expect(page.getByText('No reset token provided')).toBeVisible();
	});

	test('shows password form with token', async ({ page }) => {
		await page.goto('/auth/reset-password/confirm?token=some-token');
		await expect(page.getByText('Set New Password')).toBeVisible();
		await expect(page.getByLabel('New Password')).toBeVisible();
		await expect(page.getByLabel('Confirm Password')).toBeVisible();
		await expect(page.getByRole('button', { name: 'Reset Password' })).toBeVisible();
	});
});

test.describe('Home page', () => {
	test('shows hero section', async ({ page }) => {
		await page.goto('/');
		await expect(page.getByText('Welcome to Comics Galore')).toBeVisible();
	});
});
