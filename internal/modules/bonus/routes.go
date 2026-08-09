package bonus

import (
	"pleco-api/internal/middleware"
	"pleco-api/internal/modules/permission"
	"pleco-api/internal/services"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(api *gin.RouterGroup, handler *Handler, jwtService *services.JWTService, permissionService *permission.Service, tokenVersionSrc middleware.AccessTokenVersionSource) {
	me := api.Group("/bonuses/me")
	me.Use(middleware.AuthMiddleware(jwtService))
	me.Use(middleware.RequireAccessTokenVersion(tokenVersionSrc))
	me.GET("", middleware.RequirePermission(permissionService, "bonus:read_own"), handler.ListMyBonuses)

	bonuses := api.Group("/bonuses")
	bonuses.Use(middleware.AuthMiddleware(jwtService))
	bonuses.Use(middleware.RequireAccessTokenVersion(tokenVersionSrc))
	bonuses.GET("", middleware.RequirePermission(permissionService, "bonus:read_all"), handler.ListAllBonuses)

	adminBonuses := api.Group("/admin/bonuses")
	adminBonuses.Use(middleware.AuthMiddleware(jwtService))
	adminBonuses.Use(middleware.RequireAccessTokenVersion(tokenVersionSrc))
	adminBonuses.PUT("/:id/status", middleware.RequirePermission(permissionService, "bonus:manage_payout"), handler.UpdateBonusStatus)

	admin := api.Group("/admin/bonus-rules")
	admin.Use(middleware.AuthMiddleware(jwtService))
	admin.Use(middleware.RequireAccessTokenVersion(tokenVersionSrc))
	admin.GET("", middleware.RequirePermission(permissionService, "bonus:read_all"), handler.ListBonusRules)
	admin.POST("", middleware.RequirePermission(permissionService, "bonus:manage_rules"), handler.CreateBonusRule)
	admin.PUT("/:id", middleware.RequirePermission(permissionService, "bonus:manage_rules"), handler.UpdateBonusRule)
	admin.DELETE("/:id", middleware.RequirePermission(permissionService, "bonus:manage_rules"), handler.DeleteBonusRule)
}
