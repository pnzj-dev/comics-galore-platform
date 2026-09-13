# Temporal Worker — Comics Galore

Standalone Go program that hosts the payment workflows. It is a separate module
(outside `backend/`, which Encore scans) because Temporal workers are long-running
processes and must not live inside an Encore service.

## Workflows

- **`SubscribeWorkflow`** — combined checkout: funds via a deposit when the user
  has no balance, then creates the subscription and activates it.
- **`BoostWorkflow`** — quota boost: creates a boost deposit and grants the boost
  once it completes.

## Contract with the Encore backend

Task queue: `subscription`. Signals: `deposit_paid`, `subscription_paid`.

Activities call these Encore endpoints (all `public` + gated by `X-Worker-Token`):
- `POST /billing/internal/check-balance`
- `POST /billing/internal/create-deposit`
- `POST /billing/internal/create-subscription`
- `POST /billing/internal/create-boost-deposit`
- `POST /billing/subscriptions/:id/activate` / `expire`
- `POST /billing/deposits/:id/complete` / `expire`

## Local runbook

```bash
temporal server start-dev          # 1. Temporal
cd backend && encore run           # 2. Encore (localhost:4000)
cd temporal-worker && go run ./worker  # 3. worker
```

## Configuration (env vars)

| Var | Default | Purpose |
|---|---|---|
| `TEMPORAL_ADDRESS` | `localhost:7233` | Temporal frontend address |
| `TEMPORAL_NAMESPACE` | `default` | Temporal namespace |
| `TEMPORAL_CERT` / `TEMPORAL_KEY` | — | mTLS (Temporal Cloud) |
| `ENCORE_BACKEND_URL` | `http://localhost:4000` | Encore base URL for activity callbacks |
| `WORKER_SECRET` | — | shared secret sent as `X-Worker-Token` |

## Deployment (Fly.io)

Deployed via `.github/workflows/deploy-worker.yml` (one always-on app per env:
`cg-worker-dev`, `cg-worker-staging`, `cg-worker-prod`). Set per-app Fly secrets:

```
fly secrets set TEMPORAL_ADDRESS=... TEMPORAL_NAMESPACE=... \
  TEMPORAL_CERT="$(cat client.pem)" TEMPORAL_KEY="$(cat client.key)" \
  ENCORE_BACKEND_URL=https://<env>-comics-galore-backend-v5k2.encr.app \
  WORKER_SECRET=... --app cg-worker-<env>
```
