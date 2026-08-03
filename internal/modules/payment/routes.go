package payment

import (
	"pleco-api/internal/middleware"
	"pleco-api/internal/services"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(api *gin.RouterGroup, webhookGroup *gin.RouterGroup, handler *Handler, jwtService *services.JWTService, permSvc middleware.PermissionChecker, tokenVersionSrc middleware.AccessTokenVersionSource) {
	admin := api.Group("/admin/payments")
	admin.Use(middleware.AuthMiddleware(jwtService))
	admin.Use(middleware.RequireAccessTokenVersion(tokenVersionSrc))
	admin.Use(middleware.RequirePermission(permSvc, "payment.manage"))
	admin.POST("/create", handler.CreateTransaction)
	admin.GET("", handler.ListTransactions)

	// Self-service subscription upgrade for business owners.
	self := api.Group("/businesses")
	self.Use(middleware.AuthMiddleware(jwtService))
	self.Use(middleware.RequireAccessTokenVersion(tokenVersionSrc))
	self.POST("/me/:id/subscription/upgrade", middleware.RequirePermission(permSvc, "business.update_own"), handler.CreateSubscriptionUpgrade)

	// Webhook: NO AuthMiddleware — Midtrans doesn't send our JWT.
	// Security is enforced solely via HMAC-SHA512 signature verification in HandleNotification.
	webhookGroup.POST("/midtrans/notification", handler.HandleNotification)
}
