package destination

import (
	"pleco-api/internal/middleware"
	"pleco-api/internal/services"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(api *gin.RouterGroup, handler *Handler, jwtService *services.JWTService) {
	dest := api.Group("/destinations")
	dest.GET("", handler.GetAll)
	dest.GET("/search", handler.Search)
	dest.GET("/:id", handler.GetByID)
	dest.GET("/category/:category", handler.GetByCategory)

	// Protected routes (require admin auth)
	protected := dest.Group("")
	protected.Use(middleware.AuthMiddleware(jwtService))
	protected.PUT("/:id", handler.Update)
}
