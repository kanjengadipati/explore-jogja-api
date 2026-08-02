package payment

import (
	"gorm.io/gorm"

	"pleco-api/internal/modules/adcampaign"
	"pleco-api/internal/modules/audit"
	"pleco-api/internal/modules/partner"
	"pleco-api/internal/modules/subscription"
	midtransprovider "pleco-api/internal/providers/payment/midtrans"
	"pleco-api/internal/services"
)

type Module struct {
	Repository Repository
	Service    *Service
	Handler    *Handler
}

func BuildModule(
	db *gorm.DB,
	midtransClient *midtransprovider.Client,
	partnerSvc *partner.Service,
	adCampaignSvc *adcampaign.Service,
	subscriptionSvc *subscription.Service,
	auditSvc *audit.Service,
	emailSvc services.EmailService,
) *Module {
	repository := NewRepository(db)
	adapter := midtransprovider.NewAdapter(midtransClient)
	service := NewService(repository, adapter, partnerSvc, adCampaignSvc, subscriptionSvc, auditSvc, emailSvc)
	handler := NewHandler(service)

	return &Module{Repository: repository, Service: service, Handler: handler}
}
