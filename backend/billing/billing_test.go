package billing

import (
	"context"
	"fmt"
	"testing"

	myauth "comics-galore/backend/auth"

	"encore.dev/beta/auth"
	"encore.dev/et"
)

type mockProvider struct {
	estimatePrice      func(ctx context.Context, req EstimateRequest) (*EstimateResponse, error)
	checkBalance       func(ctx context.Context, subPartnerID string) (map[string]BalanceEntry, error)
	createCustomer     func(ctx context.Context, name string) (string, error)
	createSubscription func(ctx context.Context, req SubscriptionRequest) (*SubscriptionResponse, error)
	createDeposit      func(ctx context.Context, req DepositRequest) (*DepositResponse, error)
}

func (m *mockProvider) EstimatePrice(ctx context.Context, req EstimateRequest) (*EstimateResponse, error) {
	if m.estimatePrice != nil {
		return m.estimatePrice(ctx, req)
	}
	return nil, fmt.Errorf("EstimatePrice not implemented")
}

func (m *mockProvider) CreateCustomer(ctx context.Context, name string) (string, error) {
	if m.createCustomer != nil {
		return m.createCustomer(ctx, name)
	}
	return "", fmt.Errorf("CreateCustomer not implemented")
}

func (m *mockProvider) CheckBalance(ctx context.Context, subPartnerID string) (map[string]BalanceEntry, error) {
	if m.checkBalance != nil {
		return m.checkBalance(ctx, subPartnerID)
	}
	return nil, nil
}

func (m *mockProvider) CreateSubscription(ctx context.Context, req SubscriptionRequest) (*SubscriptionResponse, error) {
	if m.createSubscription != nil {
		return m.createSubscription(ctx, req)
	}
	return nil, fmt.Errorf("CreateSubscription not implemented")
}

func (m *mockProvider) CreateDeposit(ctx context.Context, req DepositRequest) (*DepositResponse, error) {
	if m.createDeposit != nil {
		return m.createDeposit(ctx, req)
	}
	return nil, fmt.Errorf("CreateDeposit not implemented")
}

