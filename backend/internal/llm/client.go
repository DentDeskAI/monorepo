// Package llm provides a thin, provider-agnostic LLM client.
// Swap providers by changing LLM_PROVIDER in the .env — the interface stays the same.
package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/rs/zerolog/log"
)

// ReplyRequest contains the context needed to generate a patient reply.
type ReplyRequest struct {
	ClinicName  string
	PatientName string
	UserMessage string
	// History     []Message // extend for multi-turn conversations
}

// Client is a provider-agnostic OpenAI-compatible LLM client.
// Groq, OpenAI, Together, and Anthropic (via compatibility layer) all work.
type Client struct {
	httpClient *resty.Client
	model      string
	provider   string
}

// NewClient creates an LLM client for the configured provider.
func NewClient(baseURL, apiKey, model, provider string) *Client {
	rc := resty.New().
		SetBaseURL(baseURL).
		SetAuthToken(apiKey).
		SetHeader("Content-Type", "application/json")

	return &Client{
		httpClient: rc,
		model:      model,
		provider:   provider,
	}
}

// GenerateReply asks the model to generate a helpful clinic assistant reply.
func (c *Client) GenerateReply(ctx context.Context, req ReplyRequest) (string, error) {
	systemPrompt := buildSystemPrompt(req.ClinicName)

	messages := []chatMessage{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: req.UserMessage},
	}

	payload := chatRequest{
		Model:       c.model,
		Messages:    messages,
		MaxTokens:   512,
		Temperature: 0.4,
	}

	resp, err := c.httpClient.R().
		SetContext(ctx).
		SetBody(payload).
		Post("/chat/completions")

	if err != nil {
		return "", fmt.Errorf("llm request: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return "", fmt.Errorf("llm api error %d: %s", resp.StatusCode(), resp.String())
	}

	var result chatResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return "", fmt.Errorf("decode llm response: %w", err)
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("llm returned no choices")
	}

	reply := strings.TrimSpace(result.Choices[0].Message.Content)

	log.Debug().
		Str("provider", c.provider).
		Str("model", c.model).
		Int("tokens_used", result.Usage.TotalTokens).
		Msg("LLM reply generated")

	return reply, nil
}

// ─── System prompt ────────────────────────────────────────────────────────────

// buildSystemPrompt returns the dental clinic assistant system prompt.
// TODO: Make this configurable per-clinic (stored in DB or Redis).
func buildSystemPrompt(clinicName string) string {
	return fmt.Sprintf(`You are a friendly and professional dental clinic assistant for %s.

Your responsibilities:
- Answer questions about appointments, clinic hours, and services
- Help patients book, reschedule, or cancel appointments
- Provide basic information about common dental procedures
- Remind patients about upcoming visits
- Collect patient information when needed

Guidelines:
- Always respond in the same language the patient writes in
- Keep responses concise and warm (2-4 sentences for simple queries)
- Never provide specific medical diagnoses or detailed clinical advice
- If you cannot help with something, politely ask the patient to call the clinic
- Do NOT make up appointment times — always say you will have staff confirm

Clinic: %s`, clinicName, clinicName)
}

// ─── OpenAI-compatible request/response types ─────────────────────────────────

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model       string        `json:"model"`
	Messages    []chatMessage `json:"messages"`
	MaxTokens   int           `json:"max_tokens"`
	Temperature float64       `json:"temperature"`
}

type chatResponse struct {
	ID      string `json:"id"`
	Model   string `json:"model"`
	Choices []struct {
		Message      chatMessage `json:"message"`
		FinishReason string      `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}
