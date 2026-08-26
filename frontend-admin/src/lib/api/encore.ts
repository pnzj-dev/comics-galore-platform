import Client from '$lib/api/encore-client';

// The browser client routes through the same-origin SvelteKit proxy
// (/api/[...path]) because the session cookie is HttpOnly and cannot be read
// by client-side JS.
export const encore = new Client('/api', {});
