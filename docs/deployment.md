# Deployment — Comics Galore

Production-grade deployment runbook for Comics Galore: Encore Cloud (Go backend) + Fly.io (two SvelteKit frontends) + Cloudflare (DNS) + GitHub Actions (CI/CD).

## Architecture

```
Encore Cloud (Go backend, app id "comics-galore-backend-v5k2")
  ├─ dev        → https://dev-comics-galore-backend-v5k2.encr.app
  ├─ staging    → https://staging-comics-galore-backend-v5k2.encr.app
  └─ production → https://production-comics-galore-backend-v5k2.encr.app   (confirm exact URL after first deploy)

Fly.io (Bun + adapter-node)
  ├─ cg-public-dev / cg-public-staging / cg-public-prod
  └─ cg-admin-dev  / cg-admin-staging  / cg-admin-prod

Fly.io (Temporal worker, always-on)
  └─ cg-worker-dev / cg-worker-staging / cg-worker-prod

Cloudflare Workers (MCP protocol layer)
  └─ cg-mcp-dev / cg-mcp-staging / cg-mcp-prod

Logto (OIDC identity) — one tenant per env
  ├─ dev     → https://37bvfu.logto.app/
  ├─ staging → https://kbkixu.logto.app/
  └─ prod    → https://g0f9pc.logto.app/
```

## Branch → environment

```
feature/* ──PR──▶ dev ──PR──▶ staging ──PR──▶ main
                   │             │              │
                 dev env      staging env    production
```

- `ci.yml` runs checks on PRs (backend `encore test`, frontend unit tests + `bun run build`).
- **Frontends** deploy via GitHub Actions → Fly.io (`deploy-dev.yml` / `deploy-staging.yml` / `deploy-prod.yml`) on push to `dev` / `staging` / `main`.
- **Backend** deploys via **Encore Cloud's git integration** (branch watching) on the same pushes — no GitHub Actions workflow needed.

## Environment matrix

| Env | Encore base URL | public Fly app | admin Fly app | Domain |
|---|---|---|---|---|
| dev | `https://dev-comics-galore-backend-v5k2.encr.app` | `cg-public-dev` | `cg-admin-dev` | `dev.comics-galore.com` + `dev-admin.comics-galore.com` |
| staging | `https://staging-comics-galore-backend-v5k2.encr.app` | `cg-public-staging` | `cg-admin-staging` | `staging.comics-galore.com` + `staging-admin.comics-galore.com` |
| prod | `https://production-comics-galore-backend-v5k2.encr.app` | `cg-public-prod` | `cg-admin-prod` | `comics-galore.com` + `admin.comics-galore.com` |

---

## 1. Backend — Encore Cloud

### 1.1 App link and environments

The app is already linked (`backend/encore.app` → `comics-galore-backend-v5k2`). The `dev`, `staging`, and `production` environments already exist — verify in the Encore Cloud dashboard (your app → **Environments**). If any is missing, create it there (or `encore env create <name>`).

### 1.2 Set secrets per environment

Repeat for `--env dev`, `--env staging`, and `--env production` (use sandbox/test keys for dev+staging, real keys for prod):

```bash
# auth (Logto — passwords/passkeys/social/MFA live in Logto, not here)
encore secret set --env <env> LogtoIssuer "https://<tenant>.logto.app/oidc"
encore secret set --env <env> LogtoJWKSURI "https://<tenant>.logto.app/oidc/jwks"
# encore secret set --env <env> LogtoAudience "<api-resource>"    # optional; enforces aud when set
encore secret set --env <env> BootstrapSecret "<value>"          # first-admin bootstrap token
encore secret set --env <env> ResendAPIKey "<value>"             # transactional email

# NowPayments (auth + billing + tiers)
encore secret set --env <env> NowPaymentsAPIKey "<value>"
encore secret set --env <env> NowPaymentsIPNKey "<value>"
encore secret set --env <env> NowPaymentsEmail "<value>"
encore secret set --env <env> NowPaymentsPassword "<value>"
encore secret set --env <env> NgrokURL ""                         # local-only; set empty for dev/staging/prod

# comics (AI moderation)
encore secret set --env <env> AIModeratorAPIKey "<value>"

# upload (Cloudflare Images)
encore secret set --env <env> CloudflareAccountID "<value>"
encore secret set --env <env> CloudflareAPIToken "<value>"
encore secret set --env <env> CloudflareImagesHash "<value>"

# turnstile
encore secret set --env <env> TurnstileSecret "<value>"
encore secret set --env <env> TurnstileHostnames "dev.comics-galore.com,dev-admin.comics-galore.com"   # per env

# Temporal worker auth + connection (see §5)
encore secret set --env <env> WorkerSecret "<value>"              # shared secret the worker sends as X-Worker-Token
encore secret set --env <env> TemporalAddress "<namespace>.tmprl.cloud:7233"
encore secret set --env <env> TemporalNamespace "<namespace>"
encore secret set --env <env> TemporalAPIKey "<api-key>"           # Temporal Cloud API key (preferred)
# ...or mTLS instead of the API key:
encore secret set --env <env> TemporalCert "<client-cert-pem>"     # Temporal Cloud mTLS
encore secret set --env <env> TemporalKey "<client-key-pem>"       # Temporal Cloud mTLS
```

