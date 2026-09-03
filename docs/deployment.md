# Deployment — Comics Galore

Production-grade deployment runbook for Comics Galore: Encore Cloud (Go backend) + Fly.io (two SvelteKit frontends) + Cloudflare (DNS) + GitHub Actions (CI/CD).

## Architecture

```
Encore Cloud (Go backend, app id "comics-galore-backend-v5k2", slug "3cxq6")
  ├─ dev        → https://dev-3cxq6.encr.app
  ├─ staging    → https://staging-3cxq6.encr.app
  └─ production → https://production-3cxq6.encr.app   (confirm exact URL after first deploy)

Fly.io (Bun + adapter-node)
  ├─ cg-public-dev / cg-public-staging / cg-public-prod
  └─ cg-admin-dev  / cg-admin-staging  / cg-admin-prod
```

## Branch → environment

```
feature/* ──PR──▶ dev ──PR──▶ staging ──PR──▶ main
                   │             │              │
                 dev env      staging env    production
```

- `ci.yml` runs checks on PRs (backend `encore test`, frontend `bun run build`).
- `deploy-dev.yml` / `deploy-staging.yml` / `deploy-prod.yml` deploy on push to `dev` / `staging` / `main`.

## Environment matrix

| Env | Encore base URL | public Fly app | admin Fly app | Domain |
|---|---|---|---|---|
| dev | `https://dev-3cxq6.encr.app` | `cg-public-dev` | `cg-admin-dev` | `dev.comics-galore.com` + `dev-admin.comics-galore.com` |
| staging | `https://staging-3cxq6.encr.app` | `cg-public-staging` | `cg-admin-staging` | `staging.comics-galore.com` + `staging-admin.comics-galore.com` |
| prod | `https://production-3cxq6.encr.app` | `cg-public-prod` | `cg-admin-prod` | `comics-galore.com` + `admin.comics-galore.com` |

---

## 1. Backend — Encore Cloud

### 1.1 Link the app and create environments

```bash
cd backend
encore app link               # connect to the "comics-galore-backend-v5k2" app
encore env create dev         # production exists by default
encore env create staging
```

### 1.2 Set secrets per environment

Repeat for `--env dev`, `--env staging`, and `--env production` (use sandbox/test keys for dev+staging, real keys for prod):

```bash
# auth
encore secret set --env <env> JWTSecret "<value>"
encore secret set --env <env> BootstrapSecret "<value>"          # first-admin bootstrap token
encore secret set --env <env> FrontendURL "https://<public-domain>"
encore secret set --env <env> WebAuthnRPID "<domain>"
encore secret set --env <env> WebAuthnOrigins "https://<public-domain>"
encore secret set --env <env> ResendAPIKey "<value>"
encore secret set --env <env> GoogleClientID "<value>"
encore secret set --env <env> GoogleClientSecret "<value>"
# ... Facebook/Twitter/Apple OAuth as configured

# NowPayments (auth + billing + tiers)
encore secret set --env <env> NowPaymentsAPIKey "<value>"
encore secret set --env <env> NowPaymentsIPNKey "<value>"
encore secret set --env <env> NowPaymentsEmail "<value>"
encore secret set --env <env> NowPaymentsPassword "<value>"

# comics (AI moderation)
encore secret set --env <env> AIModeratorAPIKey "<value>"

# upload (Cloudflare Images)
encore secret set --env <env> CloudflareAccountID "<value>"
encore secret set --env <env> CloudflareAPIToken "<value>"
encore secret set --env <env> CloudflareImagesHash "<value>"

# turnstile
encore secret set --env <env> TurnstileSecret "<value>"
encore secret set --env <env> TurnstileHostnames "dev.comics-galore.com,dev-admin.comics-galore.com"   # per env
```

### 1.3 Deploy

```bash
encore deploy --env dev
encore deploy --env staging
encore deploy --env production
```

After each deploy, note the exact base URL (Encore prints it; `encore app env list` shows it too). Update the CI `backend_url` in the deploy workflows if it differs from `https://<env>-3cxq6.encr.app`.

### 1.4 Bootstrap the first admin (staging/prod)

```bash
curl -X POST https://<env>-3cxq6.encr.app/auth/bootstrap \
  -H 'Content-Type: application/json' \
  -d '{"token":"<BootstrapSecret>","email":"admin@comics-galore.com","password":"<strong-password>"}'
```

Returns the created admin (role `admin`, tier `platinum`). One-time only — a second call is rejected.

---

## 2. Frontends — Fly.io

### 2.1 One-time setup per app

For each of the 6 apps (`cg-public-dev`, `cg-public-staging`, `cg-public-prod`, `cg-admin-dev`, `cg-admin-staging`, `cg-admin-prod`):

```bash
# create the app (first deploy auto-creates it too)
fly apps create <app> --org <your-org>

# set the runtime backend URL (used by the /api proxy + server-side Encore client)
fly secrets set BACKEND_URL="https://<env>-3cxq6.encr.app" --app <app>
```

`fly deploy` (run by CI) then applies the staged secret and deploys the image.

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

### 3.2 Workflows

- `.github/workflows/ci.yml` — PR checks.
- `.github/workflows/deploy-app.yml` — reusable deploy (build args + `fly deploy`).
- `.github/workflows/deploy-dev.yml` / `deploy-staging.yml` / `deploy-prod.yml` — branch-triggered callers.

### 3.3 Build-time vs runtime config

- **Build args** (`--build-arg`, baked into the client bundle): `VITE_BACKEND_URL`, `VITE_API_URL`, `VITE_TURNSTILE_SITEKEY` — set per env by the workflow from `inputs.backend_url` + `TURNSTILE_SITEKEY`.
- **Runtime env** (Fly secret): `BACKEND_URL` — the Encore env base URL; read by the `/api/[...path]` proxy and the server-side Encore client. Defaults to `http://localhost:4000` (local dev).

---

## 4. Verification

1. Bootstrap admin in each env (§1.4), then sign in on the matching admin domain.
2. `encore test ./...` green; both frontends `bun run build` green (CI runs these).
3. Confirm browser-direct calls work from the frontend origin: OAuth redirect, SSE live-comments, and avatar/media URLs (all hit the backend base URL directly — verify Encore public endpoints allow the frontend origin).
