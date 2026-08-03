package business

import (
	"pleco-api/internal/middleware"
	"pleco-api/internal/services"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(api *gin.RouterGroup, handler *Handler, jwtService *services.JWTService, permSvc middleware.PermissionChecker, tokenVersionSrc middleware.AccessTokenVersionSource) {
	businesses := api.Group("/businesses")
	businesses.Use(middleware.AuthMiddleware(jwtService))
	businesses.POST("", middleware.RequirePermission(permSvc, "business.create_own"), handler.CreateMyBusiness)

	// Self-service business dashboard
	self := businesses.Group("/me")
	self.POST("", middleware.RequirePermission(permSvc, "business.create_own"), handler.CreateMyBusiness)
	self.GET("", middleware.RequirePermission(permSvc, "business.read_own"), handler.GetMyBusinesses)
	self.GET("/:id", middleware.RequirePermission(permSvc, "business.read_own"), handler.GetMyBusiness)
	self.PUT("/:id", middleware.RequirePermission(permSvc, "business.update_own"), handler.UpdateMyBusiness)

	// Owned listings (claimed via listing claims)
	self.GET("/:id/listings", middleware.RequirePermission(permSvc, "business.read_own"), handler.GetMyListings)

	// Business promotions (require partner role for management)
	self.GET("/:id/promotions", middleware.RequirePermission(permSvc, "business.read_own"), handler.ListMyPromotions)
	self.POST("/:id/promotions", middleware.RequirePermission(permSvc, "promotion.manage_own"), handler.CreateMyPromotion)
	self.PUT("/:id/promotions/:pid", middleware.RequirePermission(permSvc, "promotion.manage_own"), handler.UpdateMyPromotion)
	self.DELETE("/:id/promotions/:pid", middleware.RequirePermission(permSvc, "promotion.manage_own"), handler.DeleteMyPromotion)

	// Business reviews
	self.GET("/:id/reviews", middleware.RequirePermission(permSvc, "business.read_own"), handler.ListMyReviews)
	self.POST("/:id/reviews/:rid/reply", middleware.RequirePermission(permSvc, "review.reply_own"), handler.ReplyToReview)

	// Business subscription (self-service)
	self.GET("/:id/subscription", middleware.RequirePermission(permSvc, "business.read_own"), handler.GetMySubscription)

	// Admin approval workflow — mirrors /auth/admin/partners
	authAdmin := api.Group("/auth")
	authAdmin.Use(middleware.AuthMiddleware(jwtService))
	authAdmin.Use(middleware.RequireAccessTokenVersion(tokenVersionSrc))

	adminBiz := authAdmin.Group("/admin/businesses")
	adminBiz.GET("", middleware.RequirePermission(permSvc, "business.read_all"), handler.AdminGetAll)
	adminBiz.GET("/pending", middleware.RequirePermission(permSvc, "business.approve"), handler.AdminGetPending)
	adminBiz.GET("/:id", middleware.RequirePermission(permSvc, "business.read_all"), handler.AdminGetByID)
	adminBiz.POST("/:id/approve", middleware.RequirePermission(permSvc, "business.approve"), handler.AdminApprove)
	adminBiz.POST("/:id/reject", middleware.RequirePermission(permSvc, "business.reject"), handler.AdminReject)
	adminBiz.POST("/:id/suspend", middleware.RequirePermission(permSvc, "business.suspend"), handler.AdminSuspend)
}
