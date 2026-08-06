package destination

import (
	"pleco-api/internal/middleware"
	"pleco-api/internal/services"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(api *gin.RouterGroup, handler *Handler, jwtService *services.JWTService, permSvc middleware.PermissionChecker) {
	dest := api.Group("/destinations")
	dest.GET("", handler.GetAll)
	dest.GET("/search", handler.Search)
	dest.GET("/:id", handler.GetByID)
	dest.GET("/category/:category", handler.GetByCategory)

	// Admin-write routes (require auth + destination.manage permission)
	protected := dest.Group("")
	protected.Use(middleware.AuthMiddleware(jwtService), middleware.RequirePermission(permSvc, "destination.manage"))
	protected.POST("", handler.Create)
	protected.PUT("/:id", handler.Update)
	protected.DELETE("/:id", handler.Delete)

	// User-self routes (require auth only)
	self := dest.Group("")
	self.Use(middleware.AuthMiddleware(jwtService))
	self.GET("/my-status", handler.GetUserDestinations)
	self.POST("/my-status/:slug", handler.UpdateUserDestinationStatus)
}

// SetupAdminContentRoutes wires the content-gen queue under /admin/content-queue.
// Mounted separately so it can carry stricter permission middleware at the call site.
func SetupAdminContentRoutes(api *gin.RouterGroup, contentHandler *ContentGenHandler, jwtService *services.JWTService, permSvc middleware.PermissionChecker, tokenVersionSrc middleware.AccessTokenVersionSource) {
	admin := api.Group("/admin/content-queue")
	admin.Use(middleware.AuthMiddleware(jwtService))
	admin.Use(middleware.RequireAccessTokenVersion(tokenVersionSrc))
	admin.GET("", middleware.RequirePermission(permSvc, "destination.read_all"), contentHandler.ListQueue)
	admin.POST("/:id/generate", middleware.RequirePermission(permSvc, "destination.manage"), contentHandler.GenerateDraft)
	admin.POST("/:id/approve", middleware.RequirePermission(permSvc, "destination.manage"), contentHandler.ApproveDraft)
	admin.POST("/:id/reject", middleware.RequirePermission(permSvc, "destination.manage"), contentHandler.RejectDraft)
	admin.POST("/:id/regenerate", middleware.RequirePermission(permSvc, "destination.manage"), contentHandler.RegenerateDraft)
}
