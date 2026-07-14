package souvenir

import (
	"pleco-api/internal/middleware"
	"pleco-api/internal/services"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(api *gin.RouterGroup, handler *Handler, jwtService *services.JWTService) {
	souvenirs := api.Group("/souvenirs")
	souvenirs.GET("", handler.GetAll)
	souvenirs.GET("/search", handler.Search)
	souvenirs.GET("/:id", handler.GetByID)

	protected := souvenirs.Group("")
	protected.Use(middleware.AuthMiddleware(jwtService))
	protected.POST("", handler.Create)
	protected.PUT("/:id", handler.Update)
	protected.DELETE("/:id", handler.Delete)
}
