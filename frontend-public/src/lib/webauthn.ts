// WebAuthn browser helpers: encode/decode credentials to/from the base64url
// JSON format the backend (go-webauthn) expects. Uses native
// navigator.credentials — no JS crypto library.

import type { JSONValue } from '$lib/api/encore-client';

export function webauthnSupported(): boolean {
	return typeof window !== 'undefined' && !!window.PublicKeyCredential;
}

export async function webauthnAutofillSupported(): Promise<boolean> {
	if (!window.PublicKeyCredential) return false;
	try {
		const pc = window.PublicKeyCredential as unknown as {
			isConditionalMediationAvailable?: () => Promise<boolean>;
		};
		return !!(await pc.isConditionalMediationAvailable?.());
	} catch {
		return false;
	}
}

function bufferToBase64url(buffer: ArrayBuffer): string {
	const bytes = new Uint8Array(buffer);
	let str = '';
	for (const b of bytes) str += String.fromCharCode(b);
	return btoa(str).replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '');
}

function base64urlToBuffer(value: string): ArrayBuffer {
	const b64 = value.replace(/-/g, '+').replace(/_/g, '/');
	const pad = b64.length % 4 ? '='.repeat(4 - (b64.length % 4)) : '';
	const bin = atob(b64 + pad);
	const bytes = new Uint8Array(bin.length);
	for (let i = 0; i < bin.length; i++) bytes[i] = bin.charCodeAt(i);
	return bytes.buffer;
}

// convertCreation encodes a PublicKeyCredential for the backend.
export function convertCreation(cred: PublicKeyCredential): JSONValue {
	const response = cred.response as AuthenticatorAttestationResponse;
	return {
		id: cred.id,
		rawId: bufferToBase64url(cred.rawId),
		type: cred.type,
		response: {
			clientDataJSON: bufferToBase64url(response.clientDataJSON),
			attestationObject: bufferToBase64url(response.attestationObject),
			transports: response.getTransports?.() ?? [],
		},
	} as unknown as JSONValue;
}

// convertAssertion encodes a PublicKeyCredential for the backend.
export function convertAssertion(cred: PublicKeyCredential): JSONValue {
	const response = cred.response as AuthenticatorAssertionResponse;
	return {
		id: cred.id,
		rawId: bufferToBase64url(cred.rawId),
		type: cred.type,
		response: {
			clientDataJSON: bufferToBase64url(response.clientDataJSON),
			authenticatorData: bufferToBase64url(response.authenticatorData),
			signature: bufferToBase64url(response.signature),
			userHandle: response.userHandle ? bufferToBase64url(response.userHandle) : null,
		},
	} as unknown as JSONValue;
}

// decodeOptions converts the backend's JSON options into WebAuthn API options
// (challenge, exclude/allow credentials and user.id are base64url-encoded).
export function decodeOptions<T extends { challenge: string }>(
	options: T,
): T & { challenge: ArrayBuffer } {
	return {
		...options,
		challenge: base64urlToBuffer(options.challenge),
	} as T & { challenge: ArrayBuffer };
}

// decodeCredentialList fixes base64url-encoded ArrayBuffer fields on credential
// descriptor lists and user handles in the raw options.
export function prepareCreationOptions(options: unknown): CredentialCreationOptions {
	const o = options as {
		publicKey: {
			challenge: string;
			user: { id: string };
			excludeCredentials?: { id: string }[];
		};
	};
	const pub = o.publicKey;
	return {
		publicKey: {
			...pub,
			challenge: base64urlToBuffer(pub.challenge),
			user: { ...pub.user, id: base64urlToBuffer(pub.user.id) },
			excludeCredentials: pub.excludeCredentials?.map((c) => ({
				...c,
				id: base64urlToBuffer(c.id),
			})),
		},
	} as CredentialCreationOptions;
}

export function prepareRequestOptions(options: unknown): CredentialRequestOptions {
	const o = options as {
		publicKey: {
			challenge: string;
			allowCredentials?: { id: string }[];
		};
	};
	const pub = o.publicKey;
	return {
		publicKey: {
			...pub,
			challenge: base64urlToBuffer(pub.challenge),
			allowCredentials: pub.allowCredentials?.map((c) => ({
				...c,
				id: base64urlToBuffer(c.id),
			})),
		},
	} as CredentialRequestOptions;
}

// mapWebauthnError converts a WebAuthn/DOM exception into a user-friendly
// message.
export function mapWebauthnError(err: unknown): string {
	const name = (err as { name?: string })?.name ?? '';
	switch (name) {
		case 'AbortError':
			return 'Passkey operation was cancelled.';
		case 'NotAllowedError':
			return 'Passkey operation was cancelled or timed out.';
		case 'NotSupportedError':
			return 'Passkeys are not supported by this browser.';
		case 'SecurityError':
			return 'This passkey could not be used on this origin.';
		case 'InvalidStateError':
			// For a login (get()) this means "A request is already pending".
			return 'Another passkey request is still in progress — please try again.';
		case 'UnknownError':
			return 'An unknown error occurred during passkey authentication.';
		default:
			break;
	}
	// Some browsers surface the "already pending" state under a different name;
	// match on the message so the raw browser string never leaks to the user.
	const message = (err as { message?: string })?.message ?? '';
	if (/already pending|in progress/i.test(message)) {
		return 'Another passkey request is still in progress — please try again.';
	}
	return message || 'Passkey operation failed.';
}
