import { cn } from '$lib/utils';

describe('cn', () => {
	it('merges multiple string classes', () => {
		const result = cn('bg-red-500', 'text-white', 'p-4');

		expect(result).toContain('bg-red-500');
		expect(result).toContain('text-white');
		expect(result).toContain('p-4');
	});

	it('handles conditional classes via object syntax', () => {
		const result = cn('base', {
			'bg-blue-500': true,
			'text-lg': false,
			'hidden': false,
			'flex': true
		});

		expect(result).toContain('base');
		expect(result).toContain('bg-blue-500');
		expect(result).toContain('flex');
		expect(result).not.toContain('text-lg');
		expect(result).not.toContain('hidden');
	});

	it('handles array of classes', () => {
		const result = cn(['px-4', 'py-2'], 'rounded', ['shadow']);

		expect(result).toContain('px-4');
		expect(result).toContain('py-2');
		expect(result).toContain('rounded');
		expect(result).toContain('shadow');
	});

	it('resolves conflicting Tailwind classes with twMerge (last wins)', () => {
		const result = cn('px-4', 'px-2');

		expect(result).toContain('px-2');
		expect(result).not.toContain('px-4');

		const result2 = cn('text-red-500', 'text-blue-600');

		expect(result2).toContain('text-blue-600');
		expect(result2).not.toContain('text-red-500');
	});

	it('resolves bg-color conflicts with twMerge', () => {
		const result = cn('bg-red-500', 'bg-blue-500', 'bg-green-500');

		expect(result).toContain('bg-green-500');
		expect(result).not.toContain('bg-red-500');
		expect(result).not.toContain('bg-blue-500');
	});

	it('filters out falsy values', () => {
		const result = cn('base', false && 'should-not-appear', undefined, null, 'valid');

		expect(result).toContain('base');
		expect(result).toContain('valid');
		expect(result).not.toContain('should-not-appear');
	});

	it('handles empty input gracefully', () => {
		const result = cn();

		expect(result).toBe('');
	});

	it('handles a mix of strings, objects, arrays, and falsy values', () => {
		const result = cn(
			'base',
			{ 'variant-a': true, 'variant-b': false },
			['nested-1', 'nested-2'],
			null,
			undefined,
			false && 'gone',
			'last'
		);

		expect(result).toContain('base');
		expect(result).toContain('variant-a');
		expect(result).not.toContain('variant-b');
		expect(result).toContain('nested-1');
		expect(result).toContain('nested-2');
		expect(result).toContain('last');
		expect(result).not.toContain('gone');
	});

	it('merges Tailwind utility conflicts correctly (font-size, margin, padding)', () => {
		const result = cn('text-sm', 'text-lg', 'm-2', 'm-4', 'p-1', 'p-3');

		expect(result).toContain('text-lg');
		expect(result).not.toContain('text-sm');
		expect(result).toContain('m-4');
		expect(result).not.toContain('m-2');
		expect(result).toContain('p-3');
		expect(result).not.toContain('p-1');
	});
});
