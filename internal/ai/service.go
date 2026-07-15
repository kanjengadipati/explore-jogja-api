package ai

import (
	"context"
	"errors"
	"time"

	"pleco-api/internal/config"
)

var ErrDisabled = errors.New("ai is disabled")
var ErrTimeout = errors.New("ai request timed out")
var ErrInvalidStructuredOutput = errors.New("ai returned an invalid structured response")

type Service struct {
	enabled          bool
	model            string
	providerName     string
	provider         Provider
	fallbackProvider Provider
	fallbackModel    string
	fallbackName     string
}

func NewService(cfg config.AIConfig) (*Service, error) {
	service := &Service{
		enabled:      cfg.Enabled,
		model:        cfg.Model,
		providerName: cfg.Provider,
	}

	if !cfg.Enabled {
		return service, nil
	}

	instantiateProvider := func(name, baseURL, apiKey string) (Provider, error) {
		switch name {
		case "mock":
			return NewMockProvider(), nil
		case "ollama":
			return NewOllamaProvider(baseURL, time.Duration(cfg.TimeoutSeconds)*time.Second), nil
		case "openai":
			return NewOpenAIProvider(baseURL, apiKey, time.Duration(cfg.TimeoutSeconds)*time.Second), nil
		case "gemini":
			return NewGeminiProvider(baseURL, apiKey, time.Duration(cfg.TimeoutSeconds)*time.Second), nil
		case "anthropic":
			return NewAnthropicProvider(baseURL, apiKey, time.Duration(cfg.TimeoutSeconds)*time.Second), nil
		case "":
			return nil, nil
		default:
			return nil, errors.New("unsupported ai provider: " + name)
		}
	}

	var err error
	service.provider, err = instantiateProvider(cfg.Provider, cfg.BaseURL, cfg.APIKey)
	if err != nil {
		return nil, err
	}

	if cfg.FallbackProvider != "" {
		service.fallbackProvider, err = instantiateProvider(cfg.FallbackProvider, cfg.FallbackBaseURL, cfg.FallbackAPIKey)
		if err != nil {
			return nil, err
		}
		service.fallbackModel = cfg.FallbackModel
		service.fallbackName = cfg.FallbackProvider
	}

	return service, nil
}

func (s *Service) Enabled() bool {
	return s != nil && s.enabled && s.provider != nil
}

func (s *Service) Generate(ctx context.Context, input GenerateInput) (*GenerateResult, error) {
	if s == nil || !s.enabled || s.provider == nil {
		return nil, ErrDisabled
	}
	
	primaryInput := input
	if primaryInput.Model == "" {
		primaryInput.Model = s.model
	}

	// Try primary provider
	res, err := s.provider.Generate(ctx, primaryInput)
	if err == nil {
		return res, nil
	}

	// If primary fails and fallback provider is configured, try the fallback
	if s.fallbackProvider != nil {
		fallbackInput := input
		if fallbackInput.Model == "" {
			fallbackInput.Model = s.fallbackModel
		}
		return s.fallbackProvider.Generate(ctx, fallbackInput)
	}

	return nil, err
}

func (s *Service) ProviderName() string {
	if s == nil {
		return ""
	}
	return s.providerName
}

func (s *Service) ModelName() string {
	if s == nil {
		return ""
	}
	return s.model
}
