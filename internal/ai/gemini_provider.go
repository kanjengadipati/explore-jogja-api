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

const defaultGeminiBaseURL = "https://generativelanguage.googleapis.com"

type GeminiProvider struct {
	baseURL string
	apiKey  string
	client  *http.Client
}

type geminiGenerateRequest struct {
	SystemInstruction *geminiContent         `json:"systemInstruction,omitempty"`
	Contents          []geminiContent        `json:"contents"`
	GenerationConfig  geminiGenerationConfig `json:"generationConfig"`
}

type geminiContent struct {
	Role  string       `json:"role,omitempty"`
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text             string `json:"text"`
	ThoughtSignature string `json:"thoughtSignature,omitempty"` // present in thinking model parts; skip these
}

type geminiGenerationConfig struct {
	Temperature      *float64       `json:"temperature,omitempty"`
	MaxOutputTokens  int            `json:"maxOutputTokens,omitempty"`
	ResponseMimeType string         `json:"responseMimeType"`
	ResponseSchema   map[string]any `json:"responseSchema"`
}

type geminiGenerateResponse struct {
	Candidates []geminiCandidate `json:"candidates"`
	Error      *geminiErrorItem  `json:"error"`
}

type geminiCandidate struct {
	Content geminiContent `json:"content"`
}

type geminiErrorItem struct {
	Message string `json:"message"`
	Status  string `json:"status"`
}

func NewGeminiProvider(baseURL, apiKey string, timeout time.Duration) *GeminiProvider {
	return &GeminiProvider{
		baseURL: normalizeBaseURL(baseURL, defaultGeminiBaseURL),
		apiKey:  apiKey,
		client:  &http.Client{Timeout: timeout},
	}
}

func (p *GeminiProvider) Generate(ctx context.Context, input GenerateInput) (*GenerateResult, error) {
	reqBody := geminiGenerateRequest{
		Contents: []geminiContent{
			{
				Role: "user",
				Parts: []geminiPart{
					{Text: input.UserPrompt},
				},
			},
		},
		GenerationConfig: geminiGenerationConfig{
			ResponseMimeType: "application/json",
			ResponseSchema:   input.ResponseSchema, // nil = free-form JSON; set schema for structured output
			MaxOutputTokens:  input.MaxTokens,
		},
	}
	if strings.TrimSpace(input.SystemPrompt) != "" {
		reqBody.SystemInstruction = &geminiContent{
			Parts: []geminiPart{
				{Text: input.SystemPrompt},
			},
		}
	}
	if input.Temperature > 0 {
		reqBody.GenerationConfig.Temperature = &input.Temperature
	}

	body, err := marshalBody(reqBody)
	if err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf("%s/v1beta/models/%s:generateContent", p.baseURL, input.Model)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-goog-api-key", p.apiKey)

	resp, err := p.client.Do(req)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, fmt.Errorf("%w: %v", ErrTimeout, err)
		}
		var netErr net.Error
		if errors.As(err, &netErr) && netErr.Timeout() {
			return nil, fmt.Errorf("%w: %v", ErrTimeout, err)
		}
		return nil, fmt.Errorf("gemini is unavailable: %v", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var parsed geminiGenerateResponse
	if err := json.Unmarshal(bodyBytes, &parsed); err != nil {
		return nil, fmt.Errorf("failed to decode gemini response: %s", strings.TrimSpace(string(bodyBytes)))
	}

	if resp.StatusCode >= 400 {
		if parsed.Error != nil && parsed.Error.Message != "" {
			return nil, fmt.Errorf("gemini error: %s", parsed.Error.Message)
		}
		return nil, fmt.Errorf("gemini returned status %d", resp.StatusCode)
	}

	text := extractGeminiText(parsed.Candidates)
	if text == "" {
		return nil, ErrInvalidStructuredOutput
	}

	return &GenerateResult{Text: text}, nil
}

func extractGeminiText(candidates []geminiCandidate) string {
	// First pass: prefer parts without thoughtSignature (non-thinking parts)
	for _, candidate := range candidates {
		for _, part := range candidate.Content.Parts {
			if part.ThoughtSignature == "" && strings.TrimSpace(part.Text) != "" {
				return part.Text
			}
		}
	}
	// Second pass: if all parts have thoughtSignature (thinking model), return the one with JSON-like text
	for _, candidate := range candidates {
		for _, part := range candidate.Content.Parts {
			trimmed := strings.TrimSpace(part.Text)
			if trimmed != "" && (strings.HasPrefix(trimmed, "{") || strings.HasPrefix(trimmed, "[")) {
				return part.Text
			}
		}
	}
	// Final fallback: return any non-empty text
	for _, candidate := range candidates {
		for _, part := range candidate.Content.Parts {
			if strings.TrimSpace(part.Text) != "" {
				return part.Text
			}
		}
	}
	return ""
}
