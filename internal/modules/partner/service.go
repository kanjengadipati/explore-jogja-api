package partner

import "errors"

type Service struct {
	Repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{Repo: repo}
}

func (s *Service) GetAll() ([]Partner, error) {
	return s.Repo.FindAll()
}

func (s *Service) GetByID(externalID string) (*Partner, error) {
	return s.Repo.FindByID(externalID)
}

func (s *Service) Search(query string) ([]Partner, error) {
	return s.Repo.Search(query)
}

type UpdatePartnerRequest struct {
	Name        *string  `json:"name"`
	Description *string  `json:"description"`
	Category    *string  `json:"category"`
	Location    *string  `json:"location"`
	Address     *string  `json:"address"`
	Image       *string  `json:"image"`
	Rating      *float64 `json:"rating"`
	Price       *string  `json:"price"`
	Distance    *string  `json:"distance"`
	Phone       *string  `json:"phone"`
	Website     *string  `json:"website"`
	Latitude    *float64 `json:"latitude"`
	Longitude   *float64 `json:"longitude"`
}

func (s *Service) Create(partner *Partner) error {
	return s.Repo.Create(partner)
}

func (s *Service) Update(externalID string, req UpdatePartnerRequest) (*Partner, error) {
	partner, err := s.Repo.FindByID(externalID)
	if err != nil {
		return nil, errors.New("partner not found")
	}

	if req.Name != nil {
		partner.Name = *req.Name
	}
	if req.Description != nil {
		partner.Description = *req.Description
	}
	if req.Category != nil {
		partner.Category = *req.Category
	}
	if req.Location != nil {
		partner.Location = *req.Location
	}
	if req.Address != nil {
		partner.Address = *req.Address
	}
	if req.Image != nil {
		partner.Image = *req.Image
	}
	if req.Rating != nil {
		partner.Rating = *req.Rating
	}
	if req.Price != nil {
		partner.Price = *req.Price
	}
	if req.Distance != nil {
		partner.Distance = *req.Distance
	}
	if req.Phone != nil {
		partner.Phone = *req.Phone
	}
	if req.Website != nil {
		partner.Website = *req.Website
	}
	if req.Latitude != nil {
		partner.Latitude = *req.Latitude
	}
	if req.Longitude != nil {
		partner.Longitude = *req.Longitude
	}

	if err := s.Repo.Update(partner); err != nil {
		return nil, err
	}
	return partner, nil
}

func (s *Service) Delete(externalID string) error {
	return s.Repo.Delete(externalID)
}
