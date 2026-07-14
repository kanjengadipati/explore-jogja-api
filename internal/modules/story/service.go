package story

import "errors"

type Service struct {
	Repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{Repo: repo}
}

func (s *Service) GetAll() ([]Story, error) {
	return s.Repo.FindAll()
}

func (s *Service) GetByID(externalID string) (*Story, error) {
	return s.Repo.FindByID(externalID)
}

func (s *Service) Search(query string) ([]Story, error) {
	return s.Repo.Search(query)
}

type UpdateStoryRequest struct {
	UserID         *string  `json:"user_id"`
	Title          *string  `json:"title"`
	Content        *string  `json:"content"`
	Images         *JSONArr `json:"images"`
	DestinationIDs *JSONArr `json:"destination_ids"`
	Likes          *int     `json:"likes"`
	Status         *string  `json:"status"`
}

func (s *Service) Create(story *Story) error {
	return s.Repo.Create(story)
}

func (s *Service) Update(externalID string, req UpdateStoryRequest) (*Story, error) {
	story, err := s.Repo.FindByID(externalID)
	if err != nil {
		return nil, errors.New("story not found")
	}

	if req.UserID != nil {
		story.UserID = *req.UserID
	}
	if req.Title != nil {
		story.Title = *req.Title
	}
	if req.Content != nil {
		story.Content = *req.Content
	}
	if req.Images != nil {
		story.Images = *req.Images
	}
	if req.DestinationIDs != nil {
		story.DestinationIDs = *req.DestinationIDs
	}
	if req.Likes != nil {
		story.Likes = *req.Likes
	}
	if req.Status != nil {
		story.Status = *req.Status
	}

	if err := s.Repo.Update(story); err != nil {
		return nil, err
	}
	return story, nil
}

func (s *Service) Delete(externalID string) error {
	return s.Repo.Delete(externalID)
}
