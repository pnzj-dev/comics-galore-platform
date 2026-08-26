// Shared Zod schemas for the auth forms (client-side validation).
import { z } from 'zod';

export const loginSchema = z.object({
	email: z.string().trim().email('Enter a valid email address'),
	password: z.string().min(1, 'Password is required'),
});

export const registerSchema = z
	.object({
		username: z
			.string()
			.trim()
			.min(3, 'At least 3 characters')
			.max(20, 'At most 20 characters')
			.regex(/^[a-z0-9](?:[_-]?[a-z0-9])*$/, 'Lowercase letters, numbers, and single - or _ in between'),
		email: z.string().trim().email('Enter a valid email address'),
		password: z.string().min(8, 'At least 8 characters'),
		confirm: z.string(),
		terms: z.boolean().refine((v) => v, 'You must agree to the terms'),
	})
	.refine((d) => d.password === d.confirm, {
		message: 'Passwords do not match',
		path: ['confirm'],
	});

export const forgotPasswordSchema = z.object({
	email: z.string().trim().email('Enter a valid email address'),
});

export const resetPasswordSchema = z
	.object({
		password: z.string().min(8, 'At least 8 characters'),
		confirm: z.string(),
	})
	.refine((d) => d.password === d.confirm, {
		message: 'Passwords do not match',
		path: ['confirm'],
	});
