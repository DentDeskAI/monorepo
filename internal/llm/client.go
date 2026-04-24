// Package llm — фасад над LLM-провайдерами. Интерфейс один, реализаций три:
// anthropic (Claude), groq (Llama), mock (детерминированный для тестов/dev).
package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
)

type Message struct {
	Role    Role   `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	System     string
	Messages   []Message
	Temperature float32
	MaxTokens   int
	// JSON режим — требуем от модели чистый JSON.
	JSONOnly bool
}

type ChatResponse struct {
	Text        string
	InputTokens int
	OutputTokens int
}

type Client interface {
	Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error)
}

// -------- Anthropic (Claude) --------

type AnthropicClient struct {
	apiKey string
	model  string
	http   *http.Client
}

func NewAnthropic(apiKey, model string) *AnthropicClient {
	return &AnthropicClient{
		apiKey: apiKey,
		model:  model,
		http:   &http.Client{Timeout: 30 * time.Second},
	}
}

type anthropicReq struct {
	Model       string    `json:"model"`
	MaxTokens   int       `json:"max_tokens"`
	System      string    `json:"system,omitempty"`
	Messages    []anthMsg `json:"messages"`
	Temperature float32   `json:"temperature,omitempty"`
}
type anthMsg struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}
type anthropicResp struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Usage struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func (c *AnthropicClient) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	if c.apiKey == "" {
		return nil, errors.New("anthropic: API key not set")
	}
	max := req.MaxTokens
	if max == 0 {
		max = 512
	}
	temp := req.Temperature
	if temp == 0 {
		temp = 0.4
	}

	msgs := make([]anthMsg, 0, len(req.Messages))
	for _, m := range req.Messages {
		role := "user"
		if m.Role == RoleAssistant {
			role = "assistant"
		}
		msgs = append(msgs, anthMsg{Role: role, Content: m.Content})
	}

	body, _ := json.Marshal(anthropicReq{
		Model:       c.model,
		MaxTokens:   max,
		System:      req.System,
		Messages:    msgs,
		Temperature: temp,
	})

	httpReq, _ := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://api.anthropic.com/v1/messages", bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", c.apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	resp, err := c.http.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("anthropic http: %w", err)
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("anthropic status %d: %s", resp.StatusCode, string(raw))
	}
	var out anthropicResp
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("anthropic decode: %w", err)
	}
	if out.Error != nil {
		return nil, errors.New(out.Error.Message)
	}
	var text strings.Builder
	for _, c := range out.Content {
		if c.Type == "text" {
			text.WriteString(c.Text)
		}
	}
	return &ChatResponse{
		Text:         strings.TrimSpace(text.String()),
		InputTokens:  out.Usage.InputTokens,
		OutputTokens: out.Usage.OutputTokens,
	}, nil
}

// -------- Groq (OpenAI-compatible) --------

type GroqClient struct {
	apiKey string
	model  string
	http   *http.Client
}

func NewGroq(apiKey, model string) *GroqClient {
	return &GroqClient{apiKey: apiKey, model: model, http: &http.Client{Timeout: 30 * time.Second}}
}

type groqReq struct {
	Model       string    `json:"model"`
	Messages    []anthMsg `json:"messages"`
	Temperature float32   `json:"temperature,omitempty"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	// JSON режим (OpenAI-compatible)
	ResponseFormat *struct {
		Type string `json:"type"`
	} `json:"response_format,omitempty"`
}
type groqResp struct {
	Choices []struct {
		Message anthMsg `json:"message"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
	} `json:"usage"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func (c *GroqClient) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	if c.apiKey == "" {
		return nil, errors.New("groq: API key not set")
	}
	msgs := []anthMsg{}
	if req.System != "" {
		msgs = append(msgs, anthMsg{Role: "system", Content: req.System})
	}
	for _, m := range req.Messages {
		msgs = append(msgs, anthMsg{Role: string(m.Role), Content: m.Content})
	}
	payload := groqReq{
		Model:       c.model,
		Messages:    msgs,
		Temperature: req.Temperature,
		MaxTokens:   req.MaxTokens,
	}
	if req.JSONOnly {
		payload.ResponseFormat = &struct {
			Type string `json:"type"`
		}{Type: "json_object"}
	}
	body, _ := json.Marshal(payload)

	httpReq, _ := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://api.groq.com/openai/v1/chat/completions", bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.http.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("groq http: %w", err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("groq status %d: %s", resp.StatusCode, string(raw))
	}
	var out groqResp
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, err
	}
	if out.Error != nil {
		return nil, errors.New(out.Error.Message)
	}
	if len(out.Choices) == 0 {
		return nil, errors.New("groq: empty response")
	}
	return &ChatResponse{
		Text:         strings.TrimSpace(out.Choices[0].Message.Content),
		InputTokens:  out.Usage.PromptTokens,
		OutputTokens: out.Usage.CompletionTokens,
	}, nil
}

// -------- Mock --------

// MockClient — для локальной разработки без ключей.
// Делает примитивный intent detection по ключевым словам.
type MockClient struct{}

func NewMock() *MockClient { return &MockClient{} }

func (c *MockClient) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	// собираем последнее пользовательское сообщение
	var last string
	for i := len(req.Messages) - 1; i >= 0; i-- {
		if req.Messages[i].Role == RoleUser {
			last = strings.ToLower(req.Messages[i].Content)
			break
		}
	}
	if req.JSONOnly {
		// имитируем JSON-ответ для intent detection
		return &ChatResponse{Text: mockClassifyJSON(last)}, nil
	}
	// генерим текст Айгуль
	return &ChatResponse{Text: mockReply(last)}, nil
}

func mockClassifyJSON(msg string) string {
	intent := "other"
	switch {
	case containsAny(msg, "болит", "ноет", "кровоточ", "опух"):
		intent = "urgent_pain"
	case containsAny(msg, "запис", "прийти", "попаст", "прием"):
		intent = "booking"
	case containsAny(msg, "отмен", "перенес"):
		intent = "reschedule"
	case containsAny(msg, "привет", "салем", "здрав"):
		intent = "greeting"
	case containsAny(msg, "цена", "стоимо", "скольк"):
		intent = "pricing"
	}
	return fmt.Sprintf(`{"intent":"%s","service":null,"doctor":null,"when":null,"language":"ru"}`, intent)
}

func mockReply(msg string) string {
	switch {
	case containsAny(msg, "болит", "ноет", "опух"):
		return "Понимаю, это неприятно. Давайте подберём время — могу записать вас на сегодня или завтра. Какое время удобно?"
	case containsAny(msg, "запис", "прийти"):
		return "Конечно, помогу записаться! Ближайшие свободные окна: сегодня в 17:00, завтра в 10:00 и 15:00. Какое подходит?"
	case containsAny(msg, "привет", "салем"):
		return "Здравствуйте! Я Айгуль, администратор клиники. Подскажите, чем могу помочь — записать вас на приём?"
	default:
		return "Поняла вас. Подскажите, хотите записаться на приём? Я помогу подобрать удобное время."
	}
}

func containsAny(s string, subs ...string) bool {
	for _, sub := range subs {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}
