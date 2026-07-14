package rental

import (
	"pleco-api/internal/middleware"
	"pleco-api/internal/services"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(api *gin.RouterGroup, handler *Handler, jwtService *services.JWTService) {
	rentals := api.Group("/rentals")
	rentals.GET("", handler.GetAll)
	rentals.GET("/search", handler.Search)
	rentals.GET("/:id", handler.GetByID)

	protected := rentals.Group("")
	protected.Use(middleware.AuthMiddleware(jwtService))
	protected.POST("", handler.Create)
	protected.PUT("/:id", handler.Update)
	protected.DELETE("/:id", handler.Delete)
}
