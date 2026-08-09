import Client, { Local } from '$lib/api/encore-client';

export const encore = new Client(Local, {
	fetcher: async (input, init) => {
		const m = document.cookie.match(/(?:^|;\s*)token=([^;]*)/);
		const token = m ? m[1] : null;
		const headers = new Headers(init?.headers);
		if (token) headers.set('Authorization', `Bearer ${token}`);
		return fetch(input, { ...init, headers });
	}
});
