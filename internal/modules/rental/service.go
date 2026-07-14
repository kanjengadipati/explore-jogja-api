package rental

import "errors"

type Service struct {
	Repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{Repo: repo}
}

func (s *Service) GetAll() ([]Rental, error) {
	return s.Repo.FindAll()
}

func (s *Service) GetByID(externalID string) (*Rental, error) {
	return s.Repo.FindByID(externalID)
}

func (s *Service) Search(query string) ([]Rental, error) {
	return s.Repo.Search(query)
}

type UpdateRentalRequest struct {
	Name         *string  `json:"name"`
	Description  *string  `json:"description"`
	Location     *string  `json:"location"`
	Address      *string  `json:"address"`
	VehicleTypes *JSONArr `json:"vehicle_types"`
	PricePerDay  *string  `json:"price_per_day"`
	Images       *JSONArr `json:"images"`
	Phone        *string  `json:"phone"`
	Rating       *float64 `json:"rating"`
	Latitude     *float64 `json:"latitude"`
	Longitude    *float64 `json:"longitude"`
}

func (s *Service) Create(rental *Rental) error {
	return s.Repo.Create(rental)
}

func (s *Service) Update(externalID string, req UpdateRentalRequest) (*Rental, error) {
	rental, err := s.Repo.FindByID(externalID)
	if err != nil {
		return nil, errors.New("rental not found")
	}

	if req.Name != nil {
		rental.Name = *req.Name
	}
	if req.Description != nil {
		rental.Description = *req.Description
	}
	if req.Location != nil {
		rental.Location = *req.Location
	}
	if req.Address != nil {
		rental.Address = *req.Address
	}
	if req.VehicleTypes != nil {
		rental.VehicleTypes = *req.VehicleTypes
	}
	if req.PricePerDay != nil {
		rental.PricePerDay = *req.PricePerDay
	}
	if req.Images != nil {
		rental.Images = *req.Images
	}
	if req.Phone != nil {
		rental.Phone = *req.Phone
	}
	if req.Rating != nil {
		rental.Rating = *req.Rating
	}
	if req.Latitude != nil {
		rental.Latitude = *req.Latitude
	}
	if req.Longitude != nil {
		rental.Longitude = *req.Longitude
	}

	if err := s.Repo.Update(rental); err != nil {
		return nil, err
	}
	return rental, nil
}

func (s *Service) Delete(externalID string) error {
	return s.Repo.Delete(externalID)
}
