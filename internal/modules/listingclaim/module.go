package listingclaim

import (
	"pleco-api/internal/middleware"
	"pleco-api/internal/modules/business"
	"pleco-api/internal/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Module struct {
	Repository Repository
	Service    *Service
	Handler    *Handler
}

func BuildModule(db *gorm.DB, bizRepo business.Repository) *Module {
	repo := NewRepository(db)
	service := NewService(repo, bizRepo)
	handler := NewHandler(service)
	return &Module{
		Repository: repo,
		Service:    service,
		Handler:    handler,
	}
}

func SetupRoutes(api *gin.RouterGroup, handler *Handler, jwtService *services.JWTService, permSvc middleware.PermissionChecker) {
	self := api.Group("/listing-claims")
	self.Use(middleware.AuthMiddleware(jwtService))
	self.POST("", middleware.RequirePermission(permSvc, "listing_claim.submit_own"), handler.Submit)
	self.POST("/submit", middleware.RequirePermission(permSvc, "listing_claim.submit_own"), handler.Submit)
	self.GET("/me", middleware.RequirePermission(permSvc, "listing_claim.read_own"), handler.GetMine)

	admin := api.Group("/admin/listing-claims")
	admin.Use(middleware.AuthMiddleware(jwtService))
	admin.GET("/pending", middleware.RequirePermission(permSvc, "listing_claim.read_all"), handler.AdminGetPending)
	admin.POST("/:id/approve", middleware.RequirePermission(permSvc, "listing_claim.approve"), handler.AdminApprove)
	admin.POST("/:id/reject", middleware.RequirePermission(permSvc, "listing_claim.reject"), handler.AdminReject)
}
