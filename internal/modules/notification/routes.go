package notification

import (
	"pleco-api/internal/middleware"
	"pleco-api/internal/services"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(api *gin.RouterGroup, handler *Handler, jwtService *services.JWTService) {
	notif := api.Group("/notifications")
	notif.Use(middleware.AuthMiddleware(jwtService))
	notif.GET("", handler.GetNotifications)
	notif.PUT("/:id/read", handler.MarkRead)
}
