package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

const defaultGroqBaseURL = "https://api.groq.com/openai"

type GroqProvider struct {
	baseURL string
	apiKey  string
	client  *http.Client
}

type groqChatRequest struct {
	Model          string              `json:"model"`
	Messages       []groqChatMessage   `json:"messages"`
	Temperature    float64             `json:"temperature,omitempty"`
	MaxTokens      int                 `json:"max_tokens,omitempty"`
	ResponseFormat *groqResponseFormat `json:"response_format,omitempty"`
}

type groqChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type groqResponseFormat struct {
	Type string `json:"type"`
}

type groqChatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func NewGroqProvider(baseURL, apiKey string, timeout time.Duration) *GroqProvider {
	return &GroqProvider{
		baseURL: normalizeBaseURL(baseURL, defaultGroqBaseURL),
		apiKey:  apiKey,
		client:  &http.Client{Timeout: timeout},
	}
}

func (p *GroqProvider) Generate(ctx context.Context, input GenerateInput) (*GenerateResult, error) {
	messages := []groqChatMessage{}
	if strings.TrimSpace(input.SystemPrompt) != "" {
		messages = append(messages, groqChatMessage{
			Role:    "system",
			Content: input.SystemPrompt,
		})
	}
	messages = append(messages, groqChatMessage{
		Role:    "user",
		Content: input.UserPrompt,
	})

	reqBody := groqChatRequest{
		Model:       input.Model,
		Messages:    messages,
		Temperature: input.Temperature,
		MaxTokens:   input.MaxTokens,
	}

	// Groq supports JSON mode if requested
	reqBody.ResponseFormat = &groqResponseFormat{Type: "json_object"}

	body, err := marshalBody(reqBody)
	if err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf("%s/v1/chat/completions", p.baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.client.Do(req)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, fmt.Errorf("%w: %v", ErrTimeout, err)
		}
		var netErr net.Error
		if errors.As(err, &netErr) && netErr.Timeout() {
			return nil, fmt.Errorf("%w: %v", ErrTimeout, err)
		}
		return nil, fmt.Errorf("groq is unavailable: %v", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var parsed groqChatResponse
	if err := json.Unmarshal(bodyBytes, &parsed); err != nil {
		return nil, fmt.Errorf("failed to decode groq response: %s", strings.TrimSpace(string(bodyBytes)))
	}

	if resp.StatusCode >= 400 {
		if parsed.Error != nil && parsed.Error.Message != "" {
			return nil, fmt.Errorf("groq error: %s", parsed.Error.Message)
		}
		return nil, fmt.Errorf("groq returned status %d", resp.StatusCode)
	}

	if len(parsed.Choices) == 0 || parsed.Choices[0].Message.Content == "" {
		return nil, ErrInvalidStructuredOutput
	}

	return &GenerateResult{Text: parsed.Choices[0].Message.Content}, nil
}
