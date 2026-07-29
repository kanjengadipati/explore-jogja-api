package partner

import (
	"time"

	"pleco-api/internal/middleware"
	"pleco-api/internal/services"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(api *gin.RouterGroup, handler *Handler, jwtService *services.JWTService, permSvc middleware.PermissionChecker, tokenVersionSrc middleware.AccessTokenVersionSource, rateStore middleware.RateLimitStore) {
	partners := api.Group("/partners")

	// Public routes
	partners.GET("", handler.GetAll)
	partners.GET("/search", handler.Search)
	partners.GET("/sponsored", handler.GetSponsored)
	partners.GET("/:id", handler.GetByID)

	if rateStore == nil {
		rateStore = middleware.NewInMemoryRateLimitStore()
	}
	trackLimiter := middleware.NewRateLimiterWithStore(30, time.Minute, rateStore)
	partners.POST("/:id/track/impression", trackLimiter.Middleware(), handler.TrackImpression)
	partners.POST("/:id/track/click", trackLimiter.Middleware(), handler.TrackClick)

	// Self-service partner (auth required, no role gate — anyone logged in can apply)
	self := partners.Group("")
	self.Use(middleware.AuthMiddleware(jwtService))
	self.GET("/me", handler.GetMyListings)
	self.GET("/me/:id", middleware.RequirePermission(permSvc, "partner.read_own"), handler.GetMyListing)
	self.PUT("/me/:id", middleware.RequirePermission(permSvc, "partner.update_own"), handler.UpdateMyListing)
	self.DELETE("/me/:id", middleware.RequirePermission(permSvc, "partner.delete_own"), handler.DeleteMyListing)
	self.POST("/me/:id/submit-for-review", middleware.RequirePermission(permSvc, "partner.update_own"), handler.SubmitForReview)

	// Partner promotions (require partner role)
	self.GET("/me/:id/promotions", middleware.RequirePermission(permSvc, "partner.read_own"), handler.ListMyPromotions)
	self.POST("/me/:id/promotions", middleware.RequirePermission(permSvc, "promotion.manage_own"), handler.CreateMyPromotion)
	self.PUT("/me/:id/promotions/:pid", middleware.RequirePermission(permSvc, "promotion.manage_own"), handler.UpdateMyPromotion)
	self.DELETE("/me/:id/promotions/:pid", middleware.RequirePermission(permSvc, "promotion.manage_own"), handler.DeleteMyPromotion)

	// Partner reviews (require partner role)
	self.GET("/me/:id/reviews", middleware.RequirePermission(permSvc, "partner.read_own"), handler.ListMyReviews)
	self.POST("/me/:id/reviews/:rid/reply", middleware.RequirePermission(permSvc, "review.reply_own"), handler.ReplyToReview)

	// Admin manage-all — existing admin panel uses POST/PUT/DELETE /partners
	admin := partners.Group("")
	admin.Use(middleware.AuthMiddleware(jwtService))
	admin.Use(middleware.RequirePermission(permSvc, "partner.manage_all"))
	admin.POST("", handler.AdminCreate)
	admin.PUT("/:id", handler.AdminUpdate)
	admin.DELETE("/:id", handler.AdminDelete)

	// Admin approval workflow — follows /auth/admin/... convention (user, audit, role)
	authAdmin := api.Group("/auth")
	authAdmin.Use(middleware.AuthMiddleware(jwtService))
	authAdmin.Use(middleware.RequireAccessTokenVersion(tokenVersionSrc))

	adminPartners := authAdmin.Group("/admin/partners")
	adminPartners.GET("", middleware.RequirePermission(permSvc, "partner.read_all"), handler.AdminGetAll)
	adminPartners.GET("/pending", middleware.RequirePermission(permSvc, "partner.approve"), handler.AdminGetPending)
	adminPartners.GET("/:id", middleware.RequirePermission(permSvc, "partner.read_all"), handler.AdminGetByID)
	adminPartners.POST("/:id/approve", middleware.RequirePermission(permSvc, "partner.approve"), handler.AdminApprove)
	adminPartners.POST("/:id/reject", middleware.RequirePermission(permSvc, "partner.reject"), handler.AdminReject)
	adminPartners.POST("/:id/suspend", middleware.RequirePermission(permSvc, "partner.suspend"), handler.AdminSuspend)
}
