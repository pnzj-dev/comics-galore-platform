# ADR 0022 – Passkeys, Social OAuth & Opaque Sessions

## Status
Accepted

## Context
Comics Galore needed more authentication methods than email/password, plus
revocable sessions. The prior design (ADR 0003) used Encore Auth backed by
stateless JWTs stored in a non-HttpOnly cookie.

## Decision

1. **One user identity, many methods.** A single `users` row may have any
   combination of password credentials, WebAuthn passkeys, and linked OAuth
   accounts (`auth_accounts`). Linking is always by verified provider identity
   (`provider` + `provider_account_id`), never by email alone. A provider email
   matching an existing user does **not** auto-merge; a separate user is
   created with a NULL email to avoid account takeover.

2. **Opaque, revocable sessions.** JWTs are replaced by opaque random session
   ids stored in the `sessions` table. The Encore auth handler validates the
   session (expiry + revocation + ban/suspend) and returns the same
   `auth.UID`/`AuthData` as before. This enables logout, logout-all, and
   per-session revocation.

3. **HttpOnly session cookie.** The browser cookie (`token`, kept for
   compatibility) is now HttpOnly + SameSite=Lax (+ Secure in production). The
   browser can no longer read it, so authenticated browser calls route through
   a same-origin SvelteKit proxy (`/api/[...path]`) which forwards the session
   as a Bearer token to Encore. SSR guards use `resolveUser()` (calls
   `/auth/me`) instead of decoding the token client-side.

4. **Passkeys via go-webauthn.** Backend uses `github.com/go-webauthn/webauthn`;
   the browser uses native `navigator.credentials` (no JS crypto library).
   Challenges are single-use and short-lived, stored in `webauthn_challenges`.

5. **OAuth providers.** Google (OIDC), Facebook (Graph), Twitter/X (OAuth2 +
   PKCE), Apple (ES256 client-secret JWT). Each uses `golang.org/x/oauth2`.
   Flows use state + PKCE and are single-use (`oauth_states`). The browser
   receives a short-lived one-time code exchanged server-side for a session
   (`oauth_exchange_codes`), so the session id never appears in a URL.

6. **Recovery guard.** A user cannot remove their last usable authentication
   method (password, OAuth account, or passkey) without another fallback.

## Consequences
- `users.email` and `users.password_hash` are now nullable.
- `sveltekit` server code sets/reads the HttpOnly cookie; client-side
  `document.cookie` auth is removed.
- New secrets required: WebAuthn RPID/origins, and per-provider OAuth client
  credentials (see `docs/authentication.md`).
- This feature is **LATER** in `docs/v1-scope.md`; it is implemented early by
  explicit product decision.
