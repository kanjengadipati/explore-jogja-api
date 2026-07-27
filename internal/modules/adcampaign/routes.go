package adcampaign

import (
	"pleco-api/internal/middleware"
	"pleco-api/internal/services"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(api *gin.RouterGroup, handler *Handler, jwtService *services.JWTService) {
	ads := api.Group("/ads")
	ads.GET("/banners", handler.GetBanner)
	ads.POST("/campaigns/:id/track/impression", handler.TrackImpression)
	ads.POST("/campaigns/:id/track/click", handler.TrackClick)

	protected := ads.Group("")
	protected.Use(middleware.AuthMiddleware(jwtService))
	protected.GET("/campaigns", handler.GetAll)
	protected.GET("/campaigns/:id", handler.GetByID)
	protected.POST("/campaigns", handler.Create)
	protected.PUT("/campaigns/:id", handler.Update)
	protected.DELETE("/campaigns/:id", handler.Delete)
}
