import { type ClassValue, clsx } from 'clsx';
import { twMerge } from 'tailwind-merge';

export function cn(...inputs: ClassValue[]) {
	return twMerge(clsx(inputs));
}

export type WithElementRef<T, U extends HTMLElement = HTMLElement> = T & {
	ref?: U | null;
};

// Re-exported bits-ui utility types used by the shadcn-svelte components.
export type {
	WithoutChild,
	WithoutChildren,
	WithoutChildrenOrChild,
} from 'bits-ui';
