package business

import (
	"pleco-api/internal/modules/adcampaign"
	"pleco-api/internal/modules/audit"
	"pleco-api/internal/modules/notification"
	"pleco-api/internal/modules/promotion"
	"pleco-api/internal/modules/review"
	"pleco-api/internal/modules/subscription"
	"pleco-api/internal/modules/user"
	"pleco-api/internal/services"

	"gorm.io/gorm"
)

type Module struct {
	Repository Repository
	Service    *Service
	Handler    *Handler
}

func BuildModule(db *gorm.DB, promoSvc *promotion.Service, reviewSvc *review.Service, auditSvc *audit.Service, notifSvc *notification.Service, subSvc *subscription.Service, adCampaignSvc *adcampaign.Service, userSvc *user.Service, emailSvc services.EmailService) *Module {
	db.AutoMigrate(&BusinessOwner{}, &BusinessMemberInvite{})
	repo := NewRepository(db)
	service := NewService(repo)
	service.UserSvc = userSvc
	handler := NewHandler(service, promoSvc, reviewSvc, auditSvc, notifSvc, subSvc, adCampaignSvc, userSvc.UserRepo, emailSvc)
	return &Module{
		Repository: repo,
		Service:    service,
		Handler:    handler,
	}
}
