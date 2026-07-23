package article

import (
	"pleco-api/internal/middleware"
	"pleco-api/internal/services"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(api *gin.RouterGroup, handler *Handler, jwtService *services.JWTService) {
	articles := api.Group("/articles")

	// Public routes
	articles.GET("", handler.GetAll)
	articles.GET("/search", handler.Search)
	articles.GET("/category/:category", handler.GetByCategory)
	articles.GET("/slug/:slug", handler.GetBySlug)
	articles.GET("/:id", handler.GetByID)

	// Protected routes (admin)
	protected := articles.Group("")
	protected.Use(middleware.AuthMiddleware(jwtService))
	protected.POST("", handler.Create)
	protected.PUT("/:id", handler.Update)
	protected.DELETE("/:id", handler.Delete)
}
