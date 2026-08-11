package staging

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"pleco-api/internal/ai"
	"pleco-api/internal/middleware"
	"pleco-api/internal/modules/permission"
	"pleco-api/internal/services"
)

type Module struct {
	Handler *Handler
}

func BuildModule(db *gorm.DB, aiService *ai.Service) *Module {
	repo := NewRepository(db)
	service := NewService(repo, aiService)
	handler := NewHandler(service)
	return &Module{Handler: handler}
}

func (m *Module) RegisterRoutes(rg *gin.RouterGroup, jwtService *services.JWTService, permissionService *permission.Service) {
	staging := rg.Group("/admin/staging")
	staging.Use(middleware.AuthMiddleware(jwtService))
	staging.Use(middleware.RequirePermission(permissionService, "staging.review"))

	staging.GET("/destinations", m.Handler.GetPendingDestinations)
	staging.POST("/destinations/ai-review", m.Handler.AIReviewDestinations)
	staging.POST("/destinations/approve", m.Handler.ApproveDestinations)
	staging.POST("/destinations/reject", m.Handler.RejectDestinations)

	staging.GET("/events", m.Handler.GetPendingEvents)
	staging.POST("/events/ai-review", m.Handler.AIReviewEvents)
	staging.POST("/events/approve", m.Handler.ApproveEvents)
	staging.POST("/events/reject", m.Handler.RejectEvents)
}
