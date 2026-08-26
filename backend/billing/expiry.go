package billing

import (
	"context"

	myauth "comics-galore/backend/auth"
	myjobs "comics-galore/backend/jobs"

	"encore.dev/beta/auth"
	"encore.dev/beta/errs"
	"encore.dev/cron"
)

var _ = cron.NewJob("waiting-pay-expiry", cron.JobConfig{
	Title:    "Expire WAITING_PAY subscriptions",
	Every:    1 * cron.Hour,
	Endpoint: ExpireWaitingPaySubscriptions,
})

// ExpireWaitingPaySubscriptions downgrades users whose subscription has been
// stuck in the "waiting for payment" state longer than the configured threshold.
// The threshold and enabled flag are runtime-configurable in admin settings
// (no restart required); the job itself runs hourly and is idempotent.
//encore:api private
func ExpireWaitingPaySubscriptions(ctx context.Context) error {
	return runTrackedJob(ctx, "waiting-pay-expiry", "", func() error {
		return expireWaitingPaySubscriptions(ctx)
	})
}

// RunWaitingPayExpiry lets an admin manually trigger the WAITING_PAY sweep.
//encore:api auth method=POST path=/admin/jobs/waiting-pay-expiry/run
func RunWaitingPayExpiry(ctx context.Context) error {
	ad := auth.Data().(*myauth.AuthData)
	if ad.Role != "admin" {
		return &errs.Error{Code: errs.PermissionDenied, Message: "admin only"}
	}
	return ExpireWaitingPaySubscriptions(ctx)
}

func expireWaitingPaySubscriptions(ctx context.Context) error {
	cfg, err := myauth.GetBillingConfig(ctx)
	if err != nil {
		return err
	}
	if !cfg.WaitingPayJobEnabled {
		return nil
	}
	hours := cfg.WaitingPayExpiryHours
	if hours <= 0 {
		hours = 24
	}

	rows, err := db.Query(ctx, `
		SELECT id, user_id FROM subscriptions
		WHERE status = 'waiting_pay' AND updated_at < now() - make_interval(hours => $1)
	`, hours)
	if err != nil {
		return err
	}
	defer rows.Close()

	type sub struct {
		ID     string
		UserID string
	}
	var subs []sub
	for rows.Next() {
		var s sub
		if err := rows.Scan(&s.ID, &s.UserID); err != nil {
			return err
		}
		subs = append(subs, s)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	for _, s := range subs {
		if err := myauth.SetUserTier(ctx, &myauth.SetUserTierParams{UserID: s.UserID, Tier: "free"}); err != nil {
			return err
		}
		db.Exec(ctx, `UPDATE subscriptions SET status = 'expired', active = false, updated_at = now() WHERE id = $1`, s.ID)
	}
	return nil
}

// runTrackedJob records a job run in the jobs service around fn (best-effort).
func runTrackedJob(ctx context.Context, name, ref string, fn func() error) error {
	run, _ := myjobs.RecordJobStart(ctx, &myjobs.RecordJobStartParams{JobName: name, Ref: ref})
	err := fn()
	if run != nil {
		errStr := ""
		if err != nil {
			errStr = err.Error()
		}
		_ = myjobs.RecordJobFinish(ctx, &myjobs.RecordJobFinishParams{ID: run.ID, Error: errStr})
	}
	return err
}
