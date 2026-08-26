export const languageOptions = [
	{ value: 'en', label: 'English' },
	{ value: 'ja', label: 'Japanese' },
	{ value: 'es', label: 'Spanish' },
	{ value: 'ko', label: 'Korean' },
	{ value: 'fr', label: 'French' },
];

export const ageRatingOptions = [
	{ value: 'all_ages', label: 'All Ages' },
	{ value: 'teen', label: 'Teen' },
	{ value: 'mature', label: 'Mature' },
	{ value: 'explicit', label: 'Explicit' },
];

export const readingDirectionOptions = [
	{ value: 'ltr', label: 'Left to Right (Western)' },
	{ value: 'rtl', label: 'Right to Left (Manga)' },
];

export function optionLabel(options: { value: string; label: string }[], value: string): string {
	return options.find((o) => o.value === value)?.label ?? value;
}
