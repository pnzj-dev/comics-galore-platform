package social

import (
	"context"
	"time"

	"encore.dev/beta/auth"
	myauth "comics-galore/backend/auth"
	"comics-galore/backend/turnstile"

	"encore.dev/beta/errs"
)

// ----- Support tickets -----

type SupportTicket struct {
	ID         string    `json:"id"`
	UserID     string    `json:"user_id"`
	Subject    string    `json:"subject"`
	Status     string    `json:"status"`
	Priority   string    `json:"priority"`
	AssignedTo string    `json:"assigned_to,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	ResolvedAt *time.Time `json:"resolved_at,omitempty"`
}

type SupportMessage struct {
	ID        string    `json:"id"`
	TicketID  string    `json:"ticket_id"`
	SenderID  string    `json:"sender_id"`
	IsStaff   bool      `json:"is_staff"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateTicketParams struct {
	Subject        string `json:"subject"`
	Body           string `json:"body"`
	Priority       string `json:"priority"`
	TurnstileToken string `json:"turnstile_token"`
}

type ReplyTicketParams struct {
	Body           string `json:"body"`
	TurnstileToken string `json:"turnstile_token"`
}

type TicketResponse struct {
	Ticket   SupportTicket    `json:"ticket"`
	Messages []SupportMessage `json:"messages"`
}

type ListTicketsResponse struct {
	Tickets []SupportTicket `json:"tickets"`
}

//encore:api auth method=POST path=/support/tickets
func CreateTicket(ctx context.Context, p *CreateTicketParams) (*TicketResponse, error) {
	ad := auth.Data().(*myauth.AuthData)

	if err := turnstile.Verify(ctx, &turnstile.VerifyParams{Token: p.TurnstileToken, Action: "support_ticket"}); err != nil {
		return nil, err
	}

	if p.Subject == "" || p.Body == "" {
		return nil, &errs.Error{Code: errs.InvalidArgument, Message: "subject and body are required"}
	}
	priority := p.Priority
	if priority == "" {
		priority = "normal"
	}

	var t SupportTicket
	err := db.QueryRow(ctx, `
		INSERT INTO support_tickets (user_id, subject, priority)
		VALUES ($1, $2, $3)
		RETURNING id, user_id, subject, status, priority, COALESCE(assigned_to::text, ''), created_at
	`, ad.UserID, p.Subject, priority).Scan(&t.ID, &t.UserID, &t.Subject, &t.Status, &t.Priority, &t.AssignedTo, &t.CreatedAt)
	if err != nil {
		return nil, err
	}

	var m SupportMessage
	err = db.QueryRow(ctx, `
		INSERT INTO support_messages (ticket_id, sender_id, is_staff, body)
		VALUES ($1, $2, false, $3)
		RETURNING id, ticket_id, sender_id, is_staff, body, created_at
	`, t.ID, ad.UserID, p.Body).Scan(&m.ID, &m.TicketID, &m.SenderID, &m.IsStaff, &m.Body, &m.CreatedAt)
	if err != nil {
		return nil, err
	}

	return &TicketResponse{Ticket: t, Messages: []SupportMessage{m}}, nil
}

//encore:api auth method=GET path=/support/tickets
func ListMyTickets(ctx context.Context) (*ListTicketsResponse, error) {
	ad := auth.Data().(*myauth.AuthData)
	return listTickets(ctx, `SELECT id, user_id, subject, status, priority, COALESCE(assigned_to::text, ''), created_at, resolved_at FROM support_tickets WHERE user_id = $1 ORDER BY created_at DESC`, ad.UserID)
}

//encore:api auth method=GET path=/support/tickets/:id
func GetTicket(ctx context.Context, id string) (*TicketResponse, error) {
	ad := auth.Data().(*myauth.AuthData)

	var t SupportTicket
	var assignedTo, resolvedAt interface{}
	err := db.QueryRow(ctx, `
		SELECT id, user_id, subject, status, priority, COALESCE(assigned_to::text, ''), created_at, resolved_at
		FROM support_tickets WHERE id = $1
	`, id).Scan(&t.ID, &t.UserID, &t.Subject, &t.Status, &t.Priority, &t.AssignedTo, &t.CreatedAt, &resolvedAt)
	if err != nil {
		if isNoRows(err) {
			return nil, &errs.Error{Code: errs.NotFound, Message: "ticket not found"}
		}
		return nil, err
	}

	// Only the owner, staff, or assigned support can view.
	isStaff := ad.Role == "admin" || ad.Role == "moderator"
	if !isStaff && t.UserID != ad.UserID {
		return nil, &errs.Error{Code: errs.NotFound, Message: "ticket not found"}
	}
	_ = assignedTo

	messages, err := listMessages(ctx, id)
	if err != nil {
		return nil, err
	}
	return &TicketResponse{Ticket: t, Messages: messages}, nil
}

//encore:api auth method=POST path=/support/tickets/:id/reply
func ReplyTicket(ctx context.Context, id string, p *ReplyTicketParams) (*SupportMessage, error) {
	ad := auth.Data().(*myauth.AuthData)

	if err := turnstile.Verify(ctx, &turnstile.VerifyParams{Token: p.TurnstileToken, Action: "support_ticket"}); err != nil {
		return nil, err
	}

	if p.Body == "" {
		return nil, &errs.Error{Code: errs.InvalidArgument, Message: "body is required"}
	}

	var ownerID string
	if err := db.QueryRow(ctx, `SELECT user_id FROM support_tickets WHERE id = $1`, id).Scan(&ownerID); err != nil {
		if isNoRows(err) {
			return nil, &errs.Error{Code: errs.NotFound, Message: "ticket not found"}
		}
		return nil, err
	}

	isStaff := ad.Role == "admin" || ad.Role == "moderator"
	if !isStaff && ownerID != ad.UserID {
		return nil, &errs.Error{Code: errs.NotFound, Message: "ticket not found"}
	}

	var m SupportMessage
	err := db.QueryRow(ctx, `
		INSERT INTO support_messages (ticket_id, sender_id, is_staff, body)
		VALUES ($1, $2, $3, $4)
		RETURNING id, ticket_id, sender_id, is_staff, body, created_at
	`, id, ad.UserID, isStaff, p.Body).Scan(&m.ID, &m.TicketID, &m.SenderID, &m.IsStaff, &m.Body, &m.CreatedAt)
	if err != nil {
		return nil, err
	}

	// If staff replied, notify the ticket owner.
	if isStaff && ownerID != ad.UserID {
		var subject string
		db.QueryRow(ctx, `SELECT subject FROM support_tickets WHERE id = $1`, id).Scan(&subject)
		myauth.NotifySupportReply(ctx, &myauth.NotifySupportReplyParams{UserID: ownerID, Subject: subject})
	}

	return &m, nil
}

// ----- Admin support -----

type AdminListTicketsParams struct {
	Status string `query:"status"`
}

//encore:api auth method=GET path=/admin/support/tickets
func AdminListTickets(ctx context.Context, p *AdminListTicketsParams) (*ListTicketsResponse, error) {
	ad := auth.Data().(*myauth.AuthData)
	if ad.Role != "admin" && ad.Role != "moderator" {
		return nil, &errs.Error{Code: errs.PermissionDenied, Message: "admin or moderator only"}
	}

	if p.Status != "" {
		return listTickets(ctx, `SELECT id, user_id, subject, status, priority, COALESCE(assigned_to::text, ''), created_at, resolved_at FROM support_tickets WHERE status = $1 ORDER BY created_at DESC`, p.Status)
	}
	return listTickets(ctx, `SELECT id, user_id, subject, status, priority, COALESCE(assigned_to::text, ''), created_at, resolved_at FROM support_tickets ORDER BY created_at DESC`)
}

type AssignTicketParams struct {
	AssignedTo string `json:"assigned_to"`
}

//encore:api auth method=POST path=/admin/support/tickets/:id/assign
func AssignTicket(ctx context.Context, id string, p *AssignTicketParams) error {
	ad := auth.Data().(*myauth.AuthData)
	if ad.Role != "admin" {
		return &errs.Error{Code: errs.PermissionDenied, Message: "admin only"}
	}
	_, err := db.Exec(ctx, `UPDATE support_tickets SET assigned_to = $1 WHERE id = $2`, nilIfEmpty(p.AssignedTo), id)
	return err
}

//encore:api auth method=POST path=/admin/support/tickets/:id/resolve
func ResolveTicket(ctx context.Context, id string) error {
	ad := auth.Data().(*myauth.AuthData)
	if ad.Role != "admin" && ad.Role != "moderator" {
		return &errs.Error{Code: errs.PermissionDenied, Message: "admin or moderator only"}
	}
	_, err := db.Exec(ctx, `UPDATE support_tickets SET status = 'resolved', resolved_at = now() WHERE id = $1`, id)
	return err
}

// ----- Broadcasts -----

type Broadcast struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	Target    string    `json:"target"`
	Tier      string    `json:"tier"`
	SentAt    *time.Time `json:"sent_at,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateBroadcastParams struct {
	Title  string `json:"title"`
	Body   string `json:"body"`
	Target string `json:"target"`
	Tier   string `json:"tier"`
}

type ListBroadcastsResponse struct {
	Broadcasts []Broadcast `json:"broadcasts"`
}

//encore:api auth method=POST path=/admin/broadcasts
func CreateBroadcast(ctx context.Context, p *CreateBroadcastParams) (*Broadcast, error) {
	ad := auth.Data().(*myauth.AuthData)
	if ad.Role != "admin" {
		return nil, &errs.Error{Code: errs.PermissionDenied, Message: "admin only"}
	}
	if p.Title == "" || p.Body == "" {
		return nil, &errs.Error{Code: errs.InvalidArgument, Message: "title and body are required"}
	}
	target := p.Target
	if target == "" {
		target = "all"
	}

	var b Broadcast
	err := db.QueryRow(ctx, `
		INSERT INTO broadcasts (title, body, target, tier, sent_at)
		VALUES ($1, $2, $3, $4, now())
		RETURNING id, title, body, target, tier, sent_at, created_at
	`, p.Title, p.Body, target, p.Tier).Scan(&b.ID, &b.Title, &b.Body, &b.Target, &b.Tier, &b.SentAt, &b.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &b, nil
}

//encore:api auth method=GET path=/admin/broadcasts
func ListBroadcasts(ctx context.Context) (*ListBroadcastsResponse, error) {
	ad := auth.Data().(*myauth.AuthData)
	if ad.Role != "admin" {
		return nil, &errs.Error{Code: errs.PermissionDenied, Message: "admin only"}
	}

	rows, err := db.Query(ctx, `SELECT id, title, body, target, tier, sent_at, created_at FROM broadcasts ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var broadcasts []Broadcast
	for rows.Next() {
		var b Broadcast
		if err := rows.Scan(&b.ID, &b.Title, &b.Body, &b.Target, &b.Tier, &b.SentAt, &b.CreatedAt); err != nil {
			return nil, err
		}
		broadcasts = append(broadcasts, b)
	}
	if broadcasts == nil {
		broadcasts = []Broadcast{}
	}
	return &ListBroadcastsResponse{Broadcasts: broadcasts}, rows.Err()
}

