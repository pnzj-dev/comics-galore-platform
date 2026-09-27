package auth

import (
	"context"

	"encore.dev/beta/errs"
)

// These private endpoints are the privileged surface for the key-gated MCP
// service. The mcp service verifies the actor's key and admin role before
// calling; here the explicit ActorID is used for audit-trail attribution.

type InternalUserActionParams struct {
	ActorID string `json:"actor_id"`
	UserID  string `json:"user_id"`
	Reason  string `json:"reason"`
}

func requireActorAndTarget(p *InternalUserActionParams) error {
	if p.ActorID == "" || p.UserID == "" {
		return &errs.Error{Code: errs.InvalidArgument, Message: "actor_id and user_id are required"}
	}
	if p.ActorID == p.UserID {
		return &errs.Error{Code: errs.InvalidArgument, Message: "cannot act on yourself"}
	}
	return nil
}

//encore:api private method=POST path=/internal/users/ban
func InternalBanUser(ctx context.Context, p *InternalUserActionParams) error {
	if err := requireActorAndTarget(p); err != nil {
		return err
	}
	res, err := db.Exec(ctx, `UPDATE users SET banned_at = now() WHERE id = $1`, p.UserID)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return &errs.Error{Code: errs.NotFound, Message: "user not found"}
	}
	db.Exec(ctx, `INSERT INTO audit_logs (actor_id, action, target_type, target_id, details) VALUES ($1, 'ban_user', 'user', $2, $3)`,
		p.ActorID, p.UserID, `{"reason":"`+p.Reason+`"}`)
	return nil
}

//encore:api private method=POST path=/internal/users/unban
func InternalUnbanUser(ctx context.Context, p *InternalUserActionParams) error {
	if err := requireActorAndTarget(p); err != nil {
		return err
	}
	res, err := db.Exec(ctx, `UPDATE users SET banned_at = NULL WHERE id = $1`, p.UserID)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return &errs.Error{Code: errs.NotFound, Message: "user not found"}
	}
	db.Exec(ctx, `INSERT INTO audit_logs (actor_id, action, target_type, target_id) VALUES ($1, 'unban_user', 'user', $2)`, p.ActorID, p.UserID)
	return nil
}

//encore:api private method=POST path=/internal/users/suspend
func InternalSuspendUser(ctx context.Context, p *InternalUserActionParams) error {
	if err := requireActorAndTarget(p); err != nil {
		return err
	}
	res, err := db.Exec(ctx, `UPDATE users SET suspended_at = now() WHERE id = $1`, p.UserID)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return &errs.Error{Code: errs.NotFound, Message: "user not found"}
	}
	db.Exec(ctx, `INSERT INTO audit_logs (actor_id, action, target_type, target_id, details) VALUES ($1, 'suspend_user', 'user', $2, $3)`,
		p.ActorID, p.UserID, `{"reason":"`+p.Reason+`"}`)
	return nil
}
