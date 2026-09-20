# Authentication – Comics Galore

Authentication is delegated to **Logto** (OIDC). The Encore backend validates
Logto-issued tokens and maps each Logto identity to an internal `users` row
(which holds app-level attributes: role, tier, username, sub_partner_id,
ban/suspend, avatar, preferences).

## Methods (all in Logto)

- **Password** (email + password)
- **Social OAuth**: Google, Facebook, Twitter/X, Apple
- **Passkeys** (WebAuthn)
- **MFA** (TOTP authenticator + backup codes)
- **Password reset / forgot password**
- **Email verification**

Logto is the single source of truth for credentials and identity. Comics
Galore no longer stores passwords, passkeys, OAuth accounts, or MFA secrets.

## Architecture

```
        Password │ Social │ Passkey │ MFA │ Forgot/reset
                    └──────────┬──────────┘
                               ▼
                      Logto (OIDC identity provider)
                               │  issues ID token + access token + refresh token
                               ▼
        SvelteKit (@logto/sveltekit) — encrypted HttpOnly session cookie
                               │  forwards ID token as Authorization: Bearer
                               ▼
        Encore auth handler — validates token (JWKS + issuer) → sub
                               │  maps sub → users row (auto-provisions)
                               ▼
                    AuthData{ UserID, Email, Role, Tier }
```

### Token validation

`AuthHandler` (`backend/auth/auth.go`) validates the bearer token against the
Logto JWKS (`LogtoJWKSURI`) and issuer (`LogtoIssuer`). It extracts:

- `sub` → `users.logto_id` (provisioning a new user on first sign-in)
- `email` → `users.email` (fallback link for pre-existing/bootstrap accounts)

The role is read from `users.role` (Logto holds only identity + credentials).
`tier`, `sub_partner_id`, `username`, ban/suspend, avatar and preferences all
remain on `users`.

### Session model

Logto issues a refresh token (`offline_access` scope); `@logto/sveltekit`
stores an encrypted HttpOnly cookie holding the session. The access/id token is
rotated in the background. "Sign out everywhere" and device revocation are
available via the Logto Management API. There is no local `sessions` table.

## Roles (internal `users.role`)

Four roles: `user`, `uploader`, `moderator`, `admin`, stored on `users.role`
and enforced via `auth.Data().Role` throughout the app. Admin role changes are
made in the admin panel (`POST /admin/users/:id/role`). Logto holds no roles.

## Username / handle

Registration is handled by Logto; the public **username handle** (3–20
lowercase alphanumerics, single `_`/`-` inside) remains an app-level field on
`users.username`, validated live via `GET /auth/username-available` and set in
profile settings (`POST /me/username`).

## Bootstrap admin

`POST /auth/bootstrap` (gated by `BootstrapSecret`) creates the first
app-level `admin` row. The admin's Logto identity is linked on first sign-in
by email.

## Local development

1. **Logto** — use the `comics-galore-dev` tenant (`https://37bvfu.logto.app/`).
2. **Backend secrets** — `LogtoIssuer`, `LogtoJWKSURI` (and optionally
   `LogtoAudience`) via `encore secret set`.
3. **Frontend** — `.env` in `frontend-public/` / `frontend-admin/`:
   `LOGTO_ENDPOINT`, `LOGTO_APP_ID`, `LOGTO_APP_SECRET`,
   `LOGTO_COOKIE_ENCRYPTION_KEY`, plus `VITE_BACKEND_URL`.

## Removed (previously custom, now Logto)

- Password hashing (bcrypt), login/register endpoints
- Passkey endpoints (`go-webauthn`) and `passkeys` table
- Social OAuth endpoints (`auth_accounts`, `oauth_states`, `oauth_exchange_codes`)
- TOTP endpoints and `mfa_challenges`
- Email verification + password reset endpoints/tokens
- Opaque `sessions` table and admin impersonation (dropped)

App-level identity (`users`) keeps: `email`, `logto_id`, `role`, `tier`,
`username`, `sub_partner_id`, `avatar_key`, `banned_at`, `suspended_at`,
`email_verified_at`, notification preferences.
