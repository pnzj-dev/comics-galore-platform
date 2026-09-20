package mcp

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"strings"

	"comics-galore/backend/auth"
	"comics-galore/backend/comics"
	"comics-galore/backend/social"

	"encore.dev"
	"encore.dev/beta/errs"
	"encore.dev/storage/sqldb"
)

var db = sqldb.NewDatabase("mcpdb", sqldb.DatabaseConfig{
	Migrations: "./migrations",
})

// Cross-service seams, overridable in tests. Encore API functions can't be
// referenced directly, so each is wrapped in a closure.
var (
	getUserRole = func(ctx context.Context, p *auth.GetUserRoleParams) (*auth.GetUserRoleResponse, error) {
		return auth.GetUserRole(ctx, p)
	}
	approveComic = func(ctx context.Context, p *comics.InternalModerateParams) error {
		return comics.InternalApproveComic(ctx, p)
	}
	rejectComic = func(ctx context.Context, p *comics.InternalModerateParams) error {
		return comics.InternalRejectComic(ctx, p)
	}
	resolveFlag = func(ctx context.Context, p *comics.InternalResolveFlagParams) error {
		return comics.InternalResolveCommentFlag(ctx, p)
	}
	createComment = func(ctx context.Context, p *comics.InternalCreateCommentParams) (*comics.CommentData, error) {
		return comics.InternalCreateComment(ctx, p)
	}
	listTickets = func(ctx context.Context, p *social.InternalListTicketsParams) (*social.ListTicketsResponse, error) {
		return social.InternalListTickets(ctx, p)
	}
	replyTicket = func(ctx context.Context, p *social.InternalReplyTicketParams) (*social.SupportMessage, error) {
		return social.InternalReplyTicket(ctx, p)
	}
	resolveTicket = func(ctx context.Context, p *social.InternalResolveTicketParams) error {
		return social.InternalResolveTicket(ctx, p)
	}
)

// resolveMcpKey maps a server-issued API key to the bound user and role.
func resolveMcpKey(ctx context.Context, token string) (userID, role string, err error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return "", "", &errs.Error{Code: errs.Unauthenticated, Message: "missing mcp key"}
	}
	sum := sha256.Sum256([]byte(token))
	hash := hex.EncodeToString(sum[:])
	if err := db.QueryRow(ctx, `SELECT user_id FROM mcp_keys WHERE key_hash = $1 AND revoked_at IS NULL`, hash).Scan(&userID); err != nil {
		return "", "", &errs.Error{Code: errs.PermissionDenied, Message: "invalid mcp key"}
	}
	resp, err := getUserRole(ctx, &auth.GetUserRoleParams{UserID: userID})
	if err != nil {
		return "", "", err
	}
	return userID, resp.Role, nil
}

func bearerToken() string {
	return strings.TrimSpace(strings.TrimPrefix(encore.CurrentRequest().Headers.Get("Authorization"), "Bearer "))
}

func requireModerator(role string) error {
	if role != "admin" && role != "moderator" {
		return &errs.Error{Code: errs.PermissionDenied, Message: "requires moderator or admin"}
	}
	return nil
}

// ----- Tools -----

type ModerateComicParams struct {
	Action  string `json:"action"`
	ComicID string `json:"comic_id"`
	Reason  string `json:"reason"`
}

//encore:api public method=POST path=/mcp/moderate-comic
func ModerateComic(ctx context.Context, p *ModerateComicParams) error {
	return doModerateComic(ctx, bearerToken(), p)
}

func doModerateComic(ctx context.Context, token string, p *ModerateComicParams) error {
	actor, role, err := resolveMcpKey(ctx, token)
	if err != nil {
		return err
	}
	if err := requireModerator(role); err != nil {
		return err
	}
	switch p.Action {
	case "approve":
		return approveComic(ctx, &comics.InternalModerateParams{ActorID: actor, ComicID: p.ComicID})
	case "reject":
		return rejectComic(ctx, &comics.InternalModerateParams{ActorID: actor, ComicID: p.ComicID, Reason: p.Reason})
	default:
		return &errs.Error{Code: errs.InvalidArgument, Message: "action must be 'approve' or 'reject'"}
	}
}