#### Environment type vs named environment

Secrets set with `--env dev` / `--env staging` apply only to those named environments. The `dev` and `staging` environments are both `persistent` (development-type), so they **inherit** secrets set at the `development` type level (verify with `encore secret list` — the `Development` column). `BootstrapSecret` and `NgrokURL` were the only two not covered by that type, so they were set per-environment.

#### Production checklist (before the first `main` deploy)

The production environment does **not** inherit development secrets. Every secret currently shows `✗` under `Production`, so set **all** of them for production first (repeat the §1.2 commands with `--env production`):

```bash
encore secret set --env production LogtoIssuer "https://<prod-tenant>.logto.app/oidc"
encore secret set --env production LogtoJWKSURI "https://<prod-tenant>.logto.app/oidc/jwks"
encore secret set --env production NowPaymentsAPIKey "<value>"
encore secret set --env production NowPaymentsIPNKey "<value>"
encore secret set --env production NowPaymentsEmail "<value>"
encore secret set --env production NowPaymentsPassword "<value>"
encore secret set --env production NgrokURL ""                    # local-only; empty in prod
encore secret set --env production BootstrapSecret "<value>"      # first-admin bootstrap token
encore secret set --env production ResendAPIKey "<value>"
encore secret set --env production AIModeratorAPIKey "<value>"
encore secret set --env production CloudflareAccountID "<value>"
encore secret set --env production CloudflareAPIToken "<value>"
encore secret set --env production CloudflareImagesHash "<value>"
encore secret set --env production TurnstileSecret "<value>"
encore secret set --env production TurnstileHostnames "comics-galore.com,admin.comics-galore.com"
```

### 1.3 Deploy (automatic via git)

The backend deploys automatically through **Encore Cloud's GitHub integration**: push to a watched branch and Encore builds, tests, and deploys to the matching environment. There is no `encore deploy` command to run and no GitHub Actions workflow for the backend.

| Branch | Environment |
|---|---|
| `dev` | `dev` |
| `staging` | `staging` |
| `main` | `production` |

One-time setup (Encore Cloud dashboard → your app → **Settings → Git / integrations**):

1. Connect the GitHub repo, pointed at the `backend/` subdirectory (this is a monorepo; `encore.app` lives in `backend/`).
2. Map each environment to its source branch: `dev` → `dev`, `staging` → `staging`, `main` → `production`. The `main` → production watch is the one piece that may still be missing — confirm it is configured.

The frontends deploy separately via GitHub Actions → Fly.io (§2), triggered by the same branch pushes.

