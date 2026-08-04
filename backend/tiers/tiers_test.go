package tiers

import (
	"context"
	"testing"

	"encore.dev/et"
)

func TestListTiers_ReturnsSortedWithPlatinumLast(t *testing.T) {
	_, _ = et.NewTestDatabase(context.Background(), "tiersdb")
	ctx := context.Background()

	resp, err := ListTiers(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Tiers) == 0 {
		t.Fatal("expected tiers, got empty list")
	}

	for i := 1; i < len(resp.Tiers); i++ {
		if resp.Tiers[i].SortOrder < resp.Tiers[i-1].SortOrder {
			t.Errorf("tiers not sorted: %s (order %d) before %s (order %d)",
				resp.Tiers[i-1].Name, resp.Tiers[i-1].SortOrder,
				resp.Tiers[i].Name, resp.Tiers[i].SortOrder)
		}
	}

	last := resp.Tiers[len(resp.Tiers)-1]
	if last.Name != "Platinum" {
		t.Errorf("expected Platinum as last tier, got %s", last.Name)
	}

	names := make([]string, len(resp.Tiers))
	for i, tr := range resp.Tiers {
		names[i] = tr.Name
	}
	if names[0] != "Free" {
		t.Errorf("expected Free as first tier, got %s", names[0])
	}
}

func TestListTiers_HasAllFiveTiers(t *testing.T) {
	_, _ = et.NewTestDatabase(context.Background(), "tiersdb")
	ctx := context.Background()

	resp, err := ListTiers(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := map[string]bool{
		"Free":     false,
		"Bronze":   false,
		"Silver":   false,
		"Gold":     false,
		"Platinum": false,
	}
	for _, tr := range resp.Tiers {
		if _, ok := expected[tr.Name]; ok {
			expected[tr.Name] = true
		}
	}
	for name, found := range expected {
		if !found {
			t.Errorf("expected tier %s, not found", name)
		}
	}
}

func TestGetTier_ByID(t *testing.T) {
	_, _ = et.NewTestDatabase(context.Background(), "tiersdb")
	ctx := context.Background()

	tiers, err := ListTiers(ctx)
	if err != nil || len(tiers.Tiers) == 0 {
		t.Fatalf("list tiers error: %v", err)
	}

	for _, tr := range tiers.Tiers {
		t.Run(tr.Name, func(t *testing.T) {
			found, err := GetTier(ctx, tr.ID)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if found.ID != tr.ID {
				t.Errorf("expected ID %s, got %s", tr.ID, found.ID)
			}
			if found.Name != tr.Name {
				t.Errorf("expected name %s, got %s", tr.Name, found.Name)
			}
		})
	}
}

func TestGetTier_NotFound(t *testing.T) {
	_, _ = et.NewTestDatabase(context.Background(), "tiersdb")
	ctx := context.Background()

	_, err := GetTier(ctx, "00000000-0000-0000-0000-000000000000")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestListPlans_IncludesFeatures(t *testing.T) {
	_, _ = et.NewTestDatabase(context.Background(), "tiersdb")
	ctx := context.Background()

	resp, err := ListPlans(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Plans) == 0 {
		t.Fatal("expected plans, got empty list")
	}

	foundWithFeatures := false
	for _, plan := range resp.Plans {
		if plan.Name == "" {
			t.Errorf("plan %s has empty name", plan.ID)
		}
		if plan.Interval == "" {
			t.Errorf("plan %s has empty interval", plan.ID)
		}
		if len(plan.Features) > 0 {
			foundWithFeatures = true
		}
	}
	if !foundWithFeatures {
		t.Error("expected at least some plans to have features")
	}
}

func TestListPlans_NoLifetimePlans(t *testing.T) {
	_, _ = et.NewTestDatabase(context.Background(), "tiersdb")
	ctx := context.Background()

	resp, err := ListPlans(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, plan := range resp.Plans {
		if plan.Interval == "lifetime" {
			t.Errorf("plan %s (%s) has lifetime interval which should not exist", plan.Name, plan.ID)
		}
	}
}

func TestListPlans_SortedByPrice(t *testing.T) {
	_, _ = et.NewTestDatabase(context.Background(), "tiersdb")
	ctx := context.Background()

	resp, err := ListPlans(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for i := 1; i < len(resp.Plans); i++ {
		if resp.Plans[i].PriceUsdCents < resp.Plans[i-1].PriceUsdCents {
			t.Errorf("plans not sorted by price: %s (%d) before %s (%d)",
				resp.Plans[i-1].Name, resp.Plans[i-1].PriceUsdCents,
				resp.Plans[i].Name, resp.Plans[i].PriceUsdCents)
		}
	}
}
