// Package aiprovider is a minimal client for any OpenAI-compatible
// chat-completions endpoint, used for optional AI moderation (ADR 0018).
//
// It is a shared non-service package (like `nowpayments`), imported by the
// comics service which owns the moderation decision flow.
package aiprovider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	endpoint string
	apiKey   string
	model    string
	http     *http.Client
}

// New returns a client for the given OpenAI-compatible endpoint, API key and
// model. endpoint is the full chat-completions URL, e.g.
// https://api.openai.com/v1/chat/completions.
func New(endpoint, apiKey, model string) *Client {
	if endpoint == "" {
		endpoint = "https://api.openai.com/v1/chat/completions"
	}
	return &Client{
		endpoint: endpoint,
		apiKey:   apiKey,
		model:    model,
		http:     &http.Client{Timeout: 30 * time.Second},
	}
}

// ClassifyRequest asks the model to classify a piece of content.
type ClassifyRequest struct {
	// SystemPrompt is the instruction set (e.g. "You moderate comics...").
	SystemPrompt string
	// Content is the text to classify.
	Content string
}

// ClassifyResult is the parsed moderation decision.
type ClassifyResult struct {
	Decision   string  // approved | rejected | uncertain
	Confidence float64 // 0..1
	Reason     string
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model    string        `json:"model"`
	Messages []chatMessage `json:"messages"`
}

type chatResponse struct {
	Choices []struct {
		Message chatMessage `json:"message"`
	} `json:"choices"`
}

// Classify sends the content to the model and parses a JSON moderation verdict.
// The model is instructed to reply with only
// {"decision":"approved|rejected|uncertain","confidence":0.0,"reason":"..."}.
func (c *Client) Classify(ctx context.Context, req ClassifyRequest) (*ClassifyResult, error) {
	if c.apiKey == "" {
		return nil, fmt.Errorf("aiprovider: no API key configured")
	}

	messages := []chatMessage{
		{Role: "system", Content: req.SystemPrompt},
		{Role: "user", Content: req.Content},
	}

	body, _ := json.Marshal(chatRequest{Model: c.model, Messages: messages})
	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.http.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("aiprovider: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("aiprovider: error %d: %s", resp.StatusCode, string(respBody))
	}

	var parsed chatResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return nil, fmt.Errorf("aiprovider: parse: %w", err)
	}
	if len(parsed.Choices) == 0 {
		return nil, fmt.Errorf("aiprovider: empty response")
	}

	return parseVerdict(parsed.Choices[0].Message.Content)
}

// parseVerdict extracts the JSON object from the model reply (tolerant of
// surrounding prose or code fences) and normalises it.
func parseVerdict(raw string) (*ClassifyResult, error) {
	s := raw
	if i := strings.Index(s, "{"); i >= 0 {
		if j := strings.LastIndex(s, "}"); j > i {
			s = s[i : j+1]
		}
	}

	var v struct {
		Decision   string  `json:"decision"`
		Confidence float64 `json:"confidence"`
		Reason     string  `json:"reason"`
	}
	if err := json.Unmarshal([]byte(s), &v); err != nil {
		return nil, fmt.Errorf("aiprovider: verdict parse: %w", err)
	}

	decision := strings.ToLower(strings.TrimSpace(v.Decision))
	switch decision {
	case "approved", "approve", "allow":
		decision = "approved"
	case "rejected", "reject", "block":
		decision = "rejected"
	default:
		decision = "uncertain"
	}

	return &ClassifyResult{Decision: decision, Confidence: v.Confidence, Reason: v.Reason}, nil
}
