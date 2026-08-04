import path from 'path';
import { defineConfig } from 'vitest/config';

export default defineConfig({
	resolve: {
		alias: {
			$lib: path.resolve(import.meta.dirname, 'src/lib')
		}
	},
	test: {
		environment: 'jsdom',
		include: ['tests/unit/**/*.{test,spec}.{ts,svelte}'],
		globals: true,
		setupFiles: ['tests/unit/setup.ts']
	}
});
