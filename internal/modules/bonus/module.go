package bonus

import (
	"pleco-api/internal/modules/config"
	"pleco-api/internal/modules/user"

	"gorm.io/gorm"
)

type Module struct {
	Repository Repository
	Service    *Service
	Handler    *Handler
}

func BuildModule(db *gorm.DB, configSvc *config.Service, userSvc *user.Service) *Module {
	repo := NewRepository(db)
	svc := NewService(repo, configSvc, userSvc)
	handler := NewHandler(svc)

	return &Module{
		Repository: repo,
		Service:    svc,
		Handler:    handler,
	}
}
