package trips

import "errors"

type Service struct {
	Repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{Repo: repo}
}

func (s *Service) GetByUser(userID uint) ([]Trip, error) {
	return s.Repo.FindByUser(userID)
}

func (s *Service) GetByID(externalID string, userID uint) (*Trip, error) {
	trip, err := s.Repo.FindByID(externalID, userID)
	if err != nil {
		return nil, errors.New("trip not found")
	}
	return trip, nil
}

func (s *Service) Create(trip *Trip) error {
	return s.Repo.Create(trip)
}

// UpdateRequest uses pointer fields so only provided fields are updated.
type UpdateRequest struct {
	Title        *string   `json:"title"`
	StartDate    *string   `json:"start_date"`
	EndDate      *string   `json:"end_date"`
	DurationDays *int      `json:"duration_days"`
	Days         *JSONDays `json:"days"`
	Notes        *string   `json:"notes"`
	Status       *string   `json:"status"`
}

func (s *Service) Update(externalID string, userID uint, req UpdateRequest) (*Trip, error) {
	trip, err := s.Repo.FindByID(externalID, userID)
	if err != nil {
		return nil, errors.New("trip not found")
	}

	if req.Title != nil {
		trip.Title = *req.Title
	}
	if req.StartDate != nil {
		trip.StartDate = *req.StartDate
	}
	if req.EndDate != nil {
		trip.EndDate = *req.EndDate
	}
	if req.DurationDays != nil {
		trip.DurationDays = *req.DurationDays
	}
	if req.Days != nil {
		trip.Days = *req.Days
	}
	if req.Notes != nil {
		trip.Notes = *req.Notes
	}
	if req.Status != nil {
		trip.Status = *req.Status
	}

	if err := s.Repo.Update(trip); err != nil {
		return nil, err
	}
	return trip, nil
}

func (s *Service) Delete(externalID string, userID uint) error {
	return s.Repo.Delete(externalID, userID)
}