After each deploy, note the exact base URL (Encore Cloud shows it in the environment's settings). Update the `backend_url` in `deploy-*.yml` if it differs from `https://<env>-comics-galore-backend-v5k2.encr.app`.

### 1.4 Bootstrap the first admin (staging/prod)

```bash
curl -X POST https://<env>-comics-galore-backend-v5k2.encr.app/auth/bootstrap \
  -H 'Content-Type: application/json' \
  -d '{"token":"<BootstrapSecret>","email":"admin@comics-galore.com"}'
```

Creates the app-level `admin` row (role `admin`, tier `platinum`). Credentials live in Logto, so the admin's Logto identity is linked to this row by email on first sign-in. One-time only — a second call is rejected.

---

## 2. Frontends — Fly.io

### 2.1 One-time setup per app

For each of the 6 apps (`cg-public-dev`, `cg-public-staging`, `cg-public-prod`, `cg-admin-dev`, `cg-admin-staging`, `cg-admin-prod`):

```bash
# create the app (first deploy auto-creates it too)
fly apps create <app> --org <your-org>

# set the runtime backend URL (used by the /api proxy + server-side Encore client)
fly secrets set BACKEND_URL="https://<env>-comics-galore-backend-v5k2.encr.app" --app <app>
```

`fly deploy` (run by CI) then applies the staged secrets and deploys the image.

CI also stages the **Logto** runtime secrets per app (public vs admin get different
`LOGTO_APP_ID`/`LOGTO_APP_SECRET`; see §3.1) — `LOGTO_ENDPOINT`, `LOGTO_APP_ID`,
`LOGTO_APP_SECRET`, `LOGTO_COOKIE_ENCRYPTION_KEY`. These are read at runtime by
`hooks.server.ts` (`@logto/sveltekit`), not baked into the client bundle.

### 2.2 Cloudflare DNS + certificates

The zone `comics-galore.com` is managed in Cloudflare. Add records pointing each hostname at its Fly app:

```text
# dev
dev.comics-galore.com       CNAME cg-public-dev.fly.dev
dev-admin.comics-galore.com CNAME cg-admin-dev.fly.dev

# staging
staging.comics-galore.com        CNAME cg-public-staging.fly.dev
staging-admin.comics-galore.com  CNAME cg-admin-staging.fly.dev

# prod
comics-galore.com          CNAME cg-public-prod.fly.dev   (or A/AAAA from `fly ips list`)
admin.comics-galore.com    CNAME cg-admin-prod.fly.dev
```

- Set Cloudflare SSL/TLS mode to **Full (strict)**.
- Issue Fly certs: `fly certs create <hostname> --app <app>`. If Fly's HTTP-01 challenge can't complete through the Cloudflare proxy, grey-cloud the record (DNS only) until the cert is issued, then re-enable proxying.
- Add the public/admin hostnames to `TurnstileHostnames` and configure OAuth provider redirect URIs to point at the matching backend base URL.

---

## 3. GitHub Actions

### 3.1 Required repository secrets/variables

| Name | Kind | Value |
|---|---|---|
| `FLY_API_TOKEN` | Secret | Fly.io API token (`fly tokens create`) |
| `TURNSTILE_SITEKEY` | Secret (or Variable) | Cloudflare Turnstile sitekey (public value) |
| `ENCORE_AUTH_KEY` | Secret | Encore auth key (Encore Cloud → app → Settings → Auth Keys) — needed by `ci.yml` for `encore test` |
| `CLOUDFLARE_API_TOKEN` | Secret | Cloudflare API token with Workers "Edit" scope — for `deploy-mcp.yml` |
| `CLOUDFLARE_ACCOUNT_ID` | Secret (or Variable) | Cloudflare account ID (`b879240179ed3d643bf783745c93b100`) |

#### GitHub Environments (per-env Logto)

Create three environments — `dev`, `staging`, `prod` — each with the following
**variables** (non-secret) and **secrets**:

| Name | Kind | Value |
|---|---|---|
| `LOGTO_ENDPOINT` | Variable | `https://<tenant>.logto.app/` |
| `LOGTO_PUBLIC_APP_ID` | Variable | public web app ID |
| `LOGTO_ADMIN_APP_ID` | Variable | admin web app ID |
| `LOGTO_PUBLIC_APP_SECRET` | Secret | public web app secret (long-lived) |
| `LOGTO_ADMIN_APP_SECRET` | Secret | admin web app secret (long-lived) |
| `LOGTO_COOKIE_ENCRYPTION_KEY` | Secret | random 64-hex string |

The tenants/app IDs:

| Env | Tenant | Endpoint | public app | admin app |
|---|---|---|---|---|
| dev | `comics-galore-dev` (`37bvfu`) | `https://37bvfu.logto.app/` | `69j3uu7m990jznu2686br` | `1f9owrg4e5v6aelmkg98x` |
| staging | `comics-galore-staging` (`kbkixu`) | `https://kbkixu.logto.app/` | `ay893ezhvvifejz1d04kx` | `r0rd1vnpnauzl3pun39ip` |
| prod | `comics-galore-prod` (`g0f9pc`) | `https://g0f9pc.logto.app/` | `jwtdx17294seavse34bet` | `6a0376d2rbu3z22d5vowv` |

### 3.2 Workflows

- `.github/workflows/ci.yml` — PR checks (backend `encore test`, frontend unit tests + build, MCP worker typecheck).
- `.github/workflows/deploy-app.yml` — reusable frontend deploy (build args + `fly deploy`, plus Logto runtime secrets).
- `.github/workflows/deploy-worker.yml` — reusable Temporal worker deploy (`fly deploy` in `temporal-worker/`).
- `.github/workflows/deploy-mcp.yml` — reusable MCP worker deploy (`wrangler deploy` in `mcp-worker/`).
- `.github/workflows/deploy-dev.yml` / `deploy-staging.yml` / `deploy-prod.yml` — branch-triggered frontends **+ worker + MCP**.
- Backend deployment is **not** a GitHub Actions workflow — it's Encore Cloud's git integration (§1.3).

### 3.3 Build-time vs runtime config

- **Build args** (`--build-arg`, baked into the client bundle): `VITE_BACKEND_URL`, `VITE_API_URL`, `VITE_TURNSTILE_SITEKEY` — set per env by the workflow from `inputs.backend_url` + `TURNSTILE_SITEKEY`.
- **Runtime env** (Fly secret): `BACKEND_URL` — the Encore env base URL; read by the `/api/[...path]` proxy and the server-side Encore client. Defaults to `http://localhost:4000` (local dev).
- **Runtime env** (Fly secret): `LOGTO_ENDPOINT`, `LOGTO_APP_ID`, `LOGTO_APP_SECRET`, `LOGTO_COOKIE_ENCRYPTION_KEY` — read by `hooks.server.ts` (`@logto/sveltekit`); staged by CI from the environment's Logto variables/secrets (§3.1).

---

## 4. Temporal worker (Fly)

The worker is a separate always-on Go app (`temporal-worker/`), deployed per env as
`cg-worker-{dev,staging,prod}` via `deploy-worker.yml`. Per-app Fly secrets:

```bash
fly secrets set TEMPORAL_ADDRESS="<namespace>.tmprl.cloud:7233" \
  TEMPORAL_NAMESPACE="<namespace>" \
  TEMPORAL_API_KEY="<api-key>" \
  ENCORE_BACKEND_URL="https://<env>-comics-galore-backend-v5k2.encr.app" \
  WORKER_SECRET="<shared-secret>" \
  --app cg-worker-<env>
```

Use **either** `TEMPORAL_API_KEY` **or** the mTLS pair (`TEMPORAL_CERT` +
`TEMPORAL_KEY`); the API key takes precedence when both are set.

`WORKER_SECRET` must match the Encore `WorkerSecret` (§1.2); the worker sends it as
`X-Worker-Token` on every activity call.

---

## 4b. MCP worker (Cloudflare Workers)

The MCP protocol layer (`mcp-worker/`) is a Cloudflare Worker, deployed per env as
`cg-mcp-{dev,staging,prod}` via `deploy-mcp.yml` (`wrangler deploy`). It fronts the
Encore backend's key-gated `/mcp/*` endpoints.

- `BACKEND_URL` is set per deploy via `wrangler deploy --var BACKEND_URL=...` (§3.2).
- The MCP client authenticates with a bearer **MCP key** that the Worker forwards to
  Encore. Keys live (hashed) in the backend `mcpdb.mcp_keys` table, bound to a real
  user + role.

MCP keys are created and managed in the **admin panel** (`Admin → MCP Keys`):
- `POST /admin/mcp/keys` issues a key and returns the **full key once**; the
  bound user defaults to the creating admin (optionally any user).
- `GET /admin/mcp/keys` lists keys as a masked `…abcd` suffix (never the full key).
- `DELETE /admin/mcp/keys/:id` revokes a key.

Connect an MCP client to `https://cg-mcp-<env>.<account>.workers.dev/mcp` with
`Authorization: Bearer <mcp-key>`.

---

## 5. Verification

1. Bootstrap admin in each env (§1.4), then sign in on the matching admin domain.
2. `encore test ./...` green; both frontends `bun run build` green (CI runs these).
3. Confirm browser-direct calls work from the frontend origin: OAuth redirect, SSE live-comments, and avatar/media URLs (all hit the backend base URL directly — verify Encore public endpoints allow the frontend origin).
