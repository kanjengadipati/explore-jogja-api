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
	Approved bool   `json:"approved"`
	Reason   string `json:"reason"`
}

func (s *Service) ReviewDestinations(ctx context.Context, ids []uint) error {
	dests, err := s.Repo.FindPendingDestinations()
	if err != nil {
		return err
	}

	approvedIDs := []uint{}
	rejectedIDs := []uint{}

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

		// Perform AI review
		prompt := fmt.Sprintf("Analyze this destination data for quality and appropriateness:\n%s\n\nReturn JSON: {\"approved\": boolean, \"reason\": string}", dest.RawData)
		resp, err := s.AIService.Generate(ctx, ai.GenerateInput{
			SystemPrompt: "You are a data quality reviewer for tourism destinations. Analyze the provided data and return a JSON object with 'approved' (boolean) and 'reason' (string).",
			UserPrompt:   prompt,
			Temperature: 0.3,
		})
		
		var result AIReviewResult
		if err == nil {
			err = json.Unmarshal([]byte(resp.Text), &result)
		}

		if err == nil && result.Approved {
			approvedIDs = append(approvedIDs, dest.ID)
		} else {
			rejectedIDs = append(rejectedIDs, dest.ID)
		}
	}

	if len(approvedIDs) > 0 {
		s.Repo.ApproveMultipleDestinations(approvedIDs)
	}
	if len(rejectedIDs) > 0 {
		s.Repo.RejectMultipleDestinations(rejectedIDs)
	}

	return nil
}
