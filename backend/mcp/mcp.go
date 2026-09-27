package mcp

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"strings"

	"comics-galore/backend/auth"
	"comics-galore/backend/comics"
	"comics-galore/backend/social"
	"comics-galore/backend/upload"

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
	createComic = func(ctx context.Context, p *comics.InternalCreateComicParams) (*comics.InternalCreateComicResponse, error) {
		return comics.InternalCreateComic(ctx, p)
	}
	listFlaggedComments = func(ctx context.Context, p *comics.InternalListFlagsParams) (*comics.ListFlaggedCommentsResponse, error) {
		return comics.InternalListFlaggedComments(ctx, p)
	}
	deleteComment = func(ctx context.Context, p *comics.InternalDeleteCommentParams) error {
		return comics.InternalDeleteComment(ctx, p)
	}
	banUser = func(ctx context.Context, p *auth.InternalUserActionParams) error {
		return auth.InternalBanUser(ctx, p)
	}
	unbanUser = func(ctx context.Context, p *auth.InternalUserActionParams) error {
		return auth.InternalUnbanUser(ctx, p)
	}
	suspendUser = func(ctx context.Context, p *auth.InternalUserActionParams) error {
		return auth.InternalSuspendUser(ctx, p)
	}
	presignUpload = func(ctx context.Context, p *upload.InternalPresignUploadParams) (*upload.InternalPresignUploadResponse, error) {
		return upload.InternalPresignUpload(ctx, p)
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

func requireUploader(role string) error {
	if role != "uploader" && role != "admin" {
		return &errs.Error{Code: errs.PermissionDenied, Message: "requires uploader or admin"}
	}
	return nil
}

func requireAdminRole(role string) error {
	if role != "admin" {
		return &errs.Error{Code: errs.PermissionDenied, Message: "requires admin"}
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

// ----- Agentic creation / moderation / user actions -----

type CreateComicParams struct {
	Title            string   `json:"title"`
	Author           string   `json:"author"`
	Description      string   `json:"description"`
	ContentLanguage  string   `json:"content_language"`
	Category         string   `json:"category"`
	Genre            string   `json:"genre"`
	AgeRating        string   `json:"age_rating"`
	IsPremium        bool     `json:"is_premium"`
	Tags             []string `json:"tags"`
	ReadingDirection string   `json:"reading_direction"`
	CoverKey         string   `json:"cover_key"`
	PageKeys         []string `json:"page_keys"`
	FileKey          string   `json:"file_key"`
	MinTier          string   `json:"min_tier"`
	Publish          bool     `json:"publish"`
	SeriesTitle      string   `json:"series_title"`
}

//encore:api public method=POST path=/mcp/create-comic
func CreateComic(ctx context.Context, p *CreateComicParams) (*comics.InternalCreateComicResponse, error) {
	return doCreateComic(ctx, bearerToken(), p)
}

func doCreateComic(ctx context.Context, token string, p *CreateComicParams) (*comics.InternalCreateComicResponse, error) {
	actor, role, err := resolveMcpKey(ctx, token)
	if err != nil {
		return nil, err
	}
	if err := requireUploader(role); err != nil {
		return nil, err
	}
	return createComic(ctx, &comics.InternalCreateComicParams{
		ActorID:          actor,
		Title:            p.Title,
		Author:           p.Author,
		Description:      p.Description,
		ContentLanguage:  p.ContentLanguage,
		Category:         p.Category,
		Genre:            p.Genre,
		AgeRating:        p.AgeRating,
		IsPremium:        p.IsPremium,
		Tags:             p.Tags,
		ReadingDirection: p.ReadingDirection,
		CoverKey:         p.CoverKey,
		PageKeys:         p.PageKeys,
		FileKey:          p.FileKey,
		MinTier:          p.MinTier,
		Publish:          p.Publish,
		SeriesTitle:      p.SeriesTitle,
	})
}

type ListFlaggedCommentsParams struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
}

//encore:api public method=POST path=/mcp/list-flagged-comments
func ListFlaggedComments(ctx context.Context, p *ListFlaggedCommentsParams) (*comics.ListFlaggedCommentsResponse, error) {
	return doListFlaggedComments(ctx, bearerToken(), p)
}

func doListFlaggedComments(ctx context.Context, token string, p *ListFlaggedCommentsParams) (*comics.ListFlaggedCommentsResponse, error) {
	_, role, err := resolveMcpKey(ctx, token)
	if err != nil {
		return nil, err
	}
	if err := requireModerator(role); err != nil {
		return nil, err
	}
	return listFlaggedComments(ctx, &comics.InternalListFlagsParams{Page: p.Page, Limit: p.Limit})
}

type DeleteCommentParams struct {
	CommentID string `json:"comment_id"`
}

//encore:api public method=POST path=/mcp/delete-comment
func DeleteComment(ctx context.Context, p *DeleteCommentParams) error {
	return doDeleteComment(ctx, bearerToken(), p)
}

func doDeleteComment(ctx context.Context, token string, p *DeleteCommentParams) error {
	actor, role, err := resolveMcpKey(ctx, token)
	if err != nil {
		return err
	}
	if err := requireModerator(role); err != nil {
		return err
	}
	return deleteComment(ctx, &comics.InternalDeleteCommentParams{ActorID: actor, CommentID: p.CommentID})
}

type UserActionParams struct {
	UserID string `json:"user_id"`
	Reason string `json:"reason"`
}

//encore:api public method=POST path=/mcp/ban-user
func BanUser(ctx context.Context, p *UserActionParams) error {
	return doBanUser(ctx, bearerToken(), p)
}

func doBanUser(ctx context.Context, token string, p *UserActionParams) error {
	actor, role, err := resolveMcpKey(ctx, token)
	if err != nil {
		return err
	}
	if err := requireAdminRole(role); err != nil {
		return err
	}
	return banUser(ctx, &auth.InternalUserActionParams{ActorID: actor, UserID: p.UserID, Reason: p.Reason})
}

//encore:api public method=POST path=/mcp/unban-user
func UnbanUser(ctx context.Context, p *UserActionParams) error {
	return doUnbanUser(ctx, bearerToken(), p)
}

func doUnbanUser(ctx context.Context, token string, p *UserActionParams) error {
	actor, role, err := resolveMcpKey(ctx, token)
	if err != nil {
		return err
	}
	if err := requireAdminRole(role); err != nil {
		return err
	}
	return unbanUser(ctx, &auth.InternalUserActionParams{ActorID: actor, UserID: p.UserID, Reason: p.Reason})
}

//encore:api public method=POST path=/mcp/suspend-user
func SuspendUser(ctx context.Context, p *UserActionParams) error {
	return doSuspendUser(ctx, bearerToken(), p)
}

func doSuspendUser(ctx context.Context, token string, p *UserActionParams) error {
	actor, role, err := resolveMcpKey(ctx, token)
	if err != nil {
		return err
	}
	if err := requireAdminRole(role); err != nil {
		return err
	}
	return suspendUser(ctx, &auth.InternalUserActionParams{ActorID: actor, UserID: p.UserID, Reason: p.Reason})
}

type PresignUploadParams struct {
	Kind     string `json:"kind"`
	Filename string `json:"filename"`
}

//encore:api public method=POST path=/mcp/presign-upload
func PresignUpload(ctx context.Context, p *PresignUploadParams) (*upload.InternalPresignUploadResponse, error) {
	return doPresignUpload(ctx, bearerToken(), p)
}

func doPresignUpload(ctx context.Context, token string, p *PresignUploadParams) (*upload.InternalPresignUploadResponse, error) {
	actor, role, err := resolveMcpKey(ctx, token)
	if err != nil {
		return nil, err
	}
	if err := requireUploader(role); err != nil {
		return nil, err
	}
	return presignUpload(ctx, &upload.InternalPresignUploadParams{ActorID: actor, Kind: p.Kind, Filename: p.Filename})
}
