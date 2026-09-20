package social

import (
	"context"

	myauth "comics-galore/backend/auth"

	"encore.dev/beta/errs"
)

// Private endpoints for the key-gated MCP service. The mcp service verifies
// the actor's key and role before calling; actorID is used for attribution.

type InternalListTicketsParams struct {
	ActorID string `json:"actor_id"`
	Status  string `json:"status"`
}

//encore:api private method=POST path=/internal/social/list-tickets
func InternalListTickets(ctx context.Context, p *InternalListTicketsParams) (*ListTicketsResponse, error) {
	if p.Status != "" {
		return listTickets(ctx, `SELECT id, user_id, subject, status, priority, COALESCE(assigned_to::text, ''), created_at, resolved_at FROM support_tickets WHERE status = $1 ORDER BY created_at DESC`, p.Status)
	}
	return listTickets(ctx, `SELECT id, user_id, subject, status, priority, COALESCE(assigned_to::text, ''), created_at, resolved_at FROM support_tickets ORDER BY created_at DESC`)
}

type InternalReplyTicketParams struct {
	ActorID  string `json:"actor_id"`
	TicketID string `json:"ticket_id"`
	Body     string `json:"body"`
}

//encore:api private method=POST path=/internal/social/reply-ticket
func InternalReplyTicket(ctx context.Context, p *InternalReplyTicketParams) (*SupportMessage, error) {
	if p.Body == "" {
		return nil, &errs.Error{Code: errs.InvalidArgument, Message: "body is required"}
	}
	var ownerID string
	if err := db.QueryRow(ctx, `SELECT user_id FROM support_tickets WHERE id = $1`, p.TicketID).Scan(&ownerID); err != nil {
		if isNoRows(err) {
			return nil, &errs.Error{Code: errs.NotFound, Message: "ticket not found"}
		}
		return nil, err
	}
	var m SupportMessage
	err := db.QueryRow(ctx, `
		INSERT INTO support_messages (ticket_id, sender_id, is_staff, body)
		VALUES ($1, $2, true, $3)
		RETURNING id, ticket_id, sender_id, is_staff, body, created_at
	`, p.TicketID, p.ActorID, p.Body).Scan(&m.ID, &m.TicketID, &m.SenderID, &m.IsStaff, &m.Body, &m.CreatedAt)
	if err != nil {
		return nil, err
	}
	if ownerID != p.ActorID {
		var subject string
		db.QueryRow(ctx, `SELECT subject FROM support_tickets WHERE id = $1`, p.TicketID).Scan(&subject)
		myauth.NotifySupportReply(ctx, &myauth.NotifySupportReplyParams{UserID: ownerID, Subject: subject})
	}
	return &m, nil
}

type InternalResolveTicketParams struct {
	ActorID  string `json:"actor_id"`
	TicketID string `json:"ticket_id"`
}

//encore:api private method=POST path=/internal/social/resolve-ticket
func InternalResolveTicket(ctx context.Context, p *InternalResolveTicketParams) error {
	_, err := db.Exec(ctx, `UPDATE support_tickets SET status = 'resolved', resolved_at = now() WHERE id = $1`, p.TicketID)
	return err
}
