package promotion

import (
	"pleco-api/internal/middleware"
	"pleco-api/internal/services"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(api *gin.RouterGroup, handler *Handler, jwtService *services.JWTService) {
	promotions := api.Group("/promotions")
	promotions.GET("", handler.GetAll)
	promotions.GET("/search", handler.Search)
	promotions.GET("/:id", handler.GetByID)

	protected := promotions.Group("")
	protected.Use(middleware.AuthMiddleware(jwtService))
	protected.POST("", handler.Create)
	protected.PUT("/:id", handler.Update)
	protected.DELETE("/:id", handler.Delete)
}
