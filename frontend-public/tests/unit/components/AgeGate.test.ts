import { render, screen, fireEvent } from '@testing-library/svelte';
import AgeGate from '$lib/components/comics/AgeGate.svelte';
import { modal } from '$lib/stores/modal.svelte';
import { describe, it, expect, vi, beforeEach } from 'vitest';

const baseProps = {
	title: 'Mature Comic',
	author: 'Test Author',
	ageRating: 'mature',
	onConfirm: vi.fn(),
	onClose: vi.fn(),
};

beforeEach(() => {
	vi.clearAllMocks();
	modal.closeAll();
});

describe('AgeGate', () => {
	it('renders modal when open is true', () => {
		modal.open('agegate');
		render(AgeGate, baseProps);

		expect(screen.getByText('Age-Restricted Content')).toBeVisible();
		expect(screen.getByText('Mature Comic')).toBeVisible();
		expect(screen.getByText('by Test Author')).toBeVisible();
		expect(screen.getByText('mature')).toBeVisible();
	});

	it('does not render when open is false', () => {
		render(AgeGate, baseProps);

		expect(screen.queryByText('Age-Restricted Content')).not.toBeInTheDocument();
	});

	it('calls onConfirm when Continue is clicked', async () => {
		modal.open('agegate');
		render(AgeGate, baseProps);

		await fireEvent.click(screen.getByText("I'm 18+ years old, Continue"));

		expect(baseProps.onConfirm).toHaveBeenCalledTimes(1);
	});

	it('calls onClose when Go back is clicked', async () => {
		modal.open('agegate');
		render(AgeGate, baseProps);

		await fireEvent.click(screen.getByText('Go back'));

		expect(baseProps.onClose).toHaveBeenCalledTimes(1);
	});
});
