import Client, { Local } from '$lib/server/client';

export function getEncoreClient(token: string | undefined): Client {
	return new Client(Local, {
		fetcher: async (input, init) => {
			const headers = new Headers(init?.headers);
			if (token) headers.set('Authorization', `Bearer ${token}`);
			return fetch(input, { ...init, headers });
		}
	});
}
