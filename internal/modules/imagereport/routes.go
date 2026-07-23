package imagereport

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Module struct {
	Handler *Handler
}

func BuildModule(db *gorm.DB) *Module {
	repo := NewRepository(db)
	service := NewService(repo)
	handler := NewHandler(service)
	return &Module{Handler: handler}
}

// RegisterAdminRoutes registers admin-only routes under /admin/image-reports
func (m *Module) RegisterAdminRoutes(rg *gin.RouterGroup) {
	reports := rg.Group("/admin/image-reports")
	reports.GET("", m.Handler.GetAll)
	reports.GET("/stats", m.Handler.GetStats)
	reports.POST("/:id/resolve", m.Handler.Resolve)
	reports.POST("/:id/dismiss", m.Handler.Dismiss)
}

// RegisterPublicRoutes registers the public report submission endpoint
func (m *Module) RegisterPublicRoutes(dest *gin.RouterGroup) {
	dest.POST("/:id/report", m.Handler.CreateReport)
}
