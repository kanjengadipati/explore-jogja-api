package search

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
	"log/slog"
)

// Result is a single search result returned by a search provider.
type Result struct {
	Title   string
	URL     string
	Content string // snippet
}

// Client is a search/search-and-research client. It is intentionally
// fail-open: when disabled/absent it returns no results instead of erroring,
// so AI generation can degrade gracefully to its non-grounded behavior.
//
// Multiple API keys may be supplied (APIKeys). The client round-robins across
// them and, on an auth/quota failure for one key, transparently retries with
// the next — so an exhausted key degrades gracefully instead of breaking
// generation entirely.
type Client struct {
	enabled    bool
	apiKeys     []string
	baseURL    string
	maxResults int
	httpClient *http.Client

	mu        sync.Mutex
	keyIndex  int
}

// Config configures a Client. A zero value produces a disabled client.
type Config struct {
	Enabled        bool
	APIKey         string   // legacy single-key path
	APIKeys        []string // preferred multi-key path; falls back to APIKey
	BaseURL        string
	MaxResults     int
	TimeoutSeconds int
}

// NewClient constructs a Tavily search client. If Tavily is disabled or no API
// key is supplied, a disabled (no-op) client is returned so callers never need
// to nil-check: calling Search on it simply yields no results.
func NewClient(cfg Config) *Client {
	keys := dedupNonEmpty(cfg.APIKeys)
	if len(keys) == 0 && strings.TrimSpace(cfg.APIKey) != "" {
		keys = []string{cfg.APIKey}
	}
	if !cfg.Enabled || len(keys) == 0 {
		return &Client{enabled: false}
	}
	timeout := 8
	if cfg.TimeoutSeconds > 0 {
		timeout = cfg.TimeoutSeconds
	}
	base := strings.TrimSpace(cfg.BaseURL)
	if base == "" {
		base = "https://api.tavily.com/search"
	}
	max := cfg.MaxResults
	if max <= 0 {
		max = 5
	}
	return &Client{
		enabled:    true,
		apiKeys:    keys,
		baseURL:    base,
		maxResults: max,
		httpClient: &http.Client{Timeout: time.Duration(timeout) * time.Second},
	}
}

func dedupNonEmpty(in []string) []string {
	seen := make(map[string]struct{}, len(in))
	out := make([]string, 0, len(in))
	for _, k := range in {
		k = strings.TrimSpace(k)
		if k == "" {
			continue
		}
		if _, ok := seen[k]; ok {
			continue
		}
		seen[k] = struct{}{}
		out = append(out, k)
	}
	return out
}

// Enabled reports whether the search client will actually hit the network.
func (c *Client) Enabled() bool { return c != nil && c.enabled }

// currentKey returns the active API key (advancing the round-robin index).
func (c *Client) currentKey() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.apiKeys) == 0 {
		return ""
	}
	k := c.apiKeys[c.keyIndex%len(c.apiKeys)]
	c.keyIndex = (c.keyIndex + 1) % len(c.apiKeys)
	return k
}

// rotateKey advances past the currently-failing key so the next attempt uses a
// fresh one. No-op if only a single key is configured.
func (c *Client) rotateKey() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.apiKeys) > 0 {
		c.keyIndex = (c.keyIndex + 1) % len(c.apiKeys)
	}
}

// isAuthOrQuotaStatus treats a status as "this key is bad / exhausted" so the
// client can retry with the next key.
func isAuthOrQuotaStatus(statusCode int) bool {
	return statusCode == http.StatusUnauthorized ||
		statusCode == http.StatusForbidden ||
		statusCode == http.StatusTooManyRequests
}

// Search performs a search-engine query and returns its results.
// It is fail-open: errors or empty results do not abort generation; the caller
// is expected to proceed without grounding context when this returns nothing.
func (c *Client) Search(ctx context.Context, query string) ([]Result, error) {
	if c == nil || !c.enabled {
		return nil, nil
	}
	if strings.TrimSpace(query) == "" {
		return nil, nil
	}

	var lastErr error
	// Try up to len(keys) different keys (one full round) before giving up.
	for attempt := 0; attempt < len(c.apiKeys); attempt++ {
		results, err := c.searchWithKey(ctx, query, c.currentKey())
		if err == nil {
			return results, nil
		}
		lastErr = err
		// Retry with a different key only if we have more keys and the failure
		// looks key-related (auth / quota). Network errors are also retried
		// since the next key routes through the same (possibly flaky) egress.
		if !isAuthOrQuotaStatus(errStatusCode(lastErr)) && !isNetError(lastErr) {
			break
		}
		if len(c.apiKeys) <= 1 {
			break
		}
		c.rotateKey()
	}
	if lastErr != nil {
		slog.WarnContext(ctx, "tavily search exhausted all keys", "error", lastErr, "query", query, "keys", len(c.apiKeys))
	}
	return nil, lastErr
}

