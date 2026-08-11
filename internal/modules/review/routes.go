package review

import (
	"pleco-api/internal/middleware"
	"pleco-api/internal/modules/permission"
	"pleco-api/internal/services"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(api *gin.RouterGroup, handler *Handler, jwtService *services.JWTService, permissionService *permission.Service) {
	reviews := api.Group("/reviews")
	reviews.GET("", handler.GetAll)
	reviews.GET("/search", handler.Search)
	reviews.GET("/:id", handler.GetByID)

	protected := reviews.Group("")
	protected.Use(middleware.AuthMiddleware(jwtService))
	protected.POST("", handler.Create)
	protected.PUT("/:id", handler.Update)
	protected.DELETE("/:id", handler.Delete)

	// Admin moderation surface — sees all statuses, unlike the public list.
	admin := reviews.Group("/admin")
	admin.Use(middleware.AuthMiddleware(jwtService))
	admin.Use(middleware.RequirePermission(permissionService, "review.manage"))
	admin.GET("", handler.GetAllAdmin)
}