// GetAnnouncements returns broadcasts the caller should see: those targeting
// everyone (`target = 'all'`) plus those targeting their tier (`target = 'tier'`
// and matching `tier`). Anonymous callers only see the `all` broadcasts.
//encore:api public method=GET path=/announcements
func GetAnnouncements(ctx context.Context) (*ListBroadcastsResponse, error) {
	ad, hasAuth := auth.Data().(*myauth.AuthData)
	tier := ""
	if hasAuth {
		tier = ad.Tier
	}

	rows, err := db.Query(ctx, `
		SELECT id, title, body, target, tier, sent_at, created_at
		FROM broadcasts
		WHERE target = 'all' OR (target = 'tier' AND tier = $1)
		ORDER BY created_at DESC
		LIMIT 10
	`, tier)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var broadcasts []Broadcast
	for rows.Next() {
		var b Broadcast
		if err := rows.Scan(&b.ID, &b.Title, &b.Body, &b.Target, &b.Tier, &b.SentAt, &b.CreatedAt); err != nil {
			return nil, err
		}
		broadcasts = append(broadcasts, b)
	}
	if broadcasts == nil {
		broadcasts = []Broadcast{}
	}
	return &ListBroadcastsResponse{Broadcasts: broadcasts}, rows.Err()
}

// ----- Helpers -----

