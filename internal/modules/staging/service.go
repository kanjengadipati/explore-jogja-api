package staging

import (
	"context"
	"encoding/json"
	"fmt"
	"pleco-api/internal/ai"
)

type Service struct {
	Repo      Repository
	AIService *ai.Service
}

func NewService(repo Repository, aiService *ai.Service) *Service {
	return &Service{Repo: repo, AIService: aiService}
}

type AIReviewResult struct {
	ID       uint   `json:"id"`
	Approved bool   `json:"approved"`
	Reason   string `json:"reason"`
}

func (s *Service) ReviewDestinations(ctx context.Context, ids []uint) ([]AIReviewResult, error) {
	dests, err := s.Repo.FindPendingDestinations()
	if err != nil {
		return nil, err
	}

	results := []AIReviewResult{}

	for _, dest := range dests {
		shouldProcess := false
		for _, id := range ids {
			if dest.ID == id {
				shouldProcess = true
				break
			}
		}
		if !shouldProcess {
			continue
		}

		prompt := fmt.Sprintf("Analyze this destination data for quality and appropriateness for a Yogyakarta tourism platform:\n%s\n\nReturn JSON only: {\"approved\": boolean, \"reason\": string}", dest.RawData)
		resp, err := s.AIService.Generate(ctx, ai.GenerateInput{
			SystemPrompt: "You are a data quality reviewer for tourism destinations. Analyze the provided data and return a JSON object with 'approved' (boolean) and 'reason' (string explaining your recommendation briefly).",
			UserPrompt:   prompt,
			Temperature:  0.3,
		})

		var parsed struct {
			Approved bool   `json:"approved"`
			Reason   string `json:"reason"`
		}
		if err == nil {
			_ = json.Unmarshal([]byte(resp.Text), &parsed)
		} else {
			parsed.Approved = false
			parsed.Reason = "AI unavailable"
		}

		results = append(results, AIReviewResult{
			ID:       dest.ID,
			Approved: parsed.Approved,
			Reason:   parsed.Reason,
		})
	}

	return results, nil
}
