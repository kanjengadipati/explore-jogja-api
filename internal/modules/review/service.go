package review

import "errors"

var ErrForbidden = errors.New("forbidden")

type Service struct {
	Repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{Repo: repo}
}

func (s *Service) GetAll() ([]Review, error) {
	return s.Repo.FindAll()
}

func (s *Service) GetAllAdmin() ([]Review, error) {
	return s.Repo.FindAllAdmin()
}

func (s *Service) GetByDestinationID(destinationID string) ([]Review, error) {
	return s.Repo.FindByDestinationID(destinationID)
}

func (s *Service) GetByUserID(userID string) ([]Review, error) {
	return s.Repo.FindByUserID(userID)
}

func (s *Service) GetByID(externalID string) (*Review, error) {
	return s.Repo.FindByID(externalID)
}

func (s *Service) Search(query string) ([]Review, error) {
	return s.Repo.Search(query)
}

type UpdateReviewRequest struct {
	UserName     *string  `json:"user_name"`
	TravelerType *string  `json:"traveler_type"`
	Rating       *int     `json:"rating"`
	Comment      *string  `json:"comment"`
	Images       *JSONArr `json:"images"`
	Status       *string  `json:"status"` // admin-only
}

func (s *Service) Create(review *Review) error {
	return s.Repo.Create(review)
}

func (s *Service) Update(externalID string, callerUserID string, isAdmin bool, req UpdateReviewRequest) (*Review, error) {
	review, err := s.Repo.FindByID(externalID)
	if err != nil {
		return nil, errors.New("review not found")
	}

	if !isAdmin && review.UserID != callerUserID {
		return nil, ErrForbidden
	}
	if req.Status != nil && !isAdmin {
		return nil, ErrForbidden
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

func (s *Service) Delete(externalID string, callerUserID string, isAdmin bool) error {
	review, err := s.Repo.FindByID(externalID)
	if err != nil {
		return errors.New("review not found")
	}
	if !isAdmin && review.UserID != callerUserID {
		return ErrForbidden
	}
	return s.Repo.Delete(externalID)
}

func (s *Service) GetByPartnerIDAndDestination(partnerExternalID string, destinationID string) ([]Review, error) {
	return s.Repo.FindByPartnerIDAndDestination(partnerExternalID, destinationID)
}

func (s *Service) GetByPartnerID(partnerExternalID string) ([]Review, error) {
	return s.Repo.FindByPartnerID(partnerExternalID)
}
