package promotion

import "errors"

type Service struct {
	Repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{Repo: repo}
}

func (s *Service) GetAll() ([]Promotion, error) {
	return s.Repo.FindAll()
}

func (s *Service) GetByID(externalID string) (*Promotion, error) {
	return s.Repo.FindByID(externalID)
}

func (s *Service) Search(query string) ([]Promotion, error) {
	return s.Repo.Search(query)
}

type UpdatePromotionRequest struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	Discount    *string `json:"discount"`
	StartDate   *string `json:"start_date"`
	EndDate     *string `json:"end_date"`
	ImageURL    *string `json:"image_url"`
	Category    *string `json:"category"`
	Status      *string `json:"status"`
	Code        *string `json:"code"`
}

func (s *Service) Create(promotion *Promotion) error {
	return s.Repo.Create(promotion)
}

func (s *Service) Update(externalID string, req UpdatePromotionRequest) (*Promotion, error) {
	promotion, err := s.Repo.FindByID(externalID)
	if err != nil {
		return nil, errors.New("promotion not found")
	}

	if req.Title != nil {
		promotion.Title = *req.Title
	}
	if req.Description != nil {
		promotion.Description = *req.Description
	}
	if req.Discount != nil {
		promotion.Discount = *req.Discount
	}
	if req.StartDate != nil {
		promotion.StartDate = *req.StartDate
	}
	if req.EndDate != nil {
		promotion.EndDate = *req.EndDate
	}
	if req.ImageURL != nil {
		promotion.ImageURL = *req.ImageURL
	}
	if req.Category != nil {
		promotion.Category = *req.Category
	}
	if req.Status != nil {
		promotion.Status = *req.Status
	}
	if req.Code != nil {
		promotion.Code = *req.Code
	}

	if err := s.Repo.Update(promotion); err != nil {
		return nil, err
	}
	return promotion, nil
}

func (s *Service) Delete(externalID string) error {
	return s.Repo.Delete(externalID)
}
