package payment

import (
	"gorm.io/gorm"

	"pleco-api/internal/modules/adcampaign"
	"pleco-api/internal/modules/audit"
	"pleco-api/internal/modules/business"
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
	adCampaignSvc *adcampaign.Service,
	subscriptionSvc *subscription.Service,
	auditSvc *audit.Service,
	emailSvc services.EmailService,
	bizRepo business.Repository,
) *Module {
	repository := NewRepository(db)
	adapter := midtransprovider.NewAdapter(midtransClient)
	service := NewService(repository, adapter, adCampaignSvc, subscriptionSvc, auditSvc, emailSvc, bizRepo)
	handler := NewHandler(service)

	return &Module{Repository: repository, Service: service, Handler: handler}
}
