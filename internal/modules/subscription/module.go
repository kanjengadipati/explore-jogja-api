package subscription

import (
	"gorm.io/gorm"
)

type Module struct {
	Repo    Repository
	Service *Service
}

func BuildModule(db *gorm.DB) *Module {
	repo := NewRepository(db)
	svc := NewService(repo)
	return &Module{
		Repo:    repo,
		Service: svc,
	}
}
