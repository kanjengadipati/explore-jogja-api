package adcampaign

import (
	"time"

	"pleco-api/internal/middleware"
	"pleco-api/internal/services"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(api *gin.RouterGroup, handler *Handler, jwtService *services.JWTService, permSvc middleware.PermissionChecker, rateStore middleware.RateLimitStore) {
	ads := api.Group("/ads")
	ads.GET("/banners", handler.GetBanner)
	ads.GET("/house", handler.GetHouseAd)
	ads.GET("/ecosystem", handler.GetEcosystem)

	if rateStore == nil {
		rateStore = middleware.NewInMemoryRateLimitStore()
	}
	trackLimiter := middleware.NewRateLimiterWithStore(30, time.Minute, rateStore)
	ads.POST("/campaigns/:id/track/impression", trackLimiter.Middleware(), handler.TrackImpression)
	ads.POST("/campaigns/:id/track/click", trackLimiter.Middleware(), handler.TrackClick)

	protected := ads.Group("")
	protected.Use(middleware.AuthMiddleware(jwtService))
	protected.Use(middleware.RequirePermission(permSvc, "ads.manage"))
	protected.GET("/campaigns", handler.GetAll)
	protected.GET("/campaigns/:id", handler.GetByID)
	protected.POST("/campaigns", handler.Create)
	protected.PUT("/campaigns/:id", handler.Update)
	protected.DELETE("/campaigns/:id", handler.Delete)
	protected.POST("/campaigns/:id/approve", handler.Approve)
	protected.POST("/campaigns/:id/reject", handler.Reject)
	protected.GET("/house-ads", handler.GetAllHouseAds)
	protected.POST("/house-ads", handler.CreateHouseAd)
	protected.PUT("/house-ads/:id", handler.UpdateHouseAd)
	protected.DELETE("/house-ads/:id", handler.DeleteHouseAd)
}
