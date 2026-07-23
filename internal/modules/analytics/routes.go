package analytics

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Module struct {
	Handler *Handler
}

func BuildModule(db *gorm.DB) *Module {
	handler := NewHandler(db)
	return &Module{Handler: handler}
}

func (m *Module) RegisterAdminRoutes(rg *gin.RouterGroup) {
	admin := rg.Group("/admin/analytics")
	admin.GET("/overview", m.Handler.GetOverview)
	admin.GET("/top-destinations", m.Handler.GetTopDestinations)
	admin.GET("/categories", m.Handler.GetCategoryStats)
	admin.GET("/sub-regions", m.Handler.GetSubRegionStats)
	admin.GET("/recent-activity", m.Handler.GetRecentActivity)
	admin.GET("/reports", m.Handler.GetReportSummary)
}
