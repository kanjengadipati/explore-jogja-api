package user

import (
	"pleco-api/internal/middleware"
	"pleco-api/internal/modules/permission"
	"pleco-api/internal/services"

	"github.com/gin-gonic/gin"
)

// SetupReferralRoutes registers the sales referral-code endpoints. Call this
// alongside SetupRoutes in appsetup/router.go.
func SetupReferralRoutes(api *gin.RouterGroup, handler *Handler, jwtService *services.JWTService, permissionService *permission.Service, tokenVersionSrc middleware.AccessTokenVersionSource) {
	protected := api.Group("/auth")
	protected.Use(middleware.AuthMiddleware(jwtService))
	protected.Use(middleware.RequireAccessTokenVersion(tokenVersionSrc))

	protected.GET("/referral-code", middleware.RequirePermission(permissionService, "sales:manage-referral"), handler.GetMyReferralCode)
	protected.POST("/referral-code/regenerate", middleware.RequirePermission(permissionService, "sales:manage-referral"), handler.RegenerateMyReferralCode)
}
