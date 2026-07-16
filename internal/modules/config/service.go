package config

type Service struct {
	Repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{Repo: repo}
}

func (s *Service) GetAll() ([]SiteConfig, error) {
	return s.Repo.GetAll()
}

func (s *Service) GetByCategory(category string) ([]SiteConfig, error) {
	return s.Repo.GetByCategory(category)
}

func (s *Service) GetByKey(key string) (*SiteConfig, error) {
	return s.Repo.GetByKey(key)
}

type UpdateSiteConfigRequest struct {
	Key      string `json:"key" binding:"required"`
	Value    string `json:"value"`
	Category string `json:"category"`
}

func (s *Service) Update(req UpdateSiteConfigRequest) (*SiteConfig, error) {
	if err := s.Repo.Upsert(req.Key, req.Value, req.Category); err != nil {
		return nil, err
	}
	return s.Repo.GetByKey(req.Key)
}

type BulkUpdateSiteConfigRequest struct {
	Configs []UpdateSiteConfigRequest `json:"configs" binding:"required"`
}

func (s *Service) BulkUpdate(req BulkUpdateSiteConfigRequest) ([]SiteConfig, error) {
	for _, c := range req.Configs {
		_ = s.Repo.Upsert(c.Key, c.Value, c.Category)
	}
	return s.Repo.GetAll()
}

func (s *Service) GetMap() map[string]string {
	configs, _ := s.Repo.GetAll()
	result := make(map[string]string)
	for _, c := range configs {
		result[c.Key] = c.Value
	}
	return result
}
