package adcampaign

import (
	"pleco-api/internal/middleware"
	"pleco-api/internal/services"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(api *gin.RouterGroup, handler *Handler, jwtService *services.JWTService, permSvc middleware.PermissionChecker) {
	ads := api.Group("/ads")
	ads.GET("/banners", handler.GetBanner)
	ads.GET("/house", handler.GetHouseAd)
	ads.POST("/campaigns/:id/track/impression", handler.TrackImpression)
	ads.POST("/campaigns/:id/track/click", handler.TrackClick)

	protected := ads.Group("")
	protected.Use(middleware.AuthMiddleware(jwtService))
	protected.Use(middleware.RequirePermission(permSvc, "ads.manage"))
	protected.GET("/campaigns", handler.GetAll)
	protected.GET("/campaigns/:id", handler.GetByID)
	protected.POST("/campaigns", handler.Create)
	protected.PUT("/campaigns/:id", handler.Update)
	protected.DELETE("/campaigns/:id", handler.Delete)
	protected.GET("/house-ads", handler.GetAllHouseAds)
	protected.POST("/house-ads", handler.CreateHouseAd)
	protected.PUT("/house-ads/:id", handler.UpdateHouseAd)
	protected.DELETE("/house-ads/:id", handler.DeleteHouseAd)
}
