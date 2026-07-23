package imagereport

type Service struct {
	Repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{Repo: repo}
}

func (s *Service) CreateReport(report *ImageReport) error {
	report.Status = "pending"
	return s.Repo.Create(report)
}

func (s *Service) GetAll() ([]ImageReport, error) {
	return s.Repo.FindAll()
}

func (s *Service) GetByStatus(status string) ([]ImageReport, error) {
	return s.Repo.FindByStatus(status)
}

func (s *Service) Resolve(id uint) error {
	return s.Repo.UpdateStatus(id, "resolved")
}

func (s *Service) Dismiss(id uint) error {
	return s.Repo.UpdateStatus(id, "dismissed")
}

func (s *Service) GetStats() (pending int64, resolved int64, dismissed int64, err error) {
	pending, err = s.Repo.CountByStatus("pending")
	if err != nil {
		return
	}
	resolved, err = s.Repo.CountByStatus("resolved")
	if err != nil {
		return
	}
	dismissed, err = s.Repo.CountByStatus("dismissed")
	return
}
