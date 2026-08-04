package comics

import (
	"context"
	"fmt"

	"encore.dev/beta/errs"
)

type SeedEngagementResponse struct {
	Comments  int `json:"comments"`
	Reactions int `json:"reactions"`
	Skipped   int `json:"skipped"`
}

type demoComment struct {
	ComicID string
	UserID  string
	Text    string
}

//encore:api public method=POST path=/dev/seed-engagement
func DevSeedEngagement(ctx context.Context, p *SeedParams) (*SeedEngagementResponse, error) {
	if !isDevTokenValid(p.Token) {
		return nil, &errs.Error{Code: errs.PermissionDenied, Message: "invalid dev seed token"}
	}

	var totalComments int
	db.QueryRow(ctx, `SELECT COUNT(*) FROM audit_logs WHERE action = 'seed_comment'`).Scan(&totalComments)
	if totalComments > 0 {
		return &SeedEngagementResponse{Comments: 0, Reactions: 0, Skipped: totalComments}, nil
	}

	comments := []demoComment{
		{"20000000-0000-0000-0000-000000000001", "10000000-0000-0000-0000-000000000004", "Space Cat is my spirit animal! The art style is incredible."},
		{"20000000-0000-0000-0000-000000000001", "10000000-0000-0000-0000-000000000005", "Captain Whiskers deserves her own animated series."},
		{"20000000-0000-0000-0000-000000000001", "10000000-0000-0000-0000-000000000007", "The nebula scenes in issue #1 are breathtaking."},
		{"20000000-0000-0000-0000-000000000002", "10000000-0000-0000-0000-000000000004", "Neo-Tokyo feels so alive. The rain effects are gorgeous."},
		{"20000000-0000-0000-0000-000000000002", "10000000-0000-0000-0000-000000000005", "Kira is such a compelling protagonist. Love the cyberpunk noir vibe."},
		{"20000000-0000-0000-0000-000000000004", "10000000-0000-0000-0000-000000000004", "Ada's journey is so touching. The steampunk aesthetics are perfect."},
		{"20000000-0000-0000-0000-000000000004", "10000000-0000-0000-0000-000000000007", "The clockwork mechanisms are drawn with incredible detail."},
		{"20000000-0000-0000-0000-000000000008", "10000000-0000-0000-0000-000000000005", "The fight choreography in this comic is insane."},
		{"20000000-0000-0000-0000-000000000008", "10000000-0000-0000-0000-000000000004", "I never thought I'd root for a rabbit samurai, but here I am."},
		{"20000000-0000-0000-0000-000000000006", "10000000-0000-0000-0000-000000000005", "The Goblin King's design is absolutely terrifying."},
		{"20000000-0000-0000-0000-000000000005", "10000000-0000-0000-0000-000000000007", "This comic makes my brain hurt in the best possible way."},
		{"20000000-0000-0000-0000-000000000010", "10000000-0000-0000-0000-000000000004", "Scrapheap best mecha. The underdog story is so well done."},
		{"20000000-0000-0000-0000-000000000011", "10000000-0000-0000-0000-000000000005", "This is the coziest comic I've ever read."},
		{"20000000-0000-0000-0000-000000000014", "10000000-0000-0000-0000-000000000007", "Lovecraftian horror at its finest. The sense of scale is terrifying."},
		{"20000000-0000-0000-0000-000000000012", "10000000-0000-0000-0000-000000000004", "The Norse mythology integration is seamless."},
		{"20000000-0000-0000-0000-000000000019", "10000000-0000-0000-0000-000000000005", "I did NOT expect to cry reading about a 12-year-old Grim Reaper."},
		{"20000000-0000-0000-0000-000000000002", "10000000-0000-0000-0000-000000000007", "That plot twist at the end of Chapter 3! I did NOT see that coming."},
	}

	commentCount := 0
	for _, c := range comments {
		var exists bool
		db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM audit_logs WHERE action = 'seed_comment' AND target_id = $1 AND actor_id = $2)`,
			c.ComicID, c.UserID).Scan(&exists)
		if exists {
			continue
		}

		var commentID string
		db.QueryRow(ctx, `SELECT gen_random_uuid()::text`).Scan(&commentID)

		db.Exec(ctx, `
			INSERT INTO audit_logs (actor_id, action, target_type, target_id, details)
			VALUES ($1, 'seed_comment', 'comic', $2, $3)
		`, c.UserID, c.ComicID, fmt.Sprintf(`{"comment":"%s"}`, c.Text))
		commentCount++
	}

	// Add reactions (likes/favorites)
	type demoReaction struct {
		UserID   string
		ComicID  string
		Type     string // "like" or "favorite"
	}

	reactions := []demoReaction{
		{"10000000-0000-0000-0000-000000000004", "20000000-0000-0000-0000-000000000001", "like"},
		{"10000000-0000-0000-0000-000000000005", "20000000-0000-0000-0000-000000000001", "like"},
		{"10000000-0000-0000-0000-000000000001", "20000000-0000-0000-0000-000000000001", "favorite"},
		{"10000000-0000-0000-0000-000000000004", "20000000-0000-0000-0000-000000000002", "like"},
		{"10000000-0000-0000-0000-000000000007", "20000000-0000-0000-0000-000000000002", "like"},
		{"10000000-0000-0000-0000-000000000005", "20000000-0000-0000-0000-000000000002", "favorite"},
		{"10000000-0000-0000-0000-000000000004", "20000000-0000-0000-0000-000000000004", "like"},
		{"10000000-0000-0000-0000-000000000005", "20000000-0000-0000-0000-000000000004", "like"},
		{"10000000-0000-0000-0000-000000000007", "20000000-0000-0000-0000-000000000004", "like"},
		{"10000000-0000-0000-0000-000000000001", "20000000-0000-0000-0000-000000000004", "favorite"},
		{"10000000-0000-0000-0000-000000000004", "20000000-0000-0000-0000-000000000008", "like"},
		{"10000000-0000-0000-0000-000000000007", "20000000-0000-0000-0000-000000000008", "favorite"},
		{"10000000-0000-0000-0000-000000000005", "20000000-0000-0000-0000-000000000005", "like"},
		{"10000000-0000-0000-0000-000000000004", "20000000-0000-0000-0000-000000000010", "like"},
		{"10000000-0000-0000-0000-000000000007", "20000000-0000-0000-0000-000000000010", "like"},
		{"10000000-0000-0000-0000-000000000004", "20000000-0000-0000-0000-000000000011", "like"},
		{"10000000-0000-0000-0000-000000000005", "20000000-0000-0000-0000-000000000011", "like"},
		{"10000000-0000-0000-0000-000000000007", "20000000-0000-0000-0000-000000000011", "like"},
		{"10000000-0000-0000-0000-000000000004", "20000000-0000-0000-0000-000000000014", "like"},
		{"10000000-0000-0000-0000-000000000005", "20000000-0000-0000-0000-000000000014", "like"},
		{"10000000-0000-0000-0000-000000000007", "20000000-0000-0000-0000-000000000014", "like"},
		{"10000000-0000-0000-0000-000000000004", "20000000-0000-0000-0000-000000000019", "like"},
		{"10000000-0000-0000-0000-000000000005", "20000000-0000-0000-0000-000000000019", "like"},
		{"10000000-0000-0000-0000-000000000007", "20000000-0000-0000-0000-000000000019", "favorite"},
		{"10000000-0000-0000-0000-000000000005", "20000000-0000-0000-0000-000000000006", "like"},
		{"10000000-0000-0000-0000-000000000001", "20000000-0000-0000-0000-000000000006", "favorite"},
		{"10000000-0000-0000-0000-000000000004", "20000000-0000-0000-0000-000000000012", "like"},
		{"10000000-0000-0000-0000-000000000007", "20000000-0000-0000-0000-000000000012", "like"},
	}

	reactionCount := 0
	for _, r := range reactions {
		var exists bool
		if r.Type == "like" {
			db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM likes WHERE user_id = $1 AND comic_id = $2)`, r.UserID, r.ComicID).Scan(&exists)
			if !exists {
				db.Exec(ctx, `INSERT INTO likes (user_id, comic_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`, r.UserID, r.ComicID)
				db.Exec(ctx, `UPDATE comics SET like_count = like_count + 1 WHERE id = $1`, r.ComicID)
				reactionCount++
			}
		} else {
			db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM favorites WHERE user_id = $1 AND comic_id = $2)`, r.UserID, r.ComicID).Scan(&exists)
			if !exists {
				db.Exec(ctx, `INSERT INTO favorites (user_id, comic_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`, r.UserID, r.ComicID)
				db.Exec(ctx, `UPDATE comics SET fav_count = fav_count + 1 WHERE id = $1`, r.ComicID)
				reactionCount++
			}
		}
	}

	return &SeedEngagementResponse{
		Comments:  commentCount,
		Reactions: reactionCount,
		Skipped:   0,
	}, nil
}
