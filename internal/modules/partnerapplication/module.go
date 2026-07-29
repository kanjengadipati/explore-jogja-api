package partnerapplication

import (
	"gorm.io/gorm"

	"pleco-api/internal/modules/audit"
	"pleco-api/internal/modules/notification"
	"pleco-api/internal/modules/partner"
	"pleco-api/internal/modules/user"
)

type Module struct {
	Repository Repository
	Service    *Service
	Handler    *Handler
}

func BuildModule(db *gorm.DB, partnerSvc *partner.Service, userSvc *user.Service, auditSvc *audit.Service, notifSvc *notification.Service) *Module {
	repository := NewRepository(db)
	service := NewService(repository, partnerSvc, userSvc, auditSvc, notifSvc)
	handler := NewHandler(service)
	return &Module{Repository: repository, Service: service, Handler: handler}
}
