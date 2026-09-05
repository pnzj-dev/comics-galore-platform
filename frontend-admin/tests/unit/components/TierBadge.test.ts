import { render, screen } from '@testing-library/svelte';
import TierBadge from '$lib/components/TierBadge.svelte';
import { describe, it, expect } from 'vitest';

describe('TierBadge', () => {
	it.each([
		['free', 'Free'],
		['bronze', 'Bronze'],
		['silver', 'Silver'],
		['gold', 'Gold'],
		['platinum', 'Platinum'],
	])('renders "%s" label for tier "%s"', (tier, label) => {
		render(TierBadge, { tier });
		expect(screen.getByText(label)).toBeVisible();
	});

	it('defaults unknown tiers to Free', () => {
		render(TierBadge, { tier: 'diamond' });
		expect(screen.getByText('Free')).toBeVisible();
	});

	it('is case-insensitive', () => {
		render(TierBadge, { tier: 'GOLD' });
		expect(screen.getByText('Gold')).toBeVisible();
	});
});
