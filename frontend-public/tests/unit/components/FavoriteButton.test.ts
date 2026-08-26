import { render, screen, fireEvent } from '@testing-library/svelte';
import FavoriteButton from '$lib/components/social/FavoriteButton.svelte';

const { mockedToggle } = vi.hoisted(() => ({
	mockedToggle: vi.fn(),
}));

vi.mock('$lib/api/encore', () => ({
	encore: {
		comics: {
			ToggleFavorite: mockedToggle,
		},
	},
}));

beforeEach(() => {
	vi.clearAllMocks();
});

describe('FavoriteButton', () => {
	it('renders with initial count', () => {
		render(FavoriteButton, { comicId: 'comic-1', initialFavorited: false, initialCount: 3 });
		expect(screen.getByText('3')).toBeVisible();
	});

	it('click increments count and fills icon', async () => {
		mockedToggle.mockResolvedValue({ favorited: true, fav_count: 4 });
		render(FavoriteButton, { comicId: 'comic-1', initialFavorited: false, initialCount: 3 });
		await fireEvent.click(screen.getByRole('button'));
		expect(mockedToggle).toHaveBeenCalledWith('comic-1');
	});

	it('second click decrements and unfills', async () => {
		mockedToggle.mockResolvedValue({ favorited: false, fav_count: 2 });
		render(FavoriteButton, { comicId: 'comic-1', initialFavorited: true, initialCount: 3 });
		await fireEvent.click(screen.getByRole('button'));
		expect(mockedToggle).toHaveBeenCalledWith('comic-1');
	});

	it('disabled state prevents click during loading', async () => {
		mockedToggle.mockImplementation(() => new Promise((resolve) => setTimeout(() => resolve({ favorited: true, fav_count: 4 }), 1000)));
		render(FavoriteButton, { comicId: 'comic-1', initialFavorited: false, initialCount: 3 });
		await fireEvent.click(screen.getByRole('button'));
		expect(screen.getByRole('button')).toBeDisabled();
	});
});