func setupBillingTables(t *testing.T) error {
	t.Helper()

	if _, err := db.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			sub_partner_id TEXT,
			tier TEXT NOT NULL DEFAULT 'free'
		)
	`); err != nil {
		return err
	}
	if _, err := db.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS tiers (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name TEXT NOT NULL
		)
	`); err != nil {
		return err
	}
	if _, err := db.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS plans (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			tier_id UUID NOT NULL,
			interval TEXT NOT NULL DEFAULT 'monthly',
			price_usd_cents INT NOT NULL DEFAULT 0,
			features JSONB DEFAULT '[]',
			is_active BOOLEAN NOT NULL DEFAULT true,
			provider_plan_id TEXT,
			provider_interval_days INT DEFAULT 0
		)
	`); err != nil {
		return err
	}
	return nil
}

func setMockProvider(t *testing.T, mp *mockProvider) {
	t.Helper()
	provider = mp
}

func authCtx(userID string) context.Context {
	ctx := context.Background()
	return auth.WithContext(ctx, auth.UID(userID), &myauth.AuthData{
		UserID: userID,
		Email:  "test@example.com",
		Role:   "user",
		Tier:   "free",
	})
}

func TestEstimatePrice(t *testing.T) {
	_, err := et.NewTestDatabase(context.Background(), "billingdb")
	if err != nil {
		t.Skipf("test database setup failed: %v", err)
	}
	if err := setupBillingTables(t); err != nil {
		t.Skipf("table setup failed: %v", err)
	}

	planID := "550e8400-e29b-41d4-a716-446655440010"
	tierID := "550e8400-e29b-41d4-a716-446655440020"
	_, err = db.Exec(context.Background(), `INSERT INTO tiers (id, name) VALUES ($1, 'Bronze')`, tierID)
	if err != nil {
		t.Skipf("insert tier failed: %v", err)
	}
	_, err = db.Exec(context.Background(), `
		INSERT INTO plans (id, tier_id, interval, price_usd_cents)
		VALUES ($1, $2, 'monthly', 500)
	`, planID, tierID)
	if err != nil {
		t.Skipf("insert plan failed: %v", err)
	}

	mp := &mockProvider{
		estimatePrice: func(ctx context.Context, req EstimateRequest) (*EstimateResponse, error) {
			return &EstimateResponse{
				EstimatedAmount: 0.01,
				FromCurrency:    "usd",
				ToCurrency:      "btc",
			}, nil
		},
	}
	setMockProvider(t, mp)

	ctx := authCtx("550e8400-e29b-41d4-a716-446655440030")
	resp, err := EstimatePrice(ctx, &EstimatePriceParams{
		PlanID: planID,
		Crypto: "btc",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.FromCurrency != "usd" {
		t.Errorf("expected from currency usd, got %s", resp.FromCurrency)
	}
	if resp.ToCurrency != "btc" {
		t.Errorf("expected to currency btc, got %s", resp.ToCurrency)
	}
}

func TestCreateSubscription_Valid(t *testing.T) {
	_, err := et.NewTestDatabase(context.Background(), "billingdb")
	if err != nil {
		t.Skipf("test database setup failed: %v", err)
	}
	if err := setupBillingTables(t); err != nil {
		t.Skipf("table setup failed: %v", err)
	}

	planID := "550e8400-e29b-41d4-a716-446655440011"
	tierID := "550e8400-e29b-41d4-a716-446655440021"
	userID := "550e8400-e29b-41d4-a716-446655440031"

	_, err = db.Exec(context.Background(), `INSERT INTO tiers (id, name) VALUES ($1, 'Bronze')`, tierID)
	if err != nil {
		t.Skipf("insert tier failed: %v", err)
	}
	_, err = db.Exec(context.Background(), `
		INSERT INTO plans (id, tier_id, interval, price_usd_cents, provider_plan_id)
		VALUES ($1, $2, 'monthly', 500, '12345')
	`, planID, tierID)
	if err != nil {
		t.Skipf("insert plan failed: %v", err)
	}
	_, err = db.Exec(context.Background(), `
		INSERT INTO users (id, sub_partner_id, tier) VALUES ($1, 'partner-1', 'free')
	`, userID)
	if err != nil {
		t.Skipf("insert user failed: %v", err)
	}

	mp := &mockProvider{
		createSubscription: func(ctx context.Context, req SubscriptionRequest) (*SubscriptionResponse, error) {
			return &SubscriptionResponse{
				SubscriptionID: "np-sub-001",
				Status:         "active",
			}, nil
		},
	}
	setMockProvider(t, mp)

	ctx := authCtx(userID)
	resp, err := CreateSubscription(ctx, &CreateSubParams{PlanID: planID})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.SubscriptionID == "" {
		t.Error("expected subscription ID, got empty")
	}
	if resp.Status != "active" {
		t.Errorf("expected status active, got %s", resp.Status)
	}
}

func TestPollSubscription_ReturnsActiveStatus(t *testing.T) {
	_, _ = et.NewTestDatabase(context.Background(), "billingdb")

	subID := "550e8400-e29b-41d4-a716-446655440040"
	_, err := db.Exec(context.Background(), `
		INSERT INTO subscriptions (id, user_id, plan_id, active, tier, provider_subscription_id)
		VALUES ($1, '550e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440002', true, 'bronze', 'np-sub-001')
	`, subID)
	if err != nil {
		t.Fatalf("insert subscription error: %v", err)
	}

	ctx := authCtx("550e8400-e29b-41d4-a716-446655440050")
	resp, err := PollSubscription(ctx, subID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Active {
		t.Error("expected active=true, got false")
	}
}

func TestPollSubscription_NotFound(t *testing.T) {
	_, _ = et.NewTestDatabase(context.Background(), "billingdb")

	ctx := authCtx("550e8400-e29b-41d4-a716-446655440050")
	_, err := PollSubscription(ctx, "00000000-0000-0000-0000-000000000000")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestCreateDeposit(t *testing.T) {
	_, err := et.NewTestDatabase(context.Background(), "billingdb")
	if err != nil {
		t.Skipf("test database setup failed: %v", err)
	}
	if err := setupBillingTables(t); err != nil {
		t.Skipf("table setup failed: %v", err)
	}

	planID := "550e8400-e29b-41d4-a716-446655440012"
	userID := "550e8400-e29b-41d4-a716-446655440032"

	_, err = db.Exec(context.Background(), `
		INSERT INTO plans (id, tier_id, interval, price_usd_cents)
		VALUES ($1, '550e8400-e29b-41d4-a716-446655440022', 'monthly', 1000)
	`, planID)
	if err != nil {
		t.Skipf("insert plan failed: %v", err)
	}
	_, err = db.Exec(context.Background(), `
		INSERT INTO users (id, sub_partner_id, tier) VALUES ($1, 'partner-1', 'free')
	`, userID)
	if err != nil {
		t.Skipf("insert user failed: %v", err)
	}

	mp := &mockProvider{
		createDeposit: func(ctx context.Context, req DepositRequest) (*DepositResponse, error) {
			return &DepositResponse{
				PaymentID:   "np-dep-001",
				PayAddress:  "0xabc123def456",
				PayAmount:   0.05,
				PayCurrency: "eth",
			}, nil
		},
	}
	setMockProvider(t, mp)

	ctx := authCtx(userID)
	resp, err := CreateDeposit(ctx, &CreateDepositParams{
		PlanID: planID,
		Crypto: "eth",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.DepositID == "" {
		t.Error("expected deposit ID, got empty")
	}
	if resp.PayAddress == "" {
		t.Error("expected pay address, got empty")
	}
	if resp.PayCurrency != "eth" {
		t.Errorf("expected pay currency eth, got %s", resp.PayCurrency)
	}
}

func TestPollDeposit(t *testing.T) {
	_, _ = et.NewTestDatabase(context.Background(), "billingdb")

	depID := "550e8400-e29b-41d4-a716-446655440060"
	_, err := db.Exec(context.Background(), `
		INSERT INTO deposits (id, user_id, currency_crypto, status)
		VALUES ($1, '550e8400-e29b-41d4-a716-446655440001', 'btc', 'completed')
	`, depID)
	if err != nil {
		t.Fatalf("insert deposit error: %v", err)
	}

	ctx := authCtx("550e8400-e29b-41d4-a716-446655440050")
	resp, err := PollDeposit(ctx, depID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Completed {
		t.Error("expected completed=true, got false")
	}
}

func TestPollDeposit_Pending(t *testing.T) {
	_, _ = et.NewTestDatabase(context.Background(), "billingdb")

	depID := "550e8400-e29b-41d4-a716-446655440061"
	_, err := db.Exec(context.Background(), `
		INSERT INTO deposits (id, user_id, currency_crypto, status)
		VALUES ($1, '550e8400-e29b-41d4-a716-446655440001', 'btc', 'pending')
	`, depID)
	if err != nil {
		t.Fatalf("insert deposit error: %v", err)
	}

	ctx := authCtx("550e8400-e29b-41d4-a716-446655440050")
	resp, err := PollDeposit(ctx, depID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Completed {
		t.Error("expected completed=false for pending deposit")
	}
}

func TestPollDeposit_NotFound(t *testing.T) {
	_, _ = et.NewTestDatabase(context.Background(), "billingdb")

	ctx := authCtx("550e8400-e29b-41d4-a716-446655440050")
	_, err := PollDeposit(ctx, "00000000-0000-0000-0000-000000000000")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
