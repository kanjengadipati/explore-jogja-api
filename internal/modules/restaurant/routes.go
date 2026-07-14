package restaurant

import (
	"pleco-api/internal/middleware"
	"pleco-api/internal/services"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(api *gin.RouterGroup, handler *Handler, jwtService *services.JWTService) {
	restaurants := api.Group("/restaurants")
	restaurants.GET("", handler.GetAll)
	restaurants.GET("/search", handler.Search)
	restaurants.GET("/:id", handler.GetByID)

	protected := restaurants.Group("")
	protected.Use(middleware.AuthMiddleware(jwtService))
	protected.POST("", handler.Create)
	protected.PUT("/:id", handler.Update)
	protected.DELETE("/:id", handler.Delete)
}
