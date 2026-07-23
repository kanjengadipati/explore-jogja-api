package event

import "errors"

type Service struct {
	Repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{Repo: repo}
}

func (s *Service) GetAll() ([]Event, error) {
	return s.Repo.FindAll()
}

func (s *Service) GetByID(externalID string) (*Event, error) {
	return s.Repo.FindByID(externalID)
}

func (s *Service) Search(query string) ([]Event, error) {
	return s.Repo.Search(query)
}

type UpdateEventRequest struct {
	Title         *string  `json:"title"`
	Description   *string  `json:"description"`
	Location      *string  `json:"location"`
	StartDate     *string  `json:"start_date"`
	EndDate       *string  `json:"end_date"`
	ImageURL      *string  `json:"image_url"`
	Images        *JSONArr `json:"images"`
	Category      *string  `json:"category"`
	Status        *string  `json:"status"`
	Latitude      *float64 `json:"latitude"`
	Longitude     *float64 `json:"longitude"`
	MaxAttendees  *int     `json:"max_attendees"`
	TicketPrice   *string  `json:"ticket_price"`
	Organizer     *string  `json:"organizer"`
	DestinationID *string  `json:"destination_id"`
	VideoURL      *string  `json:"video_url"`
}

func (s *Service) Create(event *Event) error {
	return s.Repo.Create(event)
}

func (s *Service) Update(externalID string, req UpdateEventRequest) (*Event, error) {
	event, err := s.Repo.FindByID(externalID)
	if err != nil {
		return nil, errors.New("event not found")
	}

	if req.Title != nil {
		event.Title = *req.Title
	}
	if req.Description != nil {
		event.Description = *req.Description
	}
	if req.Location != nil {
		event.Location = *req.Location
	}
	if req.StartDate != nil {
		event.StartDate = *req.StartDate
	}
	if req.EndDate != nil {
		event.EndDate = *req.EndDate
	}
	if req.ImageURL != nil {
		event.ImageURL = *req.ImageURL
	}
	if req.Images != nil {
		event.Images = *req.Images
		// Keep image_url in sync with the first image for backwards compatibility
		if len(*req.Images) > 0 {
			if first, ok := (*req.Images)[0].(map[string]interface{}); ok {
				if url, ok := first["url"].(string); ok && url != "" {
					event.ImageURL = url
				}
			} else if url, ok := (*req.Images)[0].(string); ok && url != "" {
				event.ImageURL = url
			}
		}
	}
	if req.VideoURL != nil {
		event.VideoURL = *req.VideoURL
	}
	if req.Category != nil {
		event.Category = *req.Category
	}
	if req.Status != nil {
		event.Status = *req.Status
	}
	if req.Latitude != nil {
		event.Latitude = *req.Latitude
	}
	if req.Longitude != nil {
		event.Longitude = *req.Longitude
	}
	if req.MaxAttendees != nil {
		event.MaxAttendees = *req.MaxAttendees
	}
	if req.TicketPrice != nil {
		event.TicketPrice = *req.TicketPrice
	}
	if req.Organizer != nil {
		event.Organizer = *req.Organizer
	}
	if req.DestinationID != nil {
		event.DestinationID = *req.DestinationID
	}

	if err := s.Repo.Update(event); err != nil {
		return nil, err
	}
	return event, nil
}

func (s *Service) Delete(externalID string) error {
	return s.Repo.Delete(externalID)
}
