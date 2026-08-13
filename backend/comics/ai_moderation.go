package comics

import (
	"context"
	"time"

	"comics-galore/backend/aiprovider"
	myauth "comics-galore/backend/auth"

	"encore.dev/beta/errs"
	"encore.dev/pubsub"
)

// AIModeratorAPIKey is the app-global secret for the OpenAI-compatible endpoint.
var aiSecrets struct {
	AIModeratorAPIKey string
}

// ModerationEvent is published after a comic or comment is created so the AI
// worker can classify it off the request path (ADR 0018).
type ModerationEvent struct {
	TargetType string `json:"target_type"` // comic | comment
	TargetID   string `json:"target_id"`
}

var moderationTopic = pubsub.NewTopic[ModerationEvent]("ai-moderation", pubsub.TopicConfig{
	DeliveryGuarantee: pubsub.AtLeastOnce,
})

var _ = pubsub.NewSubscription(moderationTopic, "ai-moderate", pubsub.SubscriptionConfig[ModerationEvent]{
	Handler: handleModerationEvent,
})

func handleModerationEvent(ctx context.Context, ev ModerationEvent) error {
	if ev.TargetType != "comic" && ev.TargetType != "comment" {
		return nil
	}
	return runAIModeration(ctx, ev.TargetType, ev.TargetID)
}
// runAIModeration classifies a comic or comment and applies thresholds:
// auto-approve/reject, or queue for human review when uncertain. Idempotent-ish
// (at-least-once delivery may re-run; a repeated decision simply overwrites).
func runAIModeration(ctx context.Context, targetType, targetID string) error {
	cfg, err := myauth.GetAIModerationConfig(ctx)
	if err != nil || !cfg.Enabled {
		return nil // disabled or unavailable → human path
	}
	if aiSecrets.AIModeratorAPIKey == "" {
		return nil
	}

	content, err := moderationContent(ctx, targetType, targetID)
	if err != nil {
		return err
	}

	client := aiprovider.New(cfg.Endpoint, aiSecrets.AIModeratorAPIKey, cfg.Model)
	result, err := client.Classify(ctx, aiprovider.ClassifyRequest{
		SystemPrompt: cfg.Prompt,
		Content:      content,
	})
	if err != nil {
		return err
	}

	recordDecision(ctx, targetType, targetID, result, cfg.Model)

	switch {
	case result.Decision == "approved" && result.Confidence >= cfg.AutoApproveThreshold:
		applyApproval(ctx, targetType, targetID)
	case result.Decision == "rejected" && result.Confidence <= cfg.AutoRejectThreshold:
		applyRejection(ctx, targetType, targetID, result.Reason)
	default:
		queueForReview(ctx, targetType, targetID)
	}
	return nil
}

func moderationContent(ctx context.Context, targetType, targetID string) (string, error) {
	var content string
	if targetType == "comic" {
		err := db.QueryRow(ctx, `SELECT COALESCE(title,'') || ' ' || COALESCE(description,'') FROM comics WHERE id = $1`, targetID).Scan(&content)
		return content, err
	}
	err := db.QueryRow(ctx, `SELECT body_text FROM comments WHERE id = $1`, targetID).Scan(&content)
	return content, err
}

