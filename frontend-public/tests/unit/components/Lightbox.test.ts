import { render, screen, fireEvent } from '@testing-library/svelte';
import Lightbox from '$lib/components/Lightbox.svelte';

const images = ['/img1.jpg', '/img2.jpg', '/img3.jpg'];

describe('Lightbox', () => {
	it('renders images and shows the first image', () => {
		const onClose = vi.fn();

		render(Lightbox, { images, open: true, onClose });

		expect(screen.getByRole('dialog')).toBeVisible();
		expect(screen.getByAltText('Image 1')).toBeVisible();
		expect(screen.getByText('1 / 3')).toBeVisible();
	});

	it('prev button navigates to previous image', async () => {
		const onClose = vi.fn();

		render(Lightbox, { images, open: true, startIndex: 1, onClose });

		expect(screen.getByText('2 / 3')).toBeVisible();

		await fireEvent.click(screen.getByLabelText('Previous'));

		expect(screen.getByText('1 / 3')).toBeVisible();
	});

	it('next button navigates to next image', async () => {
		const onClose = vi.fn();

		render(Lightbox, { images, open: true, startIndex: 0, onClose });

		expect(screen.getByText('1 / 3')).toBeVisible();

		await fireEvent.click(screen.getByLabelText('Next'));

		expect(screen.getByText('2 / 3')).toBeVisible();
	});

	it('Escape key closes lightbox', async () => {
		const onClose = vi.fn();

		render(Lightbox, { images, open: true, onClose });

		await fireEvent.keyDown(window, { key: 'Escape' });

		expect(onClose).toHaveBeenCalledTimes(1);
	});

	it('dot indicators render correct count', () => {
		const onClose = vi.fn();

		render(Lightbox, { images, open: true, onClose });

		const dots = screen.getAllByLabelText(/Image \d/);
		expect(dots).toHaveLength(3);
	});

	it('startIndex prop is respected', () => {
		const onClose = vi.fn();

		render(Lightbox, { images, open: true, startIndex: 2, onClose });

		expect(screen.getByText('3 / 3')).toBeVisible();
	});
});
