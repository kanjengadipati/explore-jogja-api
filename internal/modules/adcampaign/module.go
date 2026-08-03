package adcampaign

import (
	"gorm.io/gorm"

	"pleco-api/internal/modules/subscription"
)

type Module struct {
	Repository Repository
	Service    *Service
	Handler    *Handler
}

func BuildModule(db *gorm.DB, subSvc *subscription.Service) *Module {
	repository := NewRepository(db)
	service := NewService(repository, subSvc)
	handler := NewHandler(service)

	return &Module{
		Repository: repository,
		Service:    service,
		Handler:    handler,
	}
}
