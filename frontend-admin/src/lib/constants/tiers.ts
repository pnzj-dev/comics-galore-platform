export const TIER_OPTIONS = [
	{ value: 'free', label: 'Free' },
	{ value: 'bronze', label: 'Bronze' },
	{ value: 'silver', label: 'Silver' },
	{ value: 'gold', label: 'Gold' },
	{ value: 'platinum', label: 'Platinum' },
];

export const PAID_TIER_OPTIONS = TIER_OPTIONS.filter((t) => t.value !== 'free');
