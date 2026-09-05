import { render, screen } from '@testing-library/svelte';
import KpiCard from '$lib/components/KpiCard.svelte';
import { describe, it, expect } from 'vitest';

describe('KpiCard', () => {
	it('renders title and value', () => {
		render(KpiCard, { title: 'Total Users', value: '1,234' });
		expect(screen.getByText('Total Users')).toBeVisible();
		expect(screen.getByText('1,234')).toBeVisible();
	});

	it('renders a hint when provided', () => {
		render(KpiCard, { title: 'Revenue', value: '$5k', hint: '+12% this month' });
		expect(screen.getByText('+12% this month')).toBeVisible();
	});

	it('omits the hint when not provided', () => {
		const { container } = render(KpiCard, { title: 'Revenue', value: '$5k' });
		expect(container.textContent).not.toContain('hint');
	});

	it('applies the accent styling when accent is true', () => {
		const { container } = render(KpiCard, { title: 'X', value: '1', accent: true });
		const root = container.querySelector('div');
		expect(root?.className).toContain('border-primary/30');
	});
});
