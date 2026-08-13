package event

import (
	"pleco-api/internal/middleware"
	"pleco-api/internal/services"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(api *gin.RouterGroup, handler *Handler, jwtService *services.JWTService, permSvc middleware.PermissionChecker) {
	events := api.Group("/events")
	events.GET("", handler.GetAll)
	events.GET("/search", handler.Search)
	events.GET("/:id", handler.GetByID)

	protected := events.Group("")
	protected.Use(middleware.AuthMiddleware(jwtService), middleware.RequirePermission(permSvc, "event.manage"))
	protected.POST("", handler.Create)
	// Dashboard compatibility: the admin UI posts create payloads to
	// PUT /events/create. Route it to the create handler (must be declared
	// before the /:id update route).
	protected.PUT("/create", handler.Create)
	protected.PUT("/:id", handler.Update)
	protected.DELETE("/:id", handler.Delete)
}