func recordDecision(ctx context.Context, targetType, targetID string, r *aiprovider.ClassifyResult, model string) {
	db.Exec(ctx, `
		INSERT INTO ai_decisions (target_type, target_id, decision, confidence, reason, model)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, targetType, targetID, r.Decision, r.Confidence, r.Reason, model)
}

func applyApproval(ctx context.Context, targetType, targetID string) {
	if targetType == "comic" {
		db.Exec(ctx, `UPDATE comics SET status = 'published', published_at = now(), updated_at = now() WHERE id = $1 AND status = 'pending_review'`, targetID)
	}
	// comments are visible by default; nothing to do for approval.
}

func applyRejection(ctx context.Context, targetType, targetID string, reason string) {
	if targetType == "comic" {
		db.Exec(ctx, `UPDATE comics SET status = 'rejected', rejection_reason = $2, updated_at = now() WHERE id = $1 AND status = 'pending_review'`, targetID, reason)
	} else {
		db.Exec(ctx, `DELETE FROM comments WHERE id = $1`, targetID)
	}
}

func queueForReview(ctx context.Context, targetType, targetID string) {
	db.Exec(ctx, `
		INSERT INTO ai_review_queue (target_type, target_id)
		VALUES ($1, $2)
		ON CONFLICT DO NOTHING
	`, targetType, targetID)
}

// ----- Admin endpoints -----

type AIReviewItem struct {
	ID         string    `json:"id"`
	TargetType string    `json:"target_type"`
	TargetID   string    `json:"target_id"`
	Preview    string    `json:"preview"`
	CreatedAt  time.Time `json:"created_at"`
}

type AIReviewQueueResponse struct {
	Items []AIReviewItem `json:"items"`
}

//encore:api auth method=GET path=/admin/ai/queue
func AIReviewQueue(ctx context.Context) (*AIReviewQueueResponse, error) {
	ad, hasAuth := getAuthData(ctx)
	if !hasAuth || (ad.Role != "moderator" && ad.Role != "admin") {
		return nil, &errs.Error{Code: errs.PermissionDenied, Message: "requires moderator or admin"}
	}

	rows, err := db.Query(ctx, `
		SELECT id, target_type, target_id, created_at
		FROM ai_review_queue WHERE status = 'pending' ORDER BY created_at ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []AIReviewItem
	for rows.Next() {
		var it AIReviewItem
		if err := rows.Scan(&it.ID, &it.TargetType, &it.TargetID, &it.CreatedAt); err != nil {
			return nil, err
		}
		it.Preview = previewFor(ctx, it.TargetType, it.TargetID)
		items = append(items, it)
	}
	if items == nil {
		items = []AIReviewItem{}
	}
	return &AIReviewQueueResponse{Items: items}, rows.Err()
}

type ResolveAIReviewParams struct {
	Action string `json:"action"` // approve | reject
}

//encore:api auth method=POST path=/admin/ai/queue/:id/resolve
func ResolveAIReview(ctx context.Context, id string, p *ResolveAIReviewParams) error {
	ad, hasAuth := getAuthData(ctx)
	if !hasAuth || (ad.Role != "moderator" && ad.Role != "admin") {
		return &errs.Error{Code: errs.PermissionDenied, Message: "requires moderator or admin"}
	}
	if p.Action != "approve" && p.Action != "reject" {
		return &errs.Error{Code: errs.InvalidArgument, Message: "action must be approve or reject"}
	}

	var targetType, targetID string
	err := db.QueryRow(ctx, `SELECT target_type, target_id FROM ai_review_queue WHERE id = $1 AND status = 'pending'`, id).Scan(&targetType, &targetID)
	if err != nil {
		if isNoRows(err) {
			return &errs.Error{Code: errs.NotFound, Message: "queue item not found"}
		}
		return err
	}

	if p.Action == "approve" {
		applyApproval(ctx, targetType, targetID)
	} else {
		applyRejection(ctx, targetType, targetID, "human review")
	}
	db.Exec(ctx, `UPDATE ai_review_queue SET status = 'resolved', resolved_by = $1, resolved_at = now() WHERE id = $2`, ad.UserID, id)
	return nil
}

type AIDecision struct {
	ID         string    `json:"id"`
	TargetType string    `json:"target_type"`
	TargetID   string    `json:"target_id"`
	Decision   string    `json:"decision"`
	Confidence float64   `json:"confidence"`
	Reason     string    `json:"reason"`
	Model      string    `json:"model"`
	CreatedAt  time.Time `json:"created_at"`
}

type AIDecisionsResponse struct {
	Decisions []AIDecision `json:"decisions"`
}

//encore:api auth method=GET path=/admin/ai/decisions
func AIDecisions(ctx context.Context) (*AIDecisionsResponse, error) {
	ad, hasAuth := getAuthData(ctx)
	if !hasAuth || ad.Role != "admin" {
		return nil, &errs.Error{Code: errs.PermissionDenied, Message: "admin only"}
	}

	rows, err := db.Query(ctx, `
		SELECT id, target_type, target_id, decision, COALESCE(confidence, 0), COALESCE(reason, ''), COALESCE(model, ''), created_at
		FROM ai_decisions ORDER BY created_at DESC LIMIT 100
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var decisions []AIDecision
	for rows.Next() {
		var d AIDecision
		if err := rows.Scan(&d.ID, &d.TargetType, &d.TargetID, &d.Decision, &d.Confidence, &d.Reason, &d.Model, &d.CreatedAt); err != nil {
			return nil, err
		}
		decisions = append(decisions, d)
	}
	if decisions == nil {
		decisions = []AIDecision{}
	}
	return &AIDecisionsResponse{Decisions: decisions}, rows.Err()
}

func previewFor(ctx context.Context, targetType, targetID string) string {
	var s string
	if targetType == "comic" {
		db.QueryRow(ctx, `SELECT title FROM comics WHERE id = $1`, targetID).Scan(&s)
	} else {
		db.QueryRow(ctx, `SELECT body_text FROM comments WHERE id = $1`, targetID).Scan(&s)
	}
	if len(s) > 120 {
		s = s[:120] + "…"
	}
	return s
}
