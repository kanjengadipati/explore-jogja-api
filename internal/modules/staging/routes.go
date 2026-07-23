package staging

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"pleco-api/internal/ai"
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

func (m *Module) RegisterRoutes(rg *gin.RouterGroup) {
	staging := rg.Group("/admin/staging")

	staging.GET("/destinations", m.Handler.GetPendingDestinations)
	staging.POST("/destinations/ai-review", m.Handler.AIReviewDestinations)
	staging.POST("/destinations/approve", m.Handler.ApproveDestinations)
	staging.POST("/destinations/reject", m.Handler.RejectDestinations)
}
