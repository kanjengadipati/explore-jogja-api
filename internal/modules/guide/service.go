package guide

import "errors"

type Service struct {
	Repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{Repo: repo}
}

func (s *Service) GetAll() ([]Guide, error) {
	return s.Repo.FindAll()
}

func (s *Service) GetByID(externalID string) (*Guide, error) {
	return s.Repo.FindByID(externalID)
}

func (s *Service) Search(query string) ([]Guide, error) {
	return s.Repo.Search(query)
}

type UpdateGuideRequest struct {
	Name           *string  `json:"name"`
	Bio            *string  `json:"bio"`
	Specialization *string  `json:"specialization"`
	Phone          *string  `json:"phone"`
	Email          *string  `json:"email"`
	Rating         *float64 `json:"rating"`
	ReviewCount    *int     `json:"review_count"`
	Languages      *JSONArr `json:"languages"`
	PricePerDay    *string  `json:"price_per_day"`
	Avatar         *string  `json:"avatar"`
}

func (s *Service) Create(guide *Guide) error {
	return s.Repo.Create(guide)
}

func (s *Service) Update(externalID string, req UpdateGuideRequest) (*Guide, error) {
	guide, err := s.Repo.FindByID(externalID)
	if err != nil {
		return nil, errors.New("guide not found")
	}

	if req.Name != nil {
		guide.Name = *req.Name
	}
	if req.Bio != nil {
		guide.Bio = *req.Bio
	}
	if req.Specialization != nil {
		guide.Specialization = *req.Specialization
	}
	if req.Phone != nil {
		guide.Phone = *req.Phone
	}
	if req.Email != nil {
		guide.Email = *req.Email
	}
	if req.Rating != nil {
		guide.Rating = *req.Rating
	}
	if req.ReviewCount != nil {
		guide.ReviewCount = *req.ReviewCount
	}
	if req.Languages != nil {
		guide.Languages = *req.Languages
	}
	if req.PricePerDay != nil {
		guide.PricePerDay = *req.PricePerDay
	}
	if req.Avatar != nil {
		guide.Avatar = *req.Avatar
	}

	if err := s.Repo.Update(guide); err != nil {
		return nil, err
	}
	return guide, nil
}

func (s *Service) Delete(externalID string) error {
	return s.Repo.Delete(externalID)
}
