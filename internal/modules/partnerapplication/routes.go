package partnerapplication

import (
	"github.com/gin-gonic/gin"

	"pleco-api/internal/middleware"
	"pleco-api/internal/services"
)

func SetupRoutes(api *gin.RouterGroup, handler *Handler, jwtService *services.JWTService, permSvc middleware.PermissionChecker) {
	self := api.Group("/partner-applications")
	self.Use(middleware.AuthMiddleware(jwtService))
	self.POST("/apply", handler.Apply)
	self.GET("/me", handler.GetMine)

	admin := api.Group("/admin/partner-applications")
	admin.Use(middleware.AuthMiddleware(jwtService))
	admin.GET("/pending", middleware.RequirePermission(permSvc, "partner_application.review"), handler.AdminGetPending)
	admin.POST("/:id/approve", middleware.RequirePermission(permSvc, "partner_application.review"), handler.AdminApprove)
	admin.POST("/:id/reject", middleware.RequirePermission(permSvc, "partner_application.review"), handler.AdminReject)
}
