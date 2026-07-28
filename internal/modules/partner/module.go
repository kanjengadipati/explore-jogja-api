package partner

import (
	"pleco-api/internal/middleware"
	"pleco-api/internal/modules/promotion"
	"pleco-api/internal/modules/review"
	"pleco-api/internal/modules/user"

	"gorm.io/gorm"
)

type Module struct {
	Repository Repository
	Service    *Service
	Handler    *Handler
}

func BuildModule(db *gorm.DB, permSvc middleware.PermissionChecker, promoSvc *promotion.Service, reviewSvc *review.Service, userSvc *user.Service) *Module {
	repository := NewRepository(db)
	service := NewService(repository, userSvc)
	handler := NewHandler(service, promoSvc, reviewSvc)

	handler.PermissionSvc = permSvc

	return &Module{
		Repository: repository,
		Service:    service,
		Handler:    handler,
	}
}
