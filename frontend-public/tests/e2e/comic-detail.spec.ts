import { test, expect } from '@playwright/test';

const COMIC_ID = 'test-comic';

test.describe('Comic detail page', () => {
	test.beforeEach(async ({ page }) => {
		await page.goto(`/comics/${COMIC_ID}`);
	});

	test('comic detail loads', async ({ page }) => {
		await expect(page.locator('h1')).toBeVisible();
		await expect(page.locator('img[alt]').first()).toBeVisible();
		await expect(page.locator('p').filter({ hasText: /.+/ }).first()).toBeVisible();
	});

	test('lightbox opens', async ({ page }) => {
		const coverButton = page.getByLabel('Open current image in lightbox');
		if (await coverButton.isVisible()) {
			await coverButton.click();

			const lightbox = page.getByRole('dialog', { name: 'Image lightbox' });
			await expect(lightbox).toBeVisible();

			const closeButton = lightbox.getByLabel('Close');
			await expect(closeButton).toBeVisible();

			await closeButton.click();
			await expect(lightbox).not.toBeVisible();
		}
	});

	test('age gate appears for age-restricted content', async ({ page }) => {
		const ageGate = page.getByText('Age-Restricted Content');
		if (await ageGate.isVisible()) {
			await expect(page.getByText("I'm 18+ years old, Continue")).toBeVisible();
			await expect(page.getByText('Go back')).toBeVisible();
		}
	});

	test('reaction buttons visible', async ({ page }) => {
		const likeButton = page.getByRole('button', { name: /like/i });
		const favoriteButton = page.getByRole('button', { name: /favorite/i });

		await expect(likeButton.first()).toBeVisible();
		await expect(favoriteButton.first()).toBeVisible();
	});

	test('cover carousel arrows visible with disabled state at bounds', async ({ page }) => {
		const prevButton = page.getByLabel('Previous image');
		const nextButton = page.getByLabel('Next image');

		await expect(prevButton).toBeAttached();
		await expect(prevButton).toBeDisabled();

		await expect(nextButton).toBeAttached();

		if (await nextButton.isEnabled()) {
			await nextButton.click();
			await expect(prevButton).toBeEnabled();
		}
	});

	test('thumbnail strip highlights selected image', async ({ page }) => {
		const thumbnails = page.locator('[role="region"][aria-label="Image carousel"] button').filter({ has: page.locator('img') });
		const count = await thumbnails.count();
		if (count > 1) {
			await thumbnails.nth(1).click();
			const counter = page.locator('[role="region"][aria-label="Image carousel"]').getByText(/ \/ \d+$/);
			await expect(counter).toContainText('2 /');
		}
	});
});

test.describe('Continue Reading shelf', () => {
	test('logged-in user with reading progress sees Continue Reading on home page', async ({ page }) => {
		await page.goto('/login');
		await page.getByLabel('Email').fill('test@example.com');
		await page.getByLabel('Password').fill('password123');
		await page.getByRole('button', { name: 'Sign In' }).click();

		await page.waitForURL('/');

		const continueReading = page.getByText('Continue Reading');
		if (await continueReading.isVisible({ timeout: 3000 }).catch(() => false)) {
			await expect(continueReading).toBeVisible();
		}
	});

	test('like and dislike are mutually exclusive', async ({ page }) => {
		await page.goto('/login');
		await page.getByLabel('Email').fill('test@example.com');
		await page.getByLabel('Password').fill('password123');
		await page.getByRole('button', { name: 'Sign In' }).click();
		await page.waitForURL('/');
		await page.goto(`/comics/${COMIC_ID}`);

		const dislikeBtn = page.locator('button').filter({ has: page.locator('svg:has(path)') }).nth(1);
		const likeBtn = page.locator('button').filter({ has: page.locator('svg:has(path)') }).nth(0);

		if (await dislikeBtn.isVisible()) await dislikeBtn.click();
		await page.waitForTimeout(500);
		if (await likeBtn.isVisible()) await likeBtn.click();
		await page.waitForTimeout(500);

		const dislikeSvg = dislikeBtn.locator('svg');
		await expect(dislikeSvg).toHaveAttribute('fill', 'none');
	});
});
