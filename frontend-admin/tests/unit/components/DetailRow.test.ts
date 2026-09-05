import { render, screen } from '@testing-library/svelte';
import DetailRow from '$lib/components/DetailRow.svelte';
import { describe, it, expect } from 'vitest';

describe('DetailRow', () => {
	it('renders label and value', () => {
		render(DetailRow, { label: 'Email', value: 'admin@example.com' });
		expect(screen.getByText('Email')).toBeVisible();
		expect(screen.getByText('admin@example.com')).toBeVisible();
	});

	it('renders an em dash for a null value', () => {
		render(DetailRow, { label: 'Email', value: null });
		expect(screen.getByText('—')).toBeVisible();
	});

	it('renders an em dash for an undefined value', () => {
		render(DetailRow, { label: 'Email', value: undefined });
		expect(screen.getByText('—')).toBeVisible();
	});
});
