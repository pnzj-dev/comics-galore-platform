import { render, screen, fireEvent } from '@testing-library/svelte';
import FavoriteButton from '$lib/components/FavoriteButton.svelte';

const { mockedApi } = vi.hoisted(() => ({
	mockedApi: {
		get: vi.fn(),
		post: vi.fn(),
		put: vi.fn(),
		delete: vi.fn(),
		patch: vi.fn(),
	},
}));

vi.mock('$lib/api/client', () => ({
	api: mockedApi,
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
		mockedApi.post.mockResolvedValue({ favorited: true, fav_count: 4 });

		render(FavoriteButton, { comicId: 'comic-1', initialFavorited: false, initialCount: 3 });

		await fireEvent.click(screen.getByRole('button'));

		expect(mockedApi.post).toHaveBeenCalledWith('/comics/comic-1/favorite');
	});

	it('second click decrements and unfills', async () => {
		mockedApi.post.mockResolvedValue({ favorited: false, fav_count: 2 });

		render(FavoriteButton, { comicId: 'comic-1', initialFavorited: true, initialCount: 3 });

		await fireEvent.click(screen.getByRole('button'));

		expect(mockedApi.post).toHaveBeenCalledWith('/comics/comic-1/favorite');
	});

	it('disabled state prevents click during loading', async () => {
		mockedApi.post.mockImplementation(
			() => new Promise((resolve) => setTimeout(() => resolve({ favorited: true, fav_count: 4 }), 1000)),
		);

		render(FavoriteButton, { comicId: 'comic-1', initialFavorited: false, initialCount: 3 });

		await fireEvent.click(screen.getByRole('button'));

		const button = screen.getByRole('button');
		expect(button).toBeDisabled();
	});
});
