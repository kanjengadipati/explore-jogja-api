package commission

import (
	"pleco-api/internal/middleware"
	"pleco-api/internal/modules/permission"
	"pleco-api/internal/services"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(api *gin.RouterGroup, handler *Handler, jwtService *services.JWTService, permissionService *permission.Service, tokenVersionSrc middleware.AccessTokenVersionSource) {
	sales := api.Group("/sales/me")
	sales.Use(middleware.AuthMiddleware(jwtService))
	sales.Use(middleware.RequireAccessTokenVersion(tokenVersionSrc))
	sales.GET("/commissions", middleware.RequirePermission(permissionService, "commission:read_own"), handler.ListMyCommissions)

	admin := api.Group("/admin")
	admin.Use(middleware.AuthMiddleware(jwtService))
	admin.Use(middleware.RequireAccessTokenVersion(tokenVersionSrc))
	admin.GET("/sales-performance", middleware.RequirePermission(permissionService, "commission:read_all"), handler.GetSalesPerformanceReport)
	admin.GET("/sales-commission-rate", middleware.RequirePermission(permissionService, "commission:read_all"), handler.GetCommissionRate)
	admin.PUT("/sales-commission-rate", middleware.RequirePermission(permissionService, "commission:manage_rate"), handler.UpdateCommissionRate)
	admin.PUT("/commissions/:id/status", middleware.RequirePermission(permissionService, "commission:manage_payout"), handler.UpdateCommissionStatus)
}
