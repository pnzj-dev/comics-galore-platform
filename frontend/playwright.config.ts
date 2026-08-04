import { defineConfig, devices } from '@playwright/test';

export default defineConfig({
	testDir: './tests/e2e',
	timeout: 30000,
	retries: 1,
	use: {
		baseURL: 'http://localhost:5173',
		headless: true,
	},
	webServer: {
		command: 'cd .. && bun run dev',
		port: 5173,
		reuseExistingServer: true,
		timeout: 15000,
	},
	projects: [
		{ name: 'chromium', use: { ...devices['Desktop Chrome'] } },
	],
});
