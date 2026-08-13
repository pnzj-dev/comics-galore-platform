import { isAgeConfirmed, confirmAge } from '$lib/ageGate';

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

let sessionStorageMock: ReturnType<typeof mockSessionStorage>;

beforeEach(() => {
	vi.clearAllMocks();
	sessionStorageMock = mockSessionStorage();
	vi.stubGlobal('sessionStorage', sessionStorageMock);
});

describe('ageGate', () => {
	it('returns false before confirmation', () => {
		expect(isAgeConfirmed('comic-1')).toBe(false);
	});

	it('returns true after confirmAge', () => {
		confirmAge('comic-1');
		expect(isAgeConfirmed('comic-1')).toBe(true);
		expect(sessionStorageMock.setItem).toHaveBeenCalledWith('age_gate_comic-1', '1');
	});

	it('scopes confirmation per comic', () => {
		confirmAge('comic-1');
		expect(isAgeConfirmed('comic-1')).toBe(true);
		expect(isAgeConfirmed('comic-2')).toBe(false);
	});
});