func listTickets(ctx context.Context, query string, args ...interface{}) (*ListTicketsResponse, error) {
	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tickets []SupportTicket
	for rows.Next() {
		var t SupportTicket
		var resolvedAt *time.Time
		if err := rows.Scan(&t.ID, &t.UserID, &t.Subject, &t.Status, &t.Priority, &t.AssignedTo, &t.CreatedAt, &resolvedAt); err != nil {
			return nil, err
		}
		t.ResolvedAt = resolvedAt
		tickets = append(tickets, t)
	}
	if tickets == nil {
		tickets = []SupportTicket{}
	}
	return &ListTicketsResponse{Tickets: tickets}, rows.Err()
}

func listMessages(ctx context.Context, ticketID string) ([]SupportMessage, error) {
	rows, err := db.Query(ctx, `
		SELECT id, ticket_id, sender_id, is_staff, body, created_at
		FROM support_messages WHERE ticket_id = $1 ORDER BY created_at ASC
	`, ticketID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []SupportMessage
	for rows.Next() {
		var m SupportMessage
		if err := rows.Scan(&m.ID, &m.TicketID, &m.SenderID, &m.IsStaff, &m.Body, &m.CreatedAt); err != nil {
			return nil, err
		}
		messages = append(messages, m)
	}
	if messages == nil {
		messages = []SupportMessage{}
	}
	return messages, rows.Err()
}

func nilIfEmpty(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}
