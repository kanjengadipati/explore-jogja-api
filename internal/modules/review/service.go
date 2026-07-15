package review

import "errors"

type Service struct {
	Repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{Repo: repo}
}

func (s *Service) GetAll() ([]Review, error) {
	return s.Repo.FindAll()
}

func (s *Service) GetByID(externalID string) (*Review, error) {
	return s.Repo.FindByID(externalID)
}

func (s *Service) Search(query string) ([]Review, error) {
	return s.Repo.Search(query)
}

type UpdateReviewRequest struct {
	UserID        *string  `json:"user_id"`
	DestinationID *string  `json:"destination_id"`
	UserName      *string  `json:"user_name"`
	TravelerType  *string  `json:"traveler_type"`
	Rating        *int     `json:"rating"`
	Comment       *string  `json:"comment"`
	Images        *JSONArr `json:"images"`
	Status        *string  `json:"status"`
}

func (s *Service) Create(review *Review) error {
	return s.Repo.Create(review)
}

func (s *Service) Update(externalID string, req UpdateReviewRequest) (*Review, error) {
	review, err := s.Repo.FindByID(externalID)
	if err != nil {
		return nil, errors.New("review not found")
	}

	if req.UserID != nil {
		review.UserID = *req.UserID
	}
	if req.DestinationID != nil {
		review.DestinationID = *req.DestinationID
	}
	if req.UserName != nil {
		review.UserName = *req.UserName
	}
	if req.TravelerType != nil {
		review.TravelerType = *req.TravelerType
	}
	if req.Rating != nil {
		review.Rating = *req.Rating
	}
	if req.Comment != nil {
		review.Comment = *req.Comment
	}
	if req.Images != nil {
		review.Images = *req.Images
	}
	if req.Status != nil {
		review.Status = *req.Status
	}

	if err := s.Repo.Update(review); err != nil {
		return nil, err
	}
	return review, nil
}

func (s *Service) Delete(externalID string) error {
	return s.Repo.Delete(externalID)
}