type ResolveFlagParams struct {
	FlagID string `json:"flag_id"`
}

//encore:api public method=POST path=/mcp/resolve-comment-flag
func ResolveCommentFlag(ctx context.Context, p *ResolveFlagParams) error {
	return doResolveCommentFlag(ctx, bearerToken(), p)
}

func doResolveCommentFlag(ctx context.Context, token string, p *ResolveFlagParams) error {
	actor, role, err := resolveMcpKey(ctx, token)
	if err != nil {
		return err
	}
	if err := requireModerator(role); err != nil {
		return err
	}
	return resolveFlag(ctx, &comics.InternalResolveFlagParams{ActorID: actor, FlagID: p.FlagID})
}

type WriteCommentParams struct {
	ComicID string `json:"comic_id"`
	Body    string `json:"body"`
}

//encore:api public method=POST path=/mcp/write-comment
func WriteComment(ctx context.Context, p *WriteCommentParams) (*comics.CommentData, error) {
	return doWriteComment(ctx, bearerToken(), p)
}

func doWriteComment(ctx context.Context, token string, p *WriteCommentParams) (*comics.CommentData, error) {
	actor, _, err := resolveMcpKey(ctx, token)
	if err != nil {
		return nil, err
	}
	return createComment(ctx, &comics.InternalCreateCommentParams{ActorID: actor, ComicID: p.ComicID, Body: p.Body})
}

type ListTicketsParams struct {
	Status string `json:"status"`
}

//encore:api public method=POST path=/mcp/list-support-tickets
func ListSupportTickets(ctx context.Context, p *ListTicketsParams) (*social.ListTicketsResponse, error) {
	return doListSupportTickets(ctx, bearerToken(), p)
}

func doListSupportTickets(ctx context.Context, token string, p *ListTicketsParams) (*social.ListTicketsResponse, error) {
	actor, role, err := resolveMcpKey(ctx, token)
	if err != nil {
		return nil, err
	}
	if err := requireModerator(role); err != nil {
		return nil, err
	}
	return listTickets(ctx, &social.InternalListTicketsParams{ActorID: actor, Status: p.Status})
}

type ReplyTicketParams struct {
	TicketID string `json:"ticket_id"`
	Body     string `json:"body"`
}

//encore:api public method=POST path=/mcp/reply-support-ticket
func ReplySupportTicket(ctx context.Context, p *ReplyTicketParams) (*social.SupportMessage, error) {
	return doReplySupportTicket(ctx, bearerToken(), p)
}

func doReplySupportTicket(ctx context.Context, token string, p *ReplyTicketParams) (*social.SupportMessage, error) {
	actor, role, err := resolveMcpKey(ctx, token)
	if err != nil {
		return nil, err
	}
	if err := requireModerator(role); err != nil {
		return nil, err
	}
	return replyTicket(ctx, &social.InternalReplyTicketParams{ActorID: actor, TicketID: p.TicketID, Body: p.Body})
}

type ResolveTicketParams struct {
	TicketID string `json:"ticket_id"`
}

//encore:api public method=POST path=/mcp/resolve-support-ticket
func ResolveSupportTicket(ctx context.Context, p *ResolveTicketParams) error {
	return doResolveSupportTicket(ctx, bearerToken(), p)
}

func doResolveSupportTicket(ctx context.Context, token string, p *ResolveTicketParams) error {
	actor, role, err := resolveMcpKey(ctx, token)
	if err != nil {
		return err
	}
	if err := requireModerator(role); err != nil {
		return err
	}
	return resolveTicket(ctx, &social.InternalResolveTicketParams{ActorID: actor, TicketID: p.TicketID})
}
