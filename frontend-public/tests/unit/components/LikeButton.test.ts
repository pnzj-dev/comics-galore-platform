import { render, screen, fireEvent } from '@testing-library/svelte';
import LikeButton from '$lib/components/LikeButton.svelte';
import { describe, it, expect, vi } from 'vitest';

describe('LikeButton', () => {
	it('renders count', () => {
		render(LikeButton, { active: false, count: 10 });
		expect(screen.getByText('10')).toBeVisible();
	});

	it('renders filled icon when active', () => {
		const { container } = render(LikeButton, { active: true, count: 5 });
		const svg = container.querySelector('svg');
		expect(svg?.getAttribute('fill')).toBe('currentColor');
	});

	it('renders outline icon when inactive', () => {
		const { container } = render(LikeButton, { active: false, count: 5 });
		const svg = container.querySelector('svg');
		expect(svg?.getAttribute('fill')).toBe('none');
	});

	it('calls onToggle on click', async () => {
		const spy = vi.fn();
		render(LikeButton, { active: false, count: 5, onToggle: spy });
		await fireEvent.click(screen.getByRole('button'));
		expect(spy).toHaveBeenCalledOnce();
	});

	it('disables button when loading', () => {
		render(LikeButton, { active: false, count: 5, loading: true });
		expect(screen.getByRole('button')).toBeDisabled();
	});

	it('count reflects bind value', () => {
		render(LikeButton, { active: false, count: 42 });
		expect(screen.getByText('42')).toBeVisible();
	});
});
