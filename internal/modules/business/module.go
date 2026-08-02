package business

import (
	"pleco-api/internal/modules/audit"
	"pleco-api/internal/modules/notification"
	"pleco-api/internal/modules/promotion"
	"pleco-api/internal/modules/review"

	"gorm.io/gorm"
)

type Module struct {
	Repository Repository
	Service    *Service
	Handler    *Handler
}

func BuildModule(db *gorm.DB, promoSvc *promotion.Service, reviewSvc *review.Service, auditSvc *audit.Service, notifSvc *notification.Service, partnerSync PartnerMirrorSyncer) *Module {
	repo := NewRepository(db)
	service := NewService(repo)
	service.PartnerSync = partnerSync
	handler := NewHandler(service, promoSvc, reviewSvc, auditSvc, notifSvc)
	return &Module{
		Repository: repo,
		Service:    service,
		Handler:    handler,
	}
}
