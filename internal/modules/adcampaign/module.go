package adcampaign

import (
	"gorm.io/gorm"

	"pleco-api/internal/modules/subscription"
	"pleco-api/internal/services"
)

type Module struct {
	Repository Repository
	Service    *Service
	Handler    *Handler
}

func BuildModule(db *gorm.DB, subSvc *subscription.Service, emailSvc services.EmailService) *Module {
	repository := NewRepository(db)
	service := NewService(repository, subSvc, emailSvc)
	handler := NewHandler(service)

	return &Module{
		Repository: repository,
		Service:    service,
		Handler:    handler,
	}
}
