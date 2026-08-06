package business

import (
	"pleco-api/internal/modules/adcampaign"
	"pleco-api/internal/modules/audit"
	"pleco-api/internal/modules/notification"
	"pleco-api/internal/modules/promotion"
	"pleco-api/internal/modules/review"
	"pleco-api/internal/modules/subscription"
	"pleco-api/internal/modules/user"

	"gorm.io/gorm"
)

type Module struct {
	Repository Repository
	Service    *Service
	Handler    *Handler
}

func BuildModule(db *gorm.DB, promoSvc *promotion.Service, reviewSvc *review.Service, auditSvc *audit.Service, notifSvc *notification.Service, partnerSync PartnerMirrorSyncer, subSvc *subscription.Service, adCampaignSvc *adcampaign.Service, userSvc *user.Service) *Module {
	db.AutoMigrate(&BusinessOwner{})
	repo := NewRepository(db)
	service := NewService(repo)
	service.PartnerSync = partnerSync
	service.UserSvc = userSvc
	handler := NewHandler(service, promoSvc, reviewSvc, auditSvc, notifSvc, subSvc, adCampaignSvc, userSvc.UserRepo)
	return &Module{
		Repository: repo,
		Service:    service,
		Handler:    handler,
	}
}
