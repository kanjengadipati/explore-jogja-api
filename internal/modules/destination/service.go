package destination

type Service struct {
	Repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{Repo: repo}
}

func (s *Service) GetAll() ([]Destination, error) {
	return s.Repo.FindAll()
}

func (s *Service) GetByID(externalID string) (*Destination, error) {
	return s.Repo.FindByID(externalID)
}

func (s *Service) GetByCategory(category string) ([]Destination, error) {
	return s.Repo.FindByCategory(category)
}

func (s *Service) Search(query string) ([]Destination, error) {
	return s.Repo.Search(query)
}
