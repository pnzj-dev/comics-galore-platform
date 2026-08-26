// Route protection helpers for the public SvelteKit app.
//
// The (app) route group (upload, favorites, lists, messages, support) is
// server-protected by its +layout.server.ts. This mirrors the same set of
// paths client-side so logout can decide whether to leave the page.

export const PROTECTED_PATHS = ['/upload', '/favorites', '/lists', '/messages', '/support'];

// isProtectedPath reports whether the given pathname belongs to a route that
// requires authentication. It matches exactly (with an optional trailing
// slash), so the public shared-list route /lists/[id] is NOT treated as
// protected.
export function isProtectedPath(pathname: string): boolean {
	const path = (pathname || '/').replace(/\/+$/, '') || '/';
	return PROTECTED_PATHS.includes(path);
}
