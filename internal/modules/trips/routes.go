package trips

import (
	"pleco-api/internal/middleware"
	"pleco-api/internal/services"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(api *gin.RouterGroup, handler *Handler, jwtService *services.JWTService) {
	// All trips routes require authentication — no public endpoints.
	group := api.Group("/trips")
	group.Use(middleware.AuthMiddleware(jwtService))

	group.GET("", handler.GetAll)
	group.GET("/:id", handler.GetByID)
	group.POST("", handler.Create)
	group.PATCH("/:id", handler.Update)
	group.DELETE("/:id", handler.Delete)
}
