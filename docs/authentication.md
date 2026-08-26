# Authentication – Comics Galore

## Methods

A single user identity supports multiple sign-in methods:

- **Password** (email + password, bcrypt-hashed)
- **Passkeys** (WebAuthn, via `github.com/go-webauthn/webauthn`)
- **Social OAuth**: Google, Facebook, Twitter/X, Apple

All methods converge on the same opaque session (stored in `sessions`) and the
same Encore auth handler (`//encore:authhandler` → `auth.UID` / `AuthData`).

## Username / handle

Registration requires a unique **username** (public handle):

- Format: `3–20` characters, lowercase `a-z0-9`, with single `_`/`-` allowed
  **only between** alphanumerics (no leading/trailing/consecutive). Regex
  `^[a-z0-9](?:[_-]?[a-z0-9])*$` + length check.
- Stored on `users.username` (nullable unique; OAuth-created and pre-existing
  accounts have `NULL`).
- Validated **live** in the register form (client regex + debounced
  `GET /auth/username-available` which returns `{available, valid, message}`),
  and re-validated server-side in `Register`.

## Auth UX (modals + minimal login page)

- **Login / Register / Forgot-password** are **modals** (`LoginModal`,
  `RegisterModal`, `ForgotPasswordModal`) opened via the shared `modal` store
  from the nav and home/pricing CTAs. They share a branded `AuthCard` header and
  password visibility toggles.
- A standalone **`/login` page** remains (minimal shell, no nav/footer) as the
  redirect target for auth-gated routes; `/register` is modal-only.
- Forgot-password (`/auth/password-reset/request`) is reachable from the login
  modal; the email link still lands on `/auth/reset-password`.

## Architecture

```
 Password │ Passkey │ Google/Facebook/Twitter/Apple
     └────────────┬─────────────┘
                  ▼
            users (single identity)
                  │  password_hash (nullable)
                  │  auth_accounts (linked providers)
                  │  passkeys (WebAuthn credentials)
                  ▼
             sessions (opaque, revocable)
                  ▼
        HttpOnly cookie on SvelteKit domain
                  ▼
      SvelteKit server → Encore (Bearer session)
                  ▼
         Encore auth handler → auth.UID
```

## Local development

1. **Database** — Encore runs PostgreSQL automatically (`encore run`).
2. **Secrets** — set via `encore secret set --type local <Name>` or a
   `.secrets.local.cue` file (gitignored) at the repo root:

```cue
WebAuthnRPID: "localhost"
WebAuthnOrigins: "http://localhost:5173,http://localhost:5174"
FrontendURL: "http://localhost:5173"
GoogleClientID: "..."
GoogleClientSecret: "..."
// ... etc.
```

3. **Frontend** — copy `.env.example` to `.env` in `frontend-public/` and
   `frontend-admin/`. `VITE_BACKEND_URL` points at the Encore backend
   (`http://localhost:4000`).

## Passkey requirements

- **RP ID**: the domain the passkey is bound to. `localhost` in dev; the real
  domain in production (`comicsgalore.com`).
- **Origin**: the frontend origin(s) allowed. Set `WebAuthnOrigins`.
- **HTTPS**: required in production (WebAuthn only works on `localhost` or
  HTTPS). Production must use HTTPS.
- **Browser support**: passkeys work in all modern browsers; the UI
  feature-detects `window.PublicKeyCredential` and conditional (autofill)
  mediation.

### Adding another passkey

Settings → Security → "Add a passkey". The browser prompts for a name and
performs `navigator.credentials.create()`; the credential is stored server-side
with a friendly name and can be removed later.

### Login with a passkey

The login page shows "Continue with Passkey" and silently attempts autofill
(conditional mediation) where supported. The browser performs
`navigator.credentials.get()` and the assertion is verified server-side.

## Account linking

- A logged-in user can connect Google/Facebook/X/Apple from
  Settings → Security. The backend marks link intent; the provider identity is
  bound to the current user.
- Linking an OAuth account already bound to another user is rejected.
- Unlinking/removing a method requires at least one other sign-in method
  remains (recovery guard).

## OAuth callback URLs

### Development

| Provider | Redirect URI |
|----------|--------------|
| Google | `http://localhost:4000/auth/oauth/google/callback` |
| Facebook | `http://localhost:4000/auth/oauth/facebook/callback` |
| Twitter/X | `http://localhost:4000/auth/oauth/twitter/callback` |
| Apple | `http://localhost:4000/auth/oauth/apple/callback` |

### Production

Replace the host with the Encore backend domain, e.g.
`https://staging-comics-galore-backend-v5k2.encr.app/auth/oauth/google/callback`
(or your custom API domain).

## OAuth provider configuration

### Google
- Client ID / Client Secret from Google Cloud Console (OAuth 2.0 Client).
- Scopes: `openid email profile`.
- Authorized redirect URI: see above.

### Facebook
- App ID / App Secret from developers.facebook.com.
- Scopes: `email public_profile`.
- Valid OAuth redirect URI required in Facebook Login settings.

### Twitter/X
- Client ID / Client Secret from the X Developer Portal (OAuth 2.0 app).
- Scopes: `users.read tweet.read`.
- Uses PKCE; callback URL must be registered.

### Apple
- Service ID / Client ID, Team ID, Key ID, and a private key (`.p8`) from the
  Apple Developer portal.
- Apple uses a generated ES256 client-secret JWT (not a static secret).
- Requires the "Sign in with Apple" capability and a registered return URL.

## Environment variables / secrets

| Secret | Purpose |
|--------|---------|
| `WebAuthnRPID` | Relying Party ID (defaults to `localhost` in dev) |
| `WebAuthnOrigins` | Comma-separated allowed origins |
| `FrontendURL` | Public frontend origin (for OAuth redirects) |
| `GoogleClientID` / `GoogleClientSecret` | Google OAuth |
| `FacebookClientID` / `FacebookClientSecret` | Facebook OAuth |
| `TwitterClientID` / `TwitterClientSecret` | Twitter/X OAuth |
| `AppleClientID` / `AppleTeamID` / `AppleKeyID` / `ApplePrivateKey` | Apple Sign In |

Frontend (`.env`): `VITE_BACKEND_URL`, `VITE_API_URL`.
