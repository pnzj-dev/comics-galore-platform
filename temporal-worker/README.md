# Temporal Worker — Comics Galore

Standalone Go program that hosts the subscription payment workflow. It is a
separate module (outside `backend/`, which Encore scans) because Temporal
workers are long-running processes and must not live inside an Encore service.

## Contract with the Encore backend

The worker and `backend/billing/temporal.go` share a contract (duplicated JSON
shapes, since they are separate Go modules):

- Workflow: `SubscriptionWorkflow`, task queue `subscription`.
- Workflow ID: the local subscription UUID.
- Signal: `payment_received` with payload `{ "status": string }`.
- Activities call Encore private endpoints:
  - `POST /billing/subscriptions/:id/activate`
  - `POST /billing/subscriptions/:id/expire`

## Local runbook

```bash
# 1. Temporal server (in a separate terminal)
temporal server start-dev

# 2. Encore backend (in another terminal)
cd backend && encore run

# 3. Worker (in a third terminal)
cd temporal-worker && go run ./worker
```

## Configuration (env vars)

| Var | Default | Purpose |
|---|---|---|
| `TEMPORAL_ADDRESS` | `localhost:7233` | Temporal frontend address |
| `TEMPORAL_NAMESPACE` | `default` | Temporal namespace |
| `ENCORE_BACKEND_URL` | `http://localhost:4000` | Encore base URL for activity callbacks |

Temporal Cloud (production, later) additionally needs mTLS `TEMPORAL_CERT` /
`TEMPORAL_KEY` and the namespace configured.
