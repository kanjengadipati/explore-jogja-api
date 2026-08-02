package partner

import (
	"pleco-api/internal/middleware"
	"pleco-api/internal/modules/audit"
	"pleco-api/internal/modules/business"
	"pleco-api/internal/modules/notification"
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

func BuildModule(db *gorm.DB, permSvc middleware.PermissionChecker, promoSvc *promotion.Service, reviewSvc *review.Service, userSvc *user.Service, auditSvc *audit.Service, notifSvc *notification.Service) *Module {
	repository := NewRepository(db)
	service := NewService(repository, userSvc, notifSvc)
	handler := NewHandler(service, promoSvc, reviewSvc, auditSvc, notifSvc)

	handler.PermissionSvc = permSvc

	// Phase 1 dual-write: every partner create/update/delete mirrors into the
	// businesses tables. Temporary scaffolding — retired in Phase 6.
	if gr, ok := repository.(*GormRepository); ok {
		bizSvc := business.NewService(business.NewRepository(db))
		gr.SetWriteHook(func(p *Partner, deleted bool) {
			if deleted {
				_ = bizSvc.DeleteForPartner(p.ExternalID)
				return
			}
			_ = bizSvc.SyncFromPartner(business.PartnerMirror{
				ExternalID:      p.ExternalID,
				Name:            p.Name,
				Description:     p.Description,
				Category:        p.Category,
				Phone:           p.Phone,
				Website:         p.Website,
				Status:          p.Status,
				RejectionReason: p.RejectionReason,
				SubmittedAt:     p.SubmittedAt,
				ReviewedAt:      p.ReviewedAt,
				ReviewedBy:      p.ReviewedBy,
				OwnerUserID:     p.OwnerUserID,
			})
		})
	}

	return &Module{
		Repository: repository,
		Service:    service,
		Handler:    handler,
	}
}
