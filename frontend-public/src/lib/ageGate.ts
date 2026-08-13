// Age-gate persistence. Confirmation is stored per-comic in sessionStorage so
// the mature-content warning does not reappear for the reader or download after
// the user has acknowledged it once this session.

const key = (comicId: string) => `age_gate_${comicId}`;

export function isAgeConfirmed(comicId: string): boolean {
	return typeof sessionStorage !== 'undefined' && sessionStorage.getItem(key(comicId)) === '1';
}

export function confirmAge(comicId: string): void {
	sessionStorage.setItem(key(comicId), '1');
}
