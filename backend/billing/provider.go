package billing

import "comics-galore/backend/nowpayments"

// Type aliases re-export the shared NowPayments types so the billing package
// (and its tests / generated interface) keep a stable surface while the client
// lives in the shared nowpayments package.

type EstimateRequest = nowpayments.EstimateRequest
type EstimateResponse = nowpayments.EstimateResponse
type BalanceEntry = nowpayments.BalanceEntry
type SubscriptionRequest = nowpayments.SubscriptionRequest
type SubscriptionResponse = nowpayments.SubscriptionResponse
type DepositRequest = nowpayments.DepositRequest
type DepositResponse = nowpayments.DepositResponse
type CreatePlanRequest = nowpayments.CreatePlanRequest
type CreatePlanResponse = nowpayments.CreatePlanResponse
type PaymentsProvider = nowpayments.PaymentsProvider
