package social

import (
	"context"
	"testing"

	"comics-galore/backend/fixtures"

	"encore.dev/et"
)

func TestStartConversation_AndList(t *testing.T) {
	_, _ = et.NewTestDatabase(context.Background(), "socialdb")

	userA := fixtures.TierGatedCtx("550e8400-e29b-41d4-a716-4466554400a1", "user", "free")
	userB := fixtures.TierGatedCtx("550e8400-e29b-41d4-a716-4466554400a2", "user", "free")

	resp, err := StartConversation(userA, "550e8400-e29b-41d4-a716-4466554400a2", &StartConversationParams{Body: "hello"})
	if err != nil {
		t.Fatalf("start conversation error: %v", err)
	}
	if len(resp.Messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(resp.Messages))
	}

	// Same pair again → same conversation (no duplicate).
	resp2, err := StartConversation(userA, "550e8400-e29b-41d4-a716-4466554400a2", &StartConversationParams{Body: "again"})
	if err != nil {
		t.Fatalf("start again error: %v", err)
	}
	if resp2.Conversation.ID != resp.Conversation.ID {
		t.Errorf("expected same conversation id, got %s vs %s", resp2.Conversation.ID, resp.Conversation.ID)
	}
	if len(resp2.Messages) != 2 {
		t.Errorf("expected 2 messages, got %d", len(resp2.Messages))
	}

	// Both participants can list it.
	listA, _ := ListConversations(userA)
	if len(listA.Conversations) != 1 {
		t.Errorf("expected 1 conversation for A, got %d", len(listA.Conversations))
	}
	listB, _ := ListConversations(userB)
	if len(listB.Conversations) != 1 {
		t.Errorf("expected 1 conversation for B, got %d", len(listB.Conversations))
	}
}

func TestSendMessage(t *testing.T) {
	_, _ = et.NewTestDatabase(context.Background(), "socialdb")

	userA := fixtures.TierGatedCtx("550e8400-e29b-41d4-a716-4466554400b1", "user", "free")
	userB := fixtures.TierGatedCtx("550e8400-e29b-41d4-a716-4466554400b2", "user", "free")

	resp, _ := StartConversation(userA, "550e8400-e29b-41d4-a716-4466554400b2", &StartConversationParams{Body: "first"})

	msg, err := SendMessage(userB, resp.Conversation.ID, &SendMessageParams{Body: "reply"})
	if err != nil {
		t.Fatalf("send error: %v", err)
	}
	if msg.Body != "reply" {
		t.Errorf("expected body 'reply', got %q", msg.Body)
	}

	// B's message is unread for A until A opens the conversation.
	listA, _ := ListConversations(userA)
	if listA.Conversations[0].UnreadCount != 1 {
		t.Errorf("expected 1 unread for A, got %d", listA.Conversations[0].UnreadCount)
	}
}

func TestStartConversation_CannotMessageSelf(t *testing.T) {
	_, _ = et.NewTestDatabase(context.Background(), "socialdb")

	userA := fixtures.TierGatedCtx("550e8400-e29b-41d4-a716-4466554400c1", "user", "free")
	_, err := StartConversation(userA, "550e8400-e29b-41d4-a716-4466554400c1", &StartConversationParams{Body: "hi"})
	if err == nil {
		t.Fatal("expected error for self-message, got nil")
	}
}

func TestCreateTicket_AndReply(t *testing.T) {
	_, _ = et.NewTestDatabase(context.Background(), "socialdb")

	user := fixtures.TierGatedCtx("550e8400-e29b-41d4-a716-4466554400d1", "user", "free")
	admin := fixtures.AdminCtx()

	resp, err := CreateTicket(user, &CreateTicketParams{Subject: "help", Body: "I need help"})
	if err != nil {
		t.Fatalf("create ticket error: %v", err)
	}
	if resp.Ticket.Status != "open" {
		t.Errorf("expected open, got %s", resp.Ticket.Status)
	}

	// Staff replies.
	staffMsg, err := ReplyTicket(admin, resp.Ticket.ID, &ReplyTicketParams{Body: "How can we help?"})
	if err != nil {
		t.Fatalf("staff reply error: %v", err)
	}
	if !staffMsg.IsStaff {
		t.Error("expected staff reply is_staff=true")
	}

	// Admin list sees it.
	list, err := AdminListTickets(admin, &AdminListTicketsParams{})
	if err != nil {
		t.Fatalf("admin list error: %v", err)
	}
	if len(list.Tickets) < 1 {
		t.Error("expected at least 1 ticket in admin list")
	}

	// Resolve.
	if err := ResolveTicket(admin, resp.Ticket.ID); err != nil {
		t.Fatalf("resolve error: %v", err)
	}
	resolved, _ := GetTicket(admin, resp.Ticket.ID)
	if resolved.Ticket.Status != "resolved" {
		t.Errorf("expected resolved, got %s", resolved.Ticket.Status)
	}
}

func TestAdminListTickets_RequiresStaff(t *testing.T) {
	_, _ = et.NewTestDatabase(context.Background(), "socialdb")

	user := fixtures.TierGatedCtx("550e8400-e29b-41d4-a716-4466554400e1", "user", "free")
	_, err := AdminListTickets(user, &AdminListTicketsParams{})
	if err == nil {
		t.Fatal("expected permission error, got nil")
	}
}
