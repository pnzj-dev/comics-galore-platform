# Encore.ts + SvelteKit + SSE — Integration Guide

A reference for building real-time features where **Encore.ts** is the backend
and **SvelteKit** is the frontend, connected via Server-Sent Events (SSE).

---

## 1. Key fact to remember

Encore's built-in Streaming APIs (`api.streamOut`, `api.streamIn`,
`api.streamInOut`) are **WebSocket-based**, not SSE. If you specifically want
`text/event-stream` (so you can use the native `EventSource` API, or SSE-only
tooling), use Encore's **raw endpoint** (`api.raw`) instead — it gives you
direct access to the HTTP response so you can write SSE chunks by hand.

Two integration patterns are shown below:

- **Pattern A — Direct**: SvelteKit's `EventSource` connects straight to the
  Encore SSE endpoint (cross-origin, needs CORS).
- **Pattern B — Proxy**: SvelteKit's own `+server.ts` proxies the Encore SSE
  stream, so the browser only ever talks to SvelteKit (same-origin, no CORS).

---

## 2. Backend — Encore.ts raw SSE endpoint

```ts
// comments/comments.ts
import { api } from "encore.dev/api";

type Comment = { id: string; author: string; text: string; createdAt: number };

let comments: Comment[] = [];
const subscribers = new Set<(data: Comment[]) => void>();

function broadcast() {
  for (const send of subscribers) send(comments);
}

// SSE stream: GET /comments/stream
export const stream = api.raw(
  { expose: true, method: "GET", path: "/comments/stream" },
  async (req, resp) => {
    resp.writeHead(200, {
      "Content-Type": "text/event-stream",
      "Cache-Control": "no-cache",
      "Connection": "keep-alive",
      // Tighten this to your SvelteKit origin in production
      "Access-Control-Allow-Origin": "*",
    });

    const send = (data: Comment[]) => {
      resp.write(`data: ${JSON.stringify(data)}\n\n`);
    };

    subscribers.add(send);
    send(comments); // send current state immediately

    const heartbeat = setInterval(() => resp.write(":\n\n"), 15000);

    req.on("close", () => {
      clearInterval(heartbeat);
      subscribers.delete(send);
    });
  },
);

// POST /comments — add a comment and broadcast to all subscribers
export const add = api(
  { expose: true, method: "POST", path: "/comments" },
  async ({ author, text }: { author: string; text: string }): Promise<Comment> => {
    const comment: Comment = {
      id: crypto.randomUUID(),
      author,
      text,
      createdAt: Date.now(),
    };
    comments = [...comments, comment];
    broadcast();
    return comment;
  },
);
```

> In-memory `comments`/`subscribers` only work on a single instance. For
> multiple instances, back this with Redis pub/sub or Postgres
> `LISTEN/NOTIFY` and broadcast from each node.

---

## 3. Pattern A — SvelteKit connects directly (cross-origin)

```ts
// src/lib/comments.ts
import { writable } from 'svelte/store';

type Comment = { id: string; author: string; text: string; createdAt: number };

export const comments = writable<Comment[]>([]);

const ENCORE_URL = 'https://your-encore-app.encr.app';

export function connectComments() {
  const es = new EventSource(`${ENCORE_URL}/comments/stream`);

  es.onmessage = (event) => {
    comments.set(JSON.parse(event.data));
  };

  es.onerror = () => {
    console.warn('SSE connection error, browser will retry');
  };

  return () => es.close();
}

export async function postComment(author: string, text: string) {
  await fetch(`${ENCORE_URL}/comments`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ author, text }),
  });
}
```

Requires the `Access-Control-Allow-Origin` header on the Encore side (set
above) to match your SvelteKit app's origin.

---

## 4. Pattern B — SvelteKit proxies the stream (same-origin, no CORS)

```ts
// src/routes/api/comments/+server.ts
import type { RequestHandler } from './$types';

const ENCORE_URL = 'https://your-encore-app.encr.app';

export const GET: RequestHandler = async ({ fetch }) => {
  const upstream = await fetch(`${ENCORE_URL}/comments/stream`);

  // Pass the upstream SSE stream straight through to the browser
  return new Response(upstream.body, {
    headers: {
      'Content-Type': 'text/event-stream',
      'Cache-Control': 'no-cache',
      Connection: 'keep-alive',
    },
  });
};

export const POST: RequestHandler = async ({ request, fetch }) => {
  const body = await request.text();
  const res = await fetch(`${ENCORE_URL}/comments`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body,
  });
  return new Response(await res.text(), { status: res.status });
};
```

```ts
// src/lib/comments.ts
import { writable } from 'svelte/store';

export const comments = writable([]);

export function connectComments() {
  const es = new EventSource('/api/comments'); // same-origin now
  es.onmessage = (e) => comments.set(JSON.parse(e.data));
  return () => es.close();
}

export async function postComment(author: string, text: string) {
  await fetch('/api/comments', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ author, text }),
  });
}
```

---

## 5. Component usage (same for either pattern)

```svelte
<!-- src/routes/+page.svelte -->
<script lang="ts">
  import { onMount } from 'svelte';
  import { comments, connectComments, postComment } from '$lib/comments';

  let text = '';

  onMount(() => {
    const disconnect = connectComments();
    return disconnect;
  });

  async function submit() {
    if (!text.trim()) return;
    await postComment('You', text);
    text = '';
  }
</script>

<ul>
  {#each $comments as c (c.id)}
    <li><strong>{c.author}:</strong> {c.text}</li>
  {/each}
</ul>

<input bind:value={text} placeholder="Write a comment..."
       on:keydown={(e) => e.key === 'Enter' && submit()} />
<button on:click={submit}>Post</button>
```

---

## 6. Which pattern to pick

| | Pattern A (Direct) | Pattern B (Proxy) |
|---|---|---|
| CORS setup needed | Yes | No |
| Extra hop / latency | No | Small (SvelteKit → Encore) |
| Hides backend URL from browser | No | Yes |
| Works behind SvelteKit auth/session logic | No (unless duplicated) | Yes, naturally |
| Serverless SvelteKit adapters (Vercel, CF Workers) | Fine — Encore handles the long-lived connection | Check adapter's streaming support; some kill long-lived connections |

**Rule of thumb**: use **Pattern B** if you deploy SvelteKit on a
long-running Node server (adapter-node) and want a clean same-origin API. Use
**Pattern A** if your SvelteKit deployment target doesn't support long-lived
proxied streams (some serverless adapters don't) — let Encore hold the
connection directly instead.

---

## 7. Gotchas checklist

- [ ] `sveltekit-sse` (the npm package) is built for producing/consuming SSE
      *from SvelteKit's own routes* — it's not designed to proxy an external
      backend's stream. For Pattern A/B above, plain `EventSource` is simpler
      and avoids fighting its assumptions.
- [ ] Send a heartbeat comment (`:\n\n`) every ~15s from Encore to keep the
      connection alive through proxies/load balancers.
- [ ] `EventSource` auto-reconnects on drop by default — no extra code needed
      client-side for basic resilience.
- [ ] For multi-instance Encore deployments, broadcast via Redis/Postgres
      pub-sub, not an in-memory `Set`.
- [ ] Tighten `Access-Control-Allow-Origin` from `*` to your actual SvelteKit
      origin before shipping Pattern A.
