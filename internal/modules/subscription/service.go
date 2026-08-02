package subscription

import (
	"errors"
	"time"
)

type Service struct {
	Repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{Repo: repo}
}

func (s *Service) GetByExternalID(externalID string) (*Subscription, error) {
	return s.Repo.FindByExternalID(externalID)
}

func (s *Service) GetByBusinessID(businessID uint) (*Subscription, error) {
	return s.Repo.FindByBusinessID(businessID)
}

func (s *Service) GetByBusinessExternalID(businessExternalID string) (*Subscription, error) {
	return s.Repo.FindByBusinessExternalID(businessExternalID)
}

func (s *Service) CanCreateAdCampaign(businessExternalID string) (bool, error) {
	sub, err := s.Repo.FindByBusinessExternalID(businessExternalID)
	if err != nil {
		return false, err
	}
	return sub.CanCreateAdCampaign(), nil
}

func (s *Service) Upgrade(externalID string, amount float64) (*Subscription, error) {
	sub, err := s.Repo.FindByExternalID(externalID)
	if err != nil {
		return nil, err
	}

	plan := PlanFree
	if amount >= 1000000 {
		plan = PlanEnterprise
	} else if amount >= 300000 {
		plan = PlanBusinessPlus
	} else if amount >= 150000 {
		plan = PlanPro
	} else {
		return nil, errors.New("insufficient amount for subscription upgrade")
	}

	sub.Plan = plan
	sub.Status = StatusActive
	end := time.Now().AddDate(0, 0, 30)
	sub.CurrentPeriodEnd = &end

	if err := s.Repo.Update(sub); err != nil {
		return nil, err
	}
	return sub, nil
}
