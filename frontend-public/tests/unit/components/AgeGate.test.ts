import { render, screen, fireEvent } from '@testing-library/svelte';
import AgeGate from '$lib/components/AgeGate.svelte';

function mockSessionStorage() {
	const store: Record<string, string> = {};
	return {
		getItem: vi.fn((key: string) => store[key] ?? null),
		setItem: vi.fn((key: string, value: string) => {
			store[key] = value;
		}),
		removeItem: vi.fn((key: string) => {
			delete store[key];
		}),
	};
}

const mockGoto = vi.fn();

vi.mock('$app/navigation', () => ({
	goto: (url: string) => mockGoto(url),
}));

const defaultProps = {
	comicId: 'comic-1',
	title: 'Mature Comic',
	author: 'Test Author',
	ageRating: 'mature',
};

let sessionStorageMock: ReturnType<typeof mockSessionStorage>;

beforeEach(() => {
	vi.clearAllMocks();
	sessionStorageMock = mockSessionStorage();
	vi.stubGlobal('sessionStorage', sessionStorageMock);
});

describe('AgeGate', () => {
	it('renders modal when no session storage key exists', () => {
		render(AgeGate, defaultProps);

		expect(screen.getByText('Age-Restricted Content')).toBeVisible();
		expect(screen.getByText('Mature Comic')).toBeVisible();
		expect(screen.getByText('by Test Author')).toBeVisible();
		expect(screen.getByText('mature')).toBeVisible();
	});

	it('hides modal when session storage key is already present', () => {
		sessionStorageMock.getItem.mockReturnValueOnce('1');
		sessionStorageMock.getItem.mockReturnValueOnce('1');

		render(AgeGate, defaultProps);

		expect(screen.queryByText('Age-Restricted Content')).not.toBeInTheDocument();
	});

	it('confirm button sets session storage and hides the gate', async () => {
		render(AgeGate, defaultProps);

		expect(screen.getByText('Age-Restricted Content')).toBeVisible();

		await fireEvent.click(screen.getByText("I'm 18+ years old, Continue"));

		expect(sessionStorageMock.setItem).toHaveBeenCalledWith('age_gate_comic-1', '1');
	});

	it('Go back button navigates to home page', async () => {
		render(AgeGate, defaultProps);

		await fireEvent.click(screen.getByText('Go back'));

		expect(mockGoto).toHaveBeenCalledWith('/');
	});
});
