package restaurant

import "errors"

type Service struct {
	Repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{Repo: repo}
}

func (s *Service) GetAll() ([]Restaurant, error) {
	return s.Repo.FindAll()
}

func (s *Service) GetByID(externalID string) (*Restaurant, error) {
	return s.Repo.FindByID(externalID)
}

func (s *Service) Search(query string) ([]Restaurant, error) {
	return s.Repo.Search(query)
}

type UpdateRestaurantRequest struct {
	Name         *string  `json:"name"`
	Description  *string  `json:"description"`
	Location     *string  `json:"location"`
	Address      *string  `json:"address"`
	CuisineType  *string  `json:"cuisine_type"`
	PriceRange   *string  `json:"price_range"`
	Images       *JSONArr `json:"images"`
	OpeningHours *string  `json:"opening_hours"`
	Phone        *string  `json:"phone"`
	Rating       *float64 `json:"rating"`
	ReviewCount  *int     `json:"review_count"`
	Latitude     *float64 `json:"latitude"`
	Longitude    *float64 `json:"longitude"`
}

func (s *Service) Create(restaurant *Restaurant) error {
	return s.Repo.Create(restaurant)
}

func (s *Service) Update(externalID string, req UpdateRestaurantRequest) (*Restaurant, error) {
	restaurant, err := s.Repo.FindByID(externalID)
	if err != nil {
		return nil, errors.New("restaurant not found")
	}

	if req.Name != nil {
		restaurant.Name = *req.Name
	}
	if req.Description != nil {
		restaurant.Description = *req.Description
	}
	if req.Location != nil {
		restaurant.Location = *req.Location
	}
	if req.Address != nil {
		restaurant.Address = *req.Address
	}
	if req.CuisineType != nil {
		restaurant.CuisineType = *req.CuisineType
	}
	if req.PriceRange != nil {
		restaurant.PriceRange = *req.PriceRange
	}
	if req.Images != nil {
		restaurant.Images = *req.Images
	}
	if req.OpeningHours != nil {
		restaurant.OpeningHours = *req.OpeningHours
	}
	if req.Phone != nil {
		restaurant.Phone = *req.Phone
	}
	if req.Rating != nil {
		restaurant.Rating = *req.Rating
	}
	if req.ReviewCount != nil {
		restaurant.ReviewCount = *req.ReviewCount
	}
	if req.Latitude != nil {
		restaurant.Latitude = *req.Latitude
	}
	if req.Longitude != nil {
		restaurant.Longitude = *req.Longitude
	}

	if err := s.Repo.Update(restaurant); err != nil {
		return nil, err
	}
	return restaurant, nil
}

func (s *Service) Delete(externalID string) error {
	return s.Repo.Delete(externalID)
}
