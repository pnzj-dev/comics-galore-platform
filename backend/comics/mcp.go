package comics

import (
	"context"

	"encore.dev/beta/errs"
)

// These private endpoints are the privileged surface for the key-gated MCP
// service. The mcp service verifies the actor's key and role before calling;
// here the explicit actorID is used only for audit-trail attribution.

type InternalModerateParams struct {
	ActorID string `json:"actor_id"`
	ComicID string `json:"comic_id"`
	Reason  string `json:"reason"`
}

//encore:api private method=POST path=/internal/comics/approve
func InternalApproveComic(ctx context.Context, p *InternalModerateParams) error {
	var status string
	if err := db.QueryRow(ctx, `SELECT status FROM comics WHERE id = $1`, p.ComicID).Scan(&status); err != nil {
		return &errs.Error{Code: errs.NotFound, Message: "comic not found"}
	}
	if status != "pending_review" {
		return &errs.Error{Code: errs.InvalidArgument, Message: "comic is not pending review"}
	}
	if _, err := db.Exec(ctx, `UPDATE comics SET status = 'published', published_at = now(), updated_at = now() WHERE id = $1`, p.ComicID); err != nil {
		return err
	}
	db.Exec(ctx, `INSERT INTO audit_logs (actor_id, action, target_type, target_id) VALUES ($1, 'approve_comic', 'comic', $2)`, p.ActorID, p.ComicID)
	go notifyUploaderFollowers(context.Background(), p.ComicID)
	return nil
}

//encore:api private method=POST path=/internal/comics/reject
func InternalRejectComic(ctx context.Context, p *InternalModerateParams) error {
	_, err := db.Exec(ctx, `UPDATE comics SET status = 'rejected', rejection_reason = $1, updated_at = now() WHERE id = $2 AND status = 'pending_review'`, p.Reason, p.ComicID)
	if err != nil {
		return err
	}
	db.Exec(ctx, `INSERT INTO audit_logs (actor_id, action, target_type, target_id, details) VALUES ($1, 'reject_comic', 'comic', $2, $3)`, p.ActorID, p.ComicID, `{"reason":"`+p.Reason+`"}`)
	return nil
}

type InternalResolveFlagParams struct {
	ActorID string `json:"actor_id"`
	FlagID  string `json:"flag_id"`
}

//encore:api private method=POST path=/internal/comics/resolve-flag
func InternalResolveCommentFlag(ctx context.Context, p *InternalResolveFlagParams) error {
	_, err := db.Exec(ctx, `UPDATE comment_flags SET status = 'resolved', resolved_at = now(), resolved_by = $1 WHERE id = $2 AND status = 'open'`, p.ActorID, p.FlagID)
	return err
}

type InternalCreateCommentParams struct {
	ActorID string `json:"actor_id"`
	ComicID string `json:"comic_id"`
	Body    string `json:"body"`
}

//encore:api private method=POST path=/internal/comics/create-comment
func InternalCreateComment(ctx context.Context, p *InternalCreateCommentParams) (*CommentData, error) {
	if p.Body == "" {
		return nil, &errs.Error{Code: errs.InvalidArgument, Message: "body is required"}
	}
	var c CommentData
	err := db.QueryRow(ctx, `
		INSERT INTO comments (comic_id, user_id, body_text)
		VALUES ($1, $2, $3)
		RETURNING id, comic_id, user_id, COALESCE(parent_id::text, ''), body_text, created_at
	`, p.ComicID, p.ActorID, p.Body).Scan(&c.ID, &c.ComicID, &c.UserID, &c.ParentID, &c.BodyText, &c.CreatedAt)
	if err != nil {
		return nil, err
	}
	publishComment(p.ComicID, c)
	moderationTopic.Publish(ctx, ModerationEvent{TargetType: "comment", TargetID: c.ID})
	return &c, nil
}