// searchWithKey runs a single request using the supplied key.
func (c *Client) searchWithKey(ctx context.Context, query string, key string) ([]Result, error) {
	body, err := json.Marshal(map[string]any{
		"api_key":      key,
		"query":        query,
		"max_results":  c.maxResults,
		"search_depth": "basic",
	})
	if err != nil {
		return nil, fmt.Errorf("encode tavily request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("build tavily request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+key)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _, _ = io.Copy(io.Discard, resp.Body); _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		slog.WarnContext(ctx, "tavily search returned non-200", "status", resp.StatusCode, "body", string(b), "query", query)
		return nil, &tavilyError{status: resp.StatusCode, body: string(b)}
	}

	var payload struct {
		Results []struct {
			Title   string `json:"title"`
			URL     string `json:"url"`
			Content string `json:"content"`
		} `json:"results"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		slog.WarnContext(ctx, "tavily response decode failed", "error", err)
		return nil, fmt.Errorf("decode tavily response: %w", err)
	}

	results := make([]Result, 0, len(payload.Results))
	for _, r := range payload.Results {
		if r.Title == "" && r.URL == "" {
			continue
		}
		results = append(results, Result{
			Title:   strings.TrimSpace(r.Title),
			URL:     strings.TrimSpace(r.URL),
			Content: strings.TrimSpace(r.Content),
		})
	}
	slog.DebugContext(ctx, "tavily search completed", "results", len(results), "query", query)
	return results, nil
}

// GroundedContext runs a search against client (if enabled) and returns a
// research block ready to inject into an AI prompt. Fail-open: returns "" on
// any error or when the client is disabled/nil, so callers can proceed
// ungrounded rather than aborting generation.
func GroundedContext(ctx context.Context, client *Client, query string) string {
	results, err := client.Search(ctx, query)
	if err != nil || len(results) == 0 {
		return ""
	}
	return FormatResultsForPrompt(query, results)
}

type tavilyError struct {
	status int
	body   string
}

func (e *tavilyError) Error() string {
	return fmt.Sprintf("tavily search status %d: %s", e.status, e.body)
}

func errStatusCode(err error) int {
	if err == nil {
		return 0
	}
	if te, ok := err.(*tavilyError); ok {
		return te.status
	}
	return 0
}

func isNetError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "timeout") ||
		strings.Contains(err.Error(), "connection refused") ||
		strings.Contains(err.Error(), "no such host")
}

// FormatResultsForPrompt renders search results into a grounded research block
// suitable for injection into an AI system/user prompt. Total output is bounded
// so it never dominates the context window. Each snippet is truncated to
// maxSnippetLen runes.
func FormatResultsForPrompt(query string, results []Result) string {
	if len(results) == 0 {
		return ""
	}
	const maxSnippetLen = 320
	var b strings.Builder
	b.WriteString("GROUNDED RESEARCH CONTEXT (use ONLY the facts present below; do not invent details, prices, hours, ratings, or review counts beyond what is written here):\n")
	b.WriteString(fmt.Sprintf("Search query: \"%s\"\n\n", query))
	idx := 1
	for _, r := range results {
		b.WriteString(fmt.Sprintf("## Finding %d\n", idx))
		if r.Title != "" {
			b.WriteString(fmt.Sprintf("Title: %s\n", r.Title))
		}
		if r.URL != "" {
			b.WriteString(fmt.Sprintf("Source: %s\n", r.URL))
		}
		content := r.Content
		if len([]rune(content)) > maxSnippetLen {
			content = string([]rune(content)[:maxSnippetLen]) + "…"
		}
		if content != "" {
			b.WriteString(fmt.Sprintf("Snippet: %s\n\n", content))
		} else {
			b.WriteString("\n")
		}
		idx++
		if idx > 5 {
			break
		}
	}
	b.WriteString("IMPORTANT: Only treat as factual the information present in the snippets above. Do not fabricate opening hours, ticket prices, coordinates, ratings, or review counts. Omit a fact entirely if it is not stated here.\n")
	return b.String()
}
