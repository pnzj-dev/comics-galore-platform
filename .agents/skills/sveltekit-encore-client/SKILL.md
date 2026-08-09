---
name: sveltekit-encore-client
description: Generate and use the Encore TypeScript client in SvelteKit server-side load functions (+page.server.ts, +layout.server.ts). Replaces raw fetch() with typed, autocompleted API calls.
---

# Encore Generated Client — SvelteKit Server Integration

Use the Encore CLI to generate a **fully typed TypeScript client** from your
Go backend schema, then use it in SvelteKit's `+page.server.ts` and
`+layout.server.ts` load functions instead of raw `fetch()`.

---

## 1. Generate the client

Run from the backend directory:

```bash
cd backend
encore gen client comics-galore --output ./gen/client.ts --lang typescript
```

This produces a typed client with methods for every public API endpoint (e.g.
`client.comics.list(...)`, `client.auth.me(...)`, etc.). Copy or symlink the
output into your SvelteKit project's `$lib/server/`:

```bash
cp backend/gen/client.ts frontend-public/src/lib/server/client.ts
cp backend/gen/client.ts frontend-admin/src/lib/server/client.ts
```

**Regenerate** after any API change (new endpoints, renamed fields, etc.).
This is a manual step — the Encore CLI is not watched by your dev server.

---

## 2. Create the client factory

`$lib/server/encore.ts` — a factory that reads the JWT token from the cookie
and returns a typed, authenticated client:

```typescript
// frontend-public/src/lib/server/encore.ts
import Client, { Local } from '$lib/server/client';

export function getEncoreClient(token: string | undefined) {
    return new Client(Local, {
        auth: token ? { Authorization: `Bearer ${token}` } : undefined,
    });
}
```

For production, replace `Local` with your production API URL.

---

## 3. Use in a load function — Before / After

**Before** (raw `fetch` — what all 18 `+page.server.ts` files currently use):

```typescript
// +page.server.ts
const API = 'http://localhost:4000';

export const load: PageServerLoad = async ({ cookies, fetch }) => {
    const token = cookies.get('token');
    const headers = token ? { Authorization: `Bearer ${token}` } : {};

    const res = await fetch(`${API}/comics?limit=4&sort=newest`, { headers })
        .then(r => r.json());

    return { latestComics: res.comics || [] };
};
```

**After** (generated client):

```typescript
// +page.server.ts
import { getEncoreClient } from '$lib/server/encore';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ cookies }) => {
    const client = getEncoreClient(cookies.get('token'));

    const latest = await client.comics.list({ limit: 4, sort: 'newest' });

    return { latestComics: latest.comics ?? [] };
};
```

**Benefits**:
- No hardcoded API URLs — the client manages routing
- No manual `Authorization` header construction — passed via `auth` option
- Full TypeScript types on request params and response shapes
- Autocompletion in your editor for every endpoint

---

## 4. Authenticated vs Public endpoints

The client factory always accepts a token. For public endpoints (no auth
required), pass `undefined`:

```typescript
const client = getEncoreClient();  // no token — public endpoints only
const plans = await client.tiers.listPlans();
```

For authenticated admin endpoints, pass the JWT cookie:

```typescript
const client = getEncoreClient(cookies.get('token'));
const users = await client.auth.adminListUsers();
```

---

## 5. Complete homepage example

The homepage `+page.server.ts` currently does 4 parallel `fetch()` calls with
manual URL building and header construction. With the generated client:

```typescript
// frontend-public/src/routes/+page.server.ts
import { getEncoreClient } from '$lib/server/encore';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ cookies }) => {
    const token = cookies.get('token');
    const client = getEncoreClient(token);

    const [latest, popular, random] = await Promise.all([
        client.comics.list({ limit: 4, sort: 'newest' }),
        client.comics.list({ limit: 4, sort: 'popular' }),
        client.comics.list({ limit: 1, sort: 'random' }),
    ]);

    let continueReading: any[] = [];
    let continueProgress: Record<string, { current_page: number; total_pages: number }> = {};

    if (token) {
        try {
            const cr = await client.reading.continueReading();
            const items = cr.items || [];
            for (const item of items) {
                continueProgress[item.comic_id] = { current_page: item.current_page, total_pages: item.total_pages };
            }
            if (items.length > 0) {
                const ids = items.map((i: any) => i.comic_id);
                const batch = await client.comics.batchGet({ ids });
                continueReading = batch.comics || [];
            }
        } catch {}
    }

    return {
        latestComics: latest.comics || [],
        popularComics: popular.comics || [],
        comicOfTheDay: random.comics?.[0] || null,
        continueReading,
        continueProgress,
        authed: !!token,
    };
};
```

---

## 6. Caveats

- **Not SvelteKit's `fetch`**: The generated client makes its own HTTP requests
  internally. It does NOT use SvelteKit's `fetch` wrapper (which provides
  request coalescing, cookie forwarding, etc.). For Encore Go backends, the
  client uses standard `fetch()` which is fine for SSR.
- **Regenerate on API changes**: The client is a static file. When you add,
  rename, or remove API endpoints, regenerate the client and commit the
  updated `client.ts`.
- **No live reload**: During development, the generated client doesn't
  auto-update when you change the backend. Add a `postdev` script or
  regenerate manually after schema changes.
- **Type safety**: The generated client reflects your Encore API schema
  **at the time of generation**. If the backend schema changes without
  regenerating the client, TypeScript types will be stale. Always
  regenerate after backend changes.

---

## 7. File layout

```
frontend-public/src/lib/server/
├── client.ts          # Generated by `encore gen client`
├── encore.ts          # Factory — getEncoreClient(token?)
└── jwt.ts             # JWT decode helper
```

The `encore.ts` factory + `client.ts` together replace the `const API =
'http://localhost:4000'` + manual `headers` pattern in ALL `+page.server.ts`
and `+layout.server.ts` files.

---

## References

- `docs/architecture.md` — Data Loading section
- `.agents/skills/sveltekit-ui/SKILL.md` — SvelteKit component patterns
- `.agents/skills/encore-backend/SKILL.md` — Backend service patterns
- `https://encore.dev/docs/go/client-generation` — Encore Go client generation docs
