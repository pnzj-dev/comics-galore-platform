// See https://svelte.dev/docs/kit/types#app.d.ts
// for information about these interfaces
declare global {
	namespace App {
		// interface Error {}
		// interface Locals {}
		// interface PageData {}
		// interface PageState {}
		// interface Platform {}
	}
}

import type { RowData } from '@tanstack/table-core';

declare module '@tanstack/table-core' {
	interface ColumnMeta<TData extends RowData, TValue> {
		filterType?: 'text' | 'select';
		filterOptions?: Array<{
			value: string;
			label: string;
		}>;
		filterPlaceholder?: string;
	}
}

export {};
