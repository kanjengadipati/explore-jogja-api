package hotel

import "errors"

type Service struct {
	Repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{Repo: repo}
}

func (s *Service) GetAll() ([]Hotel, error) {
	return s.Repo.FindAll()
}

func (s *Service) GetByID(externalID string) (*Hotel, error) {
	return s.Repo.FindByID(externalID)
}

func (s *Service) Search(query string) ([]Hotel, error) {
	return s.Repo.Search(query)
}

type UpdateHotelRequest struct {
	Name          *string  `json:"name"`
	Description   *string  `json:"description"`
	Location      *string  `json:"location"`
	Address       *string  `json:"address"`
	Stars         *int     `json:"stars"`
	PricePerNight *string  `json:"price_per_night"`
	Images        *JSONArr `json:"images"`
	Amenities     *JSONArr `json:"amenities"`
	Phone         *string  `json:"phone"`
	Email         *string  `json:"email"`
	Website       *string  `json:"website"`
	Rating        *float64 `json:"rating"`
	ReviewCount   *int     `json:"review_count"`
	Latitude      *float64 `json:"latitude"`
	Longitude     *float64 `json:"longitude"`
}

func (s *Service) Create(hotel *Hotel) error {
	return s.Repo.Create(hotel)
}

func (s *Service) Update(externalID string, req UpdateHotelRequest) (*Hotel, error) {
	hotel, err := s.Repo.FindByID(externalID)
	if err != nil {
		return nil, errors.New("hotel not found")
	}

	if req.Name != nil {
		hotel.Name = *req.Name
	}
	if req.Description != nil {
		hotel.Description = *req.Description
	}
	if req.Location != nil {
		hotel.Location = *req.Location
	}
	if req.Address != nil {
		hotel.Address = *req.Address
	}
	if req.Stars != nil {
		hotel.Stars = *req.Stars
	}
	if req.PricePerNight != nil {
		hotel.PricePerNight = *req.PricePerNight
	}
	if req.Images != nil {
		hotel.Images = *req.Images
	}
	if req.Amenities != nil {
		hotel.Amenities = *req.Amenities
	}
	if req.Phone != nil {
		hotel.Phone = *req.Phone
	}
	if req.Email != nil {
		hotel.Email = *req.Email
	}
	if req.Website != nil {
		hotel.Website = *req.Website
	}
	if req.Rating != nil {
		hotel.Rating = *req.Rating
	}
	if req.ReviewCount != nil {
		hotel.ReviewCount = *req.ReviewCount
	}
	if req.Latitude != nil {
		hotel.Latitude = *req.Latitude
	}
	if req.Longitude != nil {
		hotel.Longitude = *req.Longitude
	}

	if err := s.Repo.Update(hotel); err != nil {
		return nil, err
	}
	return hotel, nil
}

func (s *Service) Delete(externalID string) error {
	return s.Repo.Delete(externalID)
}
