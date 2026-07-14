package souvenir

import "errors"

type Service struct {
	Repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{Repo: repo}
}

func (s *Service) GetAll() ([]Souvenir, error) {
	return s.Repo.FindAll()
}

func (s *Service) GetByID(externalID string) (*Souvenir, error) {
	return s.Repo.FindByID(externalID)
}

func (s *Service) Search(query string) ([]Souvenir, error) {
	return s.Repo.Search(query)
}

type UpdateSouvenirRequest struct {
	Name         *string  `json:"name"`
	Description  *string  `json:"description"`
	Location     *string  `json:"location"`
	Address      *string  `json:"address"`
	Images       *JSONArr `json:"images"`
	ProductTypes *JSONArr `json:"product_types"`
	PriceRange   *string  `json:"price_range"`
	Phone        *string  `json:"phone"`
	Rating       *float64 `json:"rating"`
	Latitude     *float64 `json:"latitude"`
	Longitude    *float64 `json:"longitude"`
}

func (s *Service) Create(souvenir *Souvenir) error {
	return s.Repo.Create(souvenir)
}

func (s *Service) Update(externalID string, req UpdateSouvenirRequest) (*Souvenir, error) {
	souvenir, err := s.Repo.FindByID(externalID)
	if err != nil {
		return nil, errors.New("souvenir not found")
	}

	if req.Name != nil {
		souvenir.Name = *req.Name
	}
	if req.Description != nil {
		souvenir.Description = *req.Description
	}
	if req.Location != nil {
		souvenir.Location = *req.Location
	}
	if req.Address != nil {
		souvenir.Address = *req.Address
	}
	if req.Images != nil {
		souvenir.Images = *req.Images
	}
	if req.ProductTypes != nil {
		souvenir.ProductTypes = *req.ProductTypes
	}
	if req.PriceRange != nil {
		souvenir.PriceRange = *req.PriceRange
	}
	if req.Phone != nil {
		souvenir.Phone = *req.Phone
	}
	if req.Rating != nil {
		souvenir.Rating = *req.Rating
	}
	if req.Latitude != nil {
		souvenir.Latitude = *req.Latitude
	}
	if req.Longitude != nil {
		souvenir.Longitude = *req.Longitude
	}

	if err := s.Repo.Update(souvenir); err != nil {
		return nil, err
	}
	return souvenir, nil
}

func (s *Service) Delete(externalID string) error {
	return s.Repo.Delete(externalID)
}
