import { render, screen, fireEvent } from '@testing-library/svelte';
import DislikeButton from '$lib/components/DislikeButton.svelte';
import { describe, it, expect, vi } from 'vitest';

describe('DislikeButton', () => {
	it('renders count', () => {
		render(DislikeButton, { active: false, count: 10 });
		expect(screen.getByText('10')).toBeVisible();
	});

	it('renders filled icon when active', () => {
		const { container } = render(DislikeButton, { active: true, count: 5 });
		const svg = container.querySelector('svg');
		expect(svg?.getAttribute('fill')).toBe('currentColor');
	});

	it('renders outline icon when inactive', () => {
		const { container } = render(DislikeButton, { active: false, count: 5 });
		const svg = container.querySelector('svg');
		expect(svg?.getAttribute('fill')).toBe('none');
	});

	it('calls onToggle on click', async () => {
		const spy = vi.fn();
		render(DislikeButton, { active: false, count: 5, onToggle: spy });
		await fireEvent.click(screen.getByRole('button'));
		expect(spy).toHaveBeenCalledOnce();
	});

	it('disables button when loading', () => {
		render(DislikeButton, { active: false, count: 5, loading: true });
		expect(screen.getByRole('button')).toBeDisabled();
	});

	it('count reflects bind value', () => {
		render(DislikeButton, { active: false, count: 99 });
		expect(screen.getByText('99')).toBeVisible();
	});
});
