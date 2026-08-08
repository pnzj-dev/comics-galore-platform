import path from 'path';
import { svelte } from '@sveltejs/vite-plugin-svelte';
import { defineConfig } from 'vitest/config';

export default defineConfig({
	plugins: [
		svelte({
			compilerOptions: {
				runes: ({ filename }) =>
					filename.split(/[/\\]/).includes('node_modules') ? undefined : true,
			},
		}),
	],
	resolve: {
		conditions: ['browser'],
		alias: {
			$lib: path.resolve(import.meta.dirname, 'src/lib'),
			'$app/navigation': path.resolve(import.meta.dirname, 'tests/unit/mocks/navigation.ts'),
		},
	},
	test: {
		environment: 'jsdom',
		include: ['tests/unit/**/*.{test,spec}.{ts,svelte}'],
		globals: true,
		setupFiles: ['tests/unit/setup.ts'],
	},
});