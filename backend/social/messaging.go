package social

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"encore.dev/beta/auth"
	myauth "comics-galore/backend/auth"

	"encore.dev/beta/errs"
)

// ----- Types -----

type Conversation struct {
	ID             string    `json:"id"`
	OtherUserID    string    `json:"other_user_id"`
	LastMessage    string    `json:"last_message"`
	UnreadCount    int       `json:"unread_count"`
	LastMessageAt  time.Time `json:"last_message_at,omitempty"`
}

type Message struct {
	ID             string    `json:"id"`
	ConversationID string    `json:"conversation_id"`
	SenderID       string    `json:"sender_id"`
	Body           string    `json:"body"`
	ReadAt         *time.Time `json:"read_at,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}

type ListConversationsResponse struct {
	Conversations []Conversation `json:"conversations"`
}

type ConversationMessagesResponse struct {
	Conversation Conversation `json:"conversation"`
	Messages     []Message    `json:"messages"`
}

type SendMessageParams struct {
	Body string `json:"body"`
}

type StartConversationParams struct {
	Body string `json:"body"`
}

// ----- In-memory pub/sub for live message streams (mirrors comments SSE) -----

var (
	streamMu       sync.Mutex
	streamChannels = make(map[string][]chan Message)
)

func subscribeUser(userID string) chan Message {
	streamMu.Lock()
	defer streamMu.Unlock()
	ch := make(chan Message, 8)
	streamChannels[userID] = append(streamChannels[userID], ch)
	return ch
}

func unsubscribeUser(userID string, ch chan Message) {
	streamMu.Lock()
	defer streamMu.Unlock()
	list := streamChannels[userID]
	for i, c := range list {
		if c == ch {
			streamChannels[userID] = append(list[:i], list[i+1:]...)
			break
		}
	}
}

func publishMessage(userID string, m Message) {
	streamMu.Lock()
	defer streamMu.Unlock()
	for _, ch := range streamChannels[userID] {
		select {
		case ch <- m:
		default:
		}
	}
}

// ----- Helpers -----

// participants returns the normalised (a, b) pair for a conversation so the
// unique index prevents duplicate pairwise threads.
func normalizePair(a, b string) (string, string) {
	if a < b {
		return a, b
	}
	return b, a
}

// ----- Endpoints -----

//encore:api auth method=GET path=/messages/conversations
func ListConversations(ctx context.Context) (*ListConversationsResponse, error) {
	ad := auth.Data().(*myauth.AuthData)

	rows, err := db.Query(ctx, `
		SELECT c.id,
			CASE WHEN c.participant_a = $1 THEN c.participant_b ELSE c.participant_a END AS other_id,
			COALESCE((SELECT m.body FROM messages m WHERE m.conversation_id = c.id ORDER BY m.created_at DESC LIMIT 1), ''),
			(SELECT COUNT(*) FROM messages m WHERE m.conversation_id = c.id AND m.sender_id != $1 AND m.read_at IS NULL),
			c.last_message_at
		FROM conversations c
		WHERE c.participant_a = $1 OR c.participant_b = $1
		ORDER BY c.last_message_at DESC NULLS LAST
	`, ad.UserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var conversations []Conversation
	for rows.Next() {
		var c Conversation
		var lastAt *time.Time
		if err := rows.Scan(&c.ID, &c.OtherUserID, &c.LastMessage, &c.UnreadCount, &lastAt); err != nil {
			return nil, err
		}
		if lastAt != nil {
			c.LastMessageAt = *lastAt
		}
		conversations = append(conversations, c)
	}
	if conversations == nil {
		conversations = []Conversation{}
	}
	return &ListConversationsResponse{Conversations: conversations}, rows.Err()
}

//encore:api auth method=POST path=/messages/start/:userId
func StartConversation(ctx context.Context, userId string, p *StartConversationParams) (*ConversationMessagesResponse, error) {
	ad := auth.Data().(*myauth.AuthData)
	if userId == ad.UserID {
		return nil, &errs.Error{Code: errs.InvalidArgument, Message: "cannot message yourself"}
	}
	if p.Body == "" {
		return nil, &errs.Error{Code: errs.InvalidArgument, Message: "body is required"}
	}

	a, b := normalizePair(ad.UserID, userId)

	var convID string
	err := db.QueryRow(ctx, `
		INSERT INTO conversations (participant_a, participant_b, last_message_at)
		VALUES ($1, $2, now())
		ON CONFLICT (participant_a, participant_b)
		DO UPDATE SET last_message_at = now()
		RETURNING id
	`, a, b).Scan(&convID)
	if err != nil {
		return nil, err
	}

	var msg Message
	err = db.QueryRow(ctx, `
		INSERT INTO messages (conversation_id, sender_id, body)
		VALUES ($1, $2, $3)
		RETURNING id, conversation_id, sender_id, body, read_at, created_at
	`, convID, ad.UserID, p.Body).Scan(&msg.ID, &msg.ConversationID, &msg.SenderID, &msg.Body, &msg.ReadAt, &msg.CreatedAt)
	if err != nil {
		return nil, err
	}

	publishMessage(userId, msg)

	return getConversationMessages(ctx, convID, ad.UserID)
}

//encore:api auth method=GET path=/messages/conversation/:id
func GetConversation(ctx context.Context, id string) (*ConversationMessagesResponse, error) {
	ad := auth.Data().(*myauth.AuthData)
	if !isParticipant(ctx, id, ad.UserID) {
		return nil, &errs.Error{Code: errs.NotFound, Message: "conversation not found"}
	}

	// Mark my unread messages as read.
	db.Exec(ctx, `UPDATE messages SET read_at = now() WHERE conversation_id = $1 AND sender_id != $2 AND read_at IS NULL`, id, ad.UserID)

	return getConversationMessages(ctx, id, ad.UserID)
}

//encore:api auth method=POST path=/messages/conversation/:id
func SendMessage(ctx context.Context, id string, p *SendMessageParams) (*Message, error) {
	ad := auth.Data().(*myauth.AuthData)
	if p.Body == "" {
		return nil, &errs.Error{Code: errs.InvalidArgument, Message: "body is required"}
	}
	if !isParticipant(ctx, id, ad.UserID) {
		return nil, &errs.Error{Code: errs.NotFound, Message: "conversation not found"}
	}

	var msg Message
	err := db.QueryRow(ctx, `
		INSERT INTO messages (conversation_id, sender_id, body)
		VALUES ($1, $2, $3)
		RETURNING id, conversation_id, sender_id, body, read_at, created_at
	`, id, ad.UserID, p.Body).Scan(&msg.ID, &msg.ConversationID, &msg.SenderID, &msg.Body, &msg.ReadAt, &msg.CreatedAt)
	if err != nil {
		return nil, err
	}

	db.Exec(ctx, `UPDATE conversations SET last_message_at = now() WHERE id = $1`, id)

	// Notify the other participant via the live stream.
	var otherID string
	db.QueryRow(ctx, `
		SELECT CASE WHEN participant_a = $1 THEN participant_b ELSE participant_a END
		FROM conversations WHERE id = $2
	`, ad.UserID, id).Scan(&otherID)
	if otherID != "" {
		publishMessage(otherID, msg)
	}

	return &msg, nil
}

//encore:api auth method=POST path=/messages/conversation/:id/read
func MarkConversationRead(ctx context.Context, id string) error {
	ad := auth.Data().(*myauth.AuthData)
	_, err := db.Exec(ctx, `UPDATE messages SET read_at = now() WHERE conversation_id = $1 AND sender_id != $2 AND read_at IS NULL`, id, ad.UserID)
	return err
}

//encore:api public raw method=GET path=/messages-stream/:userId
func MessageStream(w http.ResponseWriter, req *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	userID := req.PathValue("userId")
	ch := subscribeUser(userID)
	defer unsubscribeUser(userID, ch)

	heartbeat := time.NewTicker(15 * time.Second)
	defer heartbeat.Stop()

	for {
		select {
		case m := <-ch:
			data, _ := json.Marshal(m)
			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()
		case <-heartbeat.C:
			fmt.Fprintf(w, ":\n\n")
			flusher.Flush()
		case <-req.Context().Done():
			return
		}
	}
}

// ----- Internal helpers -----

func isParticipant(ctx context.Context, convID, userID string) bool {
	var exists bool
	db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM conversations WHERE id = $1 AND (participant_a = $2 OR participant_b = $2))`, convID, userID).Scan(&exists)
	return exists
}

func getConversationMessages(ctx context.Context, convID, userID string) (*ConversationMessagesResponse, error) {
	var otherID string
	err := db.QueryRow(ctx, `
		SELECT CASE WHEN participant_a = $1 THEN participant_b ELSE participant_a END
		FROM conversations WHERE id = $2
	`, userID, convID).Scan(&otherID)
	if err != nil {
		if isNoRows(err) {
			return nil, &errs.Error{Code: errs.NotFound, Message: "conversation not found"}
		}
		return nil, err
	}

	rows, err := db.Query(ctx, `
		SELECT id, conversation_id, sender_id, body, read_at, created_at
		FROM messages WHERE conversation_id = $1 ORDER BY created_at ASC
	`, convID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var m Message
		if err := rows.Scan(&m.ID, &m.ConversationID, &m.SenderID, &m.Body, &m.ReadAt, &m.CreatedAt); err != nil {
			return nil, err
		}
		messages = append(messages, m)
	}
	if messages == nil {
		messages = []Message{}
	}

	return &ConversationMessagesResponse{
		Conversation: Conversation{ID: convID, OtherUserID: otherID},
		Messages:     messages,
	}, rows.Err()
}
